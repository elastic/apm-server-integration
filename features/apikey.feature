Feature: APIKey

  Scenario: A request to the intake endpoint with a valid APIKey auth is accepted
    Given an apikey created with apm-server subcommand
    Given apm-server started with apikey enabled
    When a request is sent with matching credentials in the Authorization header
    Then apm-server returns 202 - Accepted

  Scenario: A request to the intake endpoint with an invalid APIKey auth is not accepted
    Given apm-server started with apikey enabled
    When a request is sent with non-matching credentials in the Authorization header
    Then apm-server returns 401 - Authorization denied

  Scenario: A request to the intake endpoint without an APIKey is not accepted
    Given apm-server started with apikey enabled
    When a request is sent without credentials in the Authorization header
    Then apm-server returns 401 - Authorization denied
