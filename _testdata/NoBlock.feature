Feature: No blocks for different resources

  Scenario: Resource 1
    Given I sleep
    And I should not be blocked for "r1"
    When I acquire "r1"
    Then I sleep longer

  Scenario: Resource 2
    Given I sleep
    Given I should not be blocked for "r2"
    When I acquire "r2"
    Then I sleep longer

  Scenario: Resource 3
    Given I sleep
    Given I should not be blocked for "r3"
    When I acquire "r3"
    Then I sleep longer
