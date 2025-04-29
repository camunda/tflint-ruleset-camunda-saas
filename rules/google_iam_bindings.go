package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type GoogleIamAuthoritative struct {
	tflint.DefaultRule
}

func NewGoogleIamAuthoritativeTypeRule() *GoogleIamAuthoritative {
	return &GoogleIamAuthoritative{}
}

func (r *GoogleIamAuthoritative) Name() string {
	return "camunda_gcp_iam_authoritative"
}

func (r *GoogleIamAuthoritative) Enabled() bool {
	return true
}

func (r *GoogleIamAuthoritative) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *GoogleIamAuthoritative) Link() string {
	return "https://github.com/camunda/tflint-ruleset-camunda-saas/blob/master/docs/rules/camunda_gcp_iam_authoritative.md"
}

func (r *GoogleIamAuthoritative) Check(runner tflint.Runner) error {
	authoritativeIamResources := []string{"google_project_iam_policy", "google_project_iam_binding"}

	for _, authoritativeResourceType := range authoritativeIamResources {
		resources, err := runner.GetResourceContent(authoritativeResourceType, &hclext.BodySchema{
			Attributes: []hclext.AttributeSchema{},
		}, nil)

		if err != nil {
			return err
		}

		for _, resource := range resources.Blocks {
			err = runner.EnsureNoError(err, func() error {
				return runner.EmitIssue(
					r,
					fmt.Sprintf("'%s' is a dangerous resource, are you sure you want to use it?",
						resource.Labels[0]),
					resource.DefRange,
				)
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}
