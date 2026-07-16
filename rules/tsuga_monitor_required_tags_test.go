package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TsugaMonitorRequiredTags(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "all required tags present and valid",
			Content: `
resource "tsuga_monitor" "test" {
  name = "test"

  tags = [
    { key = "env", value = "dev" },
    { key = "team", value = "sre" },
    { key = "service", value = "my-service" },
    { key = "security", value = "false" },
    { key = "managed-by", value = "terraform" },
  ]
}
`,
			Expected: helper.Issues{},
		},
		{
			// Mirrors the for_each pattern used by
			// modules/tsuga/monitors/monitor-java-oom.tf in camunda/terraform,
			// where "env"'s value comes from a variable and "team"/"service"
			// come from each.value/each.key.
			//
			// "env"'s variable has a default so helper.TestRunner can resolve
			// it. When it *can't* be resolved (e.g. linting
			// modules/tsuga/monitors on its own, with no calling workspace
			// assigning a value), real tflint reports that as
			// tflint.ErrUnknownValue and this rule skips the check - verified
			// manually against the real repository, since helper.TestRunner
			// doesn't classify that scenario the same way real tflint does
			// (it returns a plain, unwrapped conversion error instead).
			//
			// "team"/"service"'s each.value/each.key references are skipped
			// unconditionally by tsugaEvaluateTagValue's each/count check,
			// since this rule can't resolve them against a specific resource
			// instance either way - see that function's comment.
			Name: "for_each monitor with a resolvable env value",
			Content: `
variable "environment_tag" {
  type    = string
  default = "dev"
}

resource "tsuga_monitor" "java_oom" {
  for_each = local.java_service_team

  name = "${var.environment_tag}-jvm-out-of-memory-${each.key}"

  tags = [
    { key = "env", value = var.environment_tag },
    { key = "team", value = each.value },
    { key = "service", value = each.key },
    { key = "security", value = "false" },
    { key = "managed-by", value = "terraform" },
  ]
}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "env tag missing",
			Content: `
resource "tsuga_monitor" "test" {
  name = "test"

  tags = [
    { key = "team", value = "sre" },
    { key = "service", value = "my-service" },
    { key = "security", value = "false" },
    { key = "managed-by", value = "terraform" },
  ]
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTsugaMonitorRequiredTagsRule(),
					Message: "'tsuga_monitor.test' is missing the required 'env' tag",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 3},
						End:      hcl.Pos{Line: 10, Column: 4},
					},
				},
			},
		},
		{
			Name: "env, team and service tags are empty strings",
			Content: `
resource "tsuga_monitor" "test" {
  name = "test"

  tags = [
    { key = "env", value = "" },
    { key = "team", value = "" },
    { key = "service", value = "" },
    { key = "security", value = "false" },
    { key = "managed-by", value = "terraform" },
  ]
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTsugaMonitorRequiredTagsRule(),
					Message: "'tsuga_monitor.test' tag 'env' must not be an empty string",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 28},
						End:      hcl.Pos{Line: 6, Column: 30},
					},
				},
				{
					Rule:    NewTsugaMonitorRequiredTagsRule(),
					Message: "'tsuga_monitor.test' tag 'team' must not be an empty string",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 29},
						End:      hcl.Pos{Line: 7, Column: 31},
					},
				},
				{
					Rule:    NewTsugaMonitorRequiredTagsRule(),
					Message: "'tsuga_monitor.test' tag 'service' must not be an empty string",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 8, Column: 32},
						End:      hcl.Pos{Line: 8, Column: 34},
					},
				},
			},
		},
		{
			Name: "security and managed-by tags have disallowed values",
			Content: `
resource "tsuga_monitor" "test" {
  name = "test"

  tags = [
    { key = "env", value = "dev" },
    { key = "team", value = "sre" },
    { key = "service", value = "my-service" },
    { key = "security", value = "yes" },
    { key = "managed-by", value = "pulumi" },
  ]
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTsugaMonitorRequiredTagsRule(),
					Message: `'tsuga_monitor.test' tag 'security' is "yes", must be one of: true, false`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 9, Column: 33},
						End:      hcl.Pos{Line: 9, Column: 38},
					},
				},
				{
					Rule:    NewTsugaMonitorRequiredTagsRule(),
					Message: `'tsuga_monitor.test' tag 'managed-by' is "pulumi", must be one of: terraform`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 10, Column: 35},
						End:      hcl.Pos{Line: 10, Column: 43},
					},
				},
			},
		},
		{
			Name: "tags attribute missing entirely",
			Content: `
resource "tsuga_monitor" "test" {
  name = "test"
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTsugaMonitorRequiredTagsRule(),
					Message: "'tsuga_monitor.test' must set a 'tags' attribute with the required tags: env, team, service, security, managed-by",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 32},
					},
				},
			},
		},
		{
			Name: "tags set from a non-literal expression can't be verified",
			Content: `
resource "tsuga_monitor" "test" {
  name = "test"
  tags = local.common_tags
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTsugaMonitorRequiredTagsRule(),
					Message: "'tsuga_monitor.test' tags must be a static list of { key = ..., value = ... } objects so the required tags can be verified",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 27},
					},
				},
			},
		},
	}

	rule := NewTsugaMonitorRequiredTagsRule()

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
