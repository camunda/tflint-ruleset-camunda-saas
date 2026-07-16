resource "tsuga_monitor" "test" {
  name        = "test-monitor"
  owner       = "some-team-id"
  permissions = "all"
  priority    = "3"

  tags = [
    { key = "team", value = "sre" },
    { key = "managed-by", value = "terraform" },
  ]

  cluster_ids = []

  configuration = {
    log = {
      queries = [
        {
          filter = "some query"
          aggregate = {
            count = {}
          }
        }
      ]
      conditions = [{
        formula   = "q1"
        operator  = "greater_than"
        threshold = 0
      }]
      timeframe               = 1
      group_by_fields         = []
      aggregation_alert_logic = "each"
      no_data_behavior        = "resolve"
    }
  }

  message = "test"
}
