package main

import (
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
	"github.com/DATA-DOG/godog/gherkin"
	"gotestpoc/stack"
	"gotestpoc/tests"
	"os"
	"testing"
	"time"
)


func featureContext(s *godog.Suite, stackVersion string) {

	var c = make(chan string, 1)
	var elasticSearch *stack.Service
	s.BeforeFeature(func(_ *gherkin.Feature) {
		elasticSearch = stack.ElasticSearch(stackVersion)
		// kibana = Kibana(stackVersion)
		elasticSearch.Container.Networks()
		// TODO how to get the internal IP address
		c <- "172.17.0.2:9200"
	})

	s.AfterFeature(func(_ *gherkin.Feature) {
		elasticSearch.Stop()
		// kibana.Stop()
	})

	tests.HealthcheckContext(s, stackVersion)
	tests.APIKeyContext(s, stackVersion, c)
}

func TestMain(m *testing.M) {
	status := godog.RunWithOptions("godog", func(s *godog.Suite) {
		featureContext(s, "7.6.0")
	}, godog.Options{
		Format:    "pretty",
		Output: colors.Colored(os.Stdout),
		Paths:     []string{"features"},
		Randomize: time.Now().UTC().UnixNano(),
	})

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)

}
