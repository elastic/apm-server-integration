package tests

import (
	"errors"
	"fmt"
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"gotestpoc/stack"
	//"gotestpoc/stack"
	"net/http"
	"strings"
)

func APIKeyContext(s *godog.Suite, stackVersion string, c chan string) {

	var apmServer *stack.Service
	var apiKeyCheck apiKeyChecker

	var startAPM = func() error {
		apmServer = stack.APMServer(stackVersion, "apm-server.api_key.enabled=true")
		apiKeyCheck.recorder = &recorder{url: apmServer.Endpoint()}
		return nil
	}

	var credentials string
	var genAPIKey = func() error {
		esUrl := <-c
		apmServer = stack.APMServerSubCommand(stackVersion, "credentials", []string{"apikey", "create", "--json"},
			"output.elasticsearch.hosts=["+esUrl+"]")
		// TODO parse output and overwrite credentials
		return nil
	}

	// this doesn't get triggered if there is a panic before
	s.AfterScenario(func(i interface{}, _ error) {
		if scenario, ok := i.(*gherkin.Scenario); ok {
			// never do this. not even in a POC
			if strings.Contains(scenario.Name, "A request to the intake endpoint") {
				apmServer.Stop()
			}
		}
	})

	s.Step("^an apikey created with apm-server subcommand$", genAPIKey)
	s.Step("^apm-server started with apikey enabled$", startAPM)
	s.Step("^a request is sent with matching credentials in the Authorization header$",
		func() error { return apiKeyCheck.makeRequest(credentials)})
	s.Step("^apm-server returns 202 - Accepted$",
		func() error { return apiKeyCheck.validate(http.StatusAccepted)})
	s.Step("^a request is sent with non-matching credentials in the Authorization header$",
		func() error { return apiKeyCheck.makeRequest("foobar")})
	s.Step("^a request is sent without credentials in the Authorization header$",
		func () error { return apiKeyCheck.makeRequest("")})
	s.Step("^apm-server returns 401 - Authorization denied$",
		func () error { return apiKeyCheck.validate(http.StatusUnauthorized)})
}


type apiKeyChecker struct {
	*recorder
}

func (c *apiKeyChecker) makeRequest(credentials string) error {
	h := make(http.Header)
	if credentials != "" {
		h.Set("Authorization", "ApiKey " + credentials)
	}
	return c.sendRequest(strings.NewReader(`{"metadata": {"user": {"username": "logged-in-user", "id": "axb123hg", "email": "user@mail.com"}, "labels": {"tag0": null, "tag1": "one", "tag2": 2}, "process": {"ppid": null, "pid": 1234, "argv": null, "title": null}, "system": null, "service": {"name": "1234_service-12a3", "node": {"configured_name": "node-1"},"language": {"version": null, "name":"ecmascript"}, "agent": {"version": "3.14.0", "name": "elastic-node"}, "environment": null, "framework": null,"version": null, "runtime": null}}}
{"metricset": { "samples": { "transaction.breakdown.count":{"value":12}, "transaction.duration.sum.us":{"value":12}, "transaction.duration.count":{"value":2}, "transaction.self_time.sum.us":{"value":10}, "transaction.self_time.count":{"value":2}, "span.self_time.count":{"value":1},"span.self_time.sum.us":{"value":633.288}, "byte_counter": { "value": 1 }, "short_counter": { "value": 227 }, "integer_gauge": { "value": 42767 }, "long_gauge": { "value": 3147483648 }, "float_gauge": { "value": 9.16 }, "double_gauge": { "value": 3.141592653589793 }, "dotted.float.gauge": { "value": 6.12 }, "negative.d.o.t.t.e.d": { "value": -1022 } }, "tags": { "some": "abc", "code": 200, "success": true }, "transaction":{"type":"request","name":"GET /"},"span":{"type":"db","subtype":"mysql"},"timestamp": 1496170422281000 }}
{"metricset": { "samples": { "go.memstats.heap.sys.bytes": { "value": 6.520832e+06 }}}}`), h)
}

func (c *apiKeyChecker) validate(expectedCode int) error {
	code, _, err := c.response()
	if err != nil {
		return err
	}
	if code != expectedCode {
		return errors.New(fmt.Sprintf("unexpected response code: %d", code))
	}
	return nil
}
