# Either "Resource 1" or "Resource 1 again" should be blocked and so the test should fail.
Feature: Block for same resource

  Scenario: Resource 1
    And I should not be blocked for "r1"
    When I acquire "r1"
    Then I sleep longer

  Scenario: Resource 1 again
    Given I sleep
    Given I should be blocked for "r1"
    When I acquire "r1"

  Scenario: Resource 3
    Given I sleep
    Given I should not be blocked for "r3"
    When I acquire "r3"
    Then I sleep longer
