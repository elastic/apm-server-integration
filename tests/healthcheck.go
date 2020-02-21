package tests

import (
	"errors"
	"fmt"
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"gotestpoc/stack"
	"net/http"
	"strings"
)

func HealthcheckContext(s *godog.Suite, stackVersion string) {

	var apmServer *stack.Service
	var authCheck authChecker
	var noAuthCheck noAuthChecker

	var startAPM = func() error {
		apmServer = stack.APMServer(stackVersion, "apm-server.secret_token=changeme_token")
		url := apmServer.Endpoint()
		authCheck.recorder = &recorder{url: url}
		noAuthCheck.recorder = &recorder{url: url}
		return nil
	}

	// this doesn't get triggered if there is a panic before
	s.AfterScenario(func(i interface{}, _ error) {
		if scenario, ok := i.(*gherkin.Scenario); ok {
			// never do this. not even in a POC
			if strings.Contains(scenario.Name, "An authorized request to the root endpoint") {
				apmServer.Stop()
			}
		}
	})

	s.Step("^apm-server started with a secret token$", startAPM)
	s.Step("^a request is sent with a matching secret token in the Authorization header$", authCheck.makeRequest)
	s.Step("^apm-server returns 200 - OK with version, build data, and commit SHA data$", authCheck.validate)
	s.Step("^a request is sent without an Authorization header$", noAuthCheck.makeRequest)
	s.Step("^apm-server only returns 200 - OK$", noAuthCheck.validate)
}

type authChecker struct {
	*recorder
}

func (c *authChecker) makeRequest() error {
	h := make(http.Header)
	h.Set("Authorization", "Bearer changeme_token")
	return c.sendRequest(nil, h)
}

func (c *authChecker) validate() error {
	code, body, err := c.response()
	if err != nil {
		return err
	}
	if code > http.StatusOK {
		return errors.New(fmt.Sprintf("unexpected response code: %d", code))
	}
	for _, attr := range []string{"build_date", "build_sha", "version"} {
		if _, ok := body[attr]; !ok {
			return errors.New(fmt.Sprintf("expected %s, but not found in %v", attr, body))
		}
	}
	return nil
}


type noAuthChecker struct {
	*recorder
}

func (c *noAuthChecker) makeRequest() error {
	return c.sendRequest(nil, nil)
}

func (c *noAuthChecker) validate() error {
	code, body, err := c.response()
	if err != nil {
		return err
	}
	if code > http.StatusOK {
		return errors.New(fmt.Sprintf("unexpected response code: %d", code))
	}
	if len(body) > 0 {
		return errors.New(fmt.Sprintf("did not expect any body %v", body))
	}
	return nil
}