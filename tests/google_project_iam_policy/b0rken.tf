data "google_iam_policy" "test" {
  binding {
    role    = "roles/test"
    members = []
  }
}

resource "google_project_iam_policy" "test" {
  project     = "your-project-id"
  policy_data = data.google_iam_policy.test.policy_data
}
