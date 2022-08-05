package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_GoogleIamAuthoritativeType(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "issue found",
			Content: `
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

resource "google_project_iam_binding" "test" {
  project = "your-project-id"

  role    = "roles/test"
  members = []
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewGoogleIamAuthoritativeTypeRule(),
					Message: "'google_project_iam_policy' is a dangerous resource, are you sure you want to use it?",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 9, Column: 1},
						End:      hcl.Pos{Line: 9, Column: 44},
					},
				},
				{
					Rule:    NewGoogleIamAuthoritativeTypeRule(),
					Message: "'google_project_iam_binding' is a dangerous resource, are you sure you want to use it?",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 14, Column: 1},
						End:      hcl.Pos{Line: 14, Column: 45},
					},
				},
			},
		},
	}

	rule := NewGoogleIamAuthoritativeTypeRule()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"resource.tf": test.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, test.Expected, runner.Issues)
		})
	}
}
