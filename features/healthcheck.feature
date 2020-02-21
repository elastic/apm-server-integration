Feature: Healthcheck

  Scenario: An authorized request to the root endpoint returns a non-empty response body
    Given apm-server started with a secret token
    When a request is sent with a matching secret token in the Authorization header
    Then apm-server returns 200 - OK with version, build data, and commit SHA data

  Scenario: An unauthorized request to the root endpoint returns an empty response body
    Given apm-server started with a secret token
    When a request is sent without an Authorization header
    Then apm-server only returns 200 - OK

