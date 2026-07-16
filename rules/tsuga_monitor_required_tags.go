package rules

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// tsugaMonitorTagRule describes one required tag on a tsuga_monitor resource.
// A tag is checked in one of two mutually exclusive ways once its value can
// be statically determined:
//   - AllowedValues set: the value must be one of these.
//   - RequireNonEmpty: the value must not be an empty string.
// Neither set means the tag is required but its value isn't constrained.
type tsugaMonitorTagRule struct {
	Key             string
	AllowedValues   []string
	RequireNonEmpty bool
}

// tsugaMonitorRequiredTags is the tagging convention documented in
// modules/tsuga/monitors/README.md#tagging-convention (camunda/terraform):
//
//   - "env" determines whether the alert is delivered at all
//     (modules/tsuga/notifications matches on it).
//   - "team" determines which Slack channel it's delivered to.
//   - "service" and "security" are for searching/filtering in the Tsuga UI.
//   - "managed-by" marks the resource as Terraform-managed.
var tsugaMonitorRequiredTags = []tsugaMonitorTagRule{
	{Key: "env", RequireNonEmpty: true},
	{Key: "team", RequireNonEmpty: true},
	{Key: "service", RequireNonEmpty: true},
	{Key: "security", AllowedValues: []string{"true", "false"}},
	{Key: "managed-by", AllowedValues: []string{"terraform"}},
}

// TsugaMonitorRequiredTags requires every tsuga_monitor resource to set the
// tags in tsugaMonitorRequiredTags, and - where a tag has AllowedValues - to
// set it to one of those values whenever that value can be statically
// determined.
type TsugaMonitorRequiredTags struct {
	tflint.DefaultRule
}

func NewTsugaMonitorRequiredTagsRule() *TsugaMonitorRequiredTags {
	return &TsugaMonitorRequiredTags{}
}

func (r *TsugaMonitorRequiredTags) Name() string {
	return "camunda_tsuga_monitor_required_tags"
}

func (r *TsugaMonitorRequiredTags) Enabled() bool {
	return true
}

func (r *TsugaMonitorRequiredTags) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TsugaMonitorRequiredTags) Link() string {
	return "https://github.com/camunda/tflint-ruleset-camunda-saas/blob/master/docs/rules/camunda_tsuga_monitor_required_tags.md"
}

func (r *TsugaMonitorRequiredTags) Check(runner tflint.Runner) error {
	resources, err := runner.GetResourceContent("tsuga_monitor", &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{{Name: "tags"}},
	}, nil)
	if err != nil {
		return err
	}

	requiredKeys := make([]string, len(tsugaMonitorRequiredTags))
	for i, tagRule := range tsugaMonitorRequiredTags {
		requiredKeys[i] = tagRule.Key
	}

	for _, resource := range resources.Blocks {
		address := fmt.Sprintf("%s.%s", resource.Labels[0], resource.Labels[1])

		attr, exists := resource.Body.Attributes["tags"]
		if !exists {
			if err := r.emitIssue(runner, fmt.Sprintf(
				"'%s' must set a 'tags' attribute with the required tags: %s",
				address, strings.Join(requiredKeys, ", "),
			), resource.DefRange); err != nil {
				return err
			}

			continue
		}

		tuple, ok := hcl.UnwrapExpression(attr.Expr).(*hclsyntax.TupleConsExpr)
		if !ok {
			if err := r.emitIssue(runner, fmt.Sprintf(
				"'%s' tags must be a static list of { key = ..., value = ... } objects so the required tags can be verified",
				address,
			), attr.Range); err != nil {
				return err
			}

			continue
		}

		for _, tagRule := range tsugaMonitorRequiredTags {
			valueExpr, found := tsugaTagValueExpr(tuple, tagRule.Key)
			if !found {
				if err := r.emitIssue(runner, fmt.Sprintf(
					"'%s' is missing the required '%s' tag", address, tagRule.Key,
				), attr.Range); err != nil {
					return err
				}

				continue
			}

			if len(tagRule.AllowedValues) == 0 && !tagRule.RequireNonEmpty {
				continue
			}

			value, ok, err := tsugaEvaluateTagValue(runner, valueExpr)
			if err != nil {
				return err
			}
			if !ok {
				// The value can't be statically determined in this lint
				// context (e.g. linting modules/tsuga/monitors on its own,
				// where "env"'s value is a variable with no known value).
				// Skip rather than fail - it may well be valid once the
				// calling workspace's value is known.
				continue
			}

			if len(tagRule.AllowedValues) > 0 {
				if !slices.Contains(tagRule.AllowedValues, value) {
					if err := r.emitIssue(runner, fmt.Sprintf(
						"'%s' tag '%s' is %q, must be one of: %s",
						address, tagRule.Key, value, strings.Join(tagRule.AllowedValues, ", "),
					), valueExpr.Range()); err != nil {
						return err
					}
				}
				continue
			}

			if tagRule.RequireNonEmpty && value == "" {
				if err := r.emitIssue(runner, fmt.Sprintf(
					"'%s' tag '%s' must not be an empty string", address, tagRule.Key,
				), valueExpr.Range()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *TsugaMonitorRequiredTags) emitIssue(runner tflint.Runner, message string, issueRange hcl.Range) error {
	return runner.EmitIssue(r, message, issueRange)
}

// tsugaTagValueExpr looks for a `{ key = "<wantedKey>", value = ... }` entry
// in tuple and, if found, returns the expression assigned to that entry's
// "value" field. It only ever evaluates the "key" field of each entry (always
// a plain string literal in this codebase's convention) to find the match -
// it never evaluates any "value" field itself, so it can't fail due to
// unresolved references (e.g. "team"'s value being a variable).
func tsugaTagValueExpr(tuple *hclsyntax.TupleConsExpr, wantedKey string) (hcl.Expression, bool) {
	for _, itemExpr := range tuple.Exprs {
		obj, ok := itemExpr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			continue
		}

		var keyMatches bool
		var valueExpr hcl.Expression

		for _, item := range obj.Items {
			switch hcl.ExprAsKeyword(item.KeyExpr) {
			case "key":
				val, diags := item.ValueExpr.Value(nil)
				keyMatches = !diags.HasErrors() && val.Type() == cty.String && val.AsString() == wantedKey
			case "value":
				valueExpr = item.ValueExpr
			}
		}

		if keyMatches && valueExpr != nil {
			return valueExpr, true
		}
	}

	return nil, false
}

// tsugaEvaluateTagValue evaluates a tag's "value" expression through the
// runner (so it resolves variables using the actual value passed by the
// calling workspace, when tflint is run with --call-module-type=all). ok is
// false when the value can't be statically determined in the current lint
// context (unknown, null, sensitive, or an each.*/count.* reference) - that's
// not an error, just something this rule can't check right now.
func tsugaEvaluateTagValue(runner tflint.Runner, expr hcl.Expression) (value string, ok bool, err error) {
	// each.key/each.value/count.index are only resolved against a specific
	// resource instance by the host when it recognizes the reference at the
	// attribute's own top level. This expr was instead pulled out of a
	// nested tags-list entry by tsugaTagValueExpr, so that per-instance
	// binding isn't available here - evaluating it fails with a plain,
	// unclassified "unknown variable" error rather than one of the sentinels
	// below. Skip it the same as any other value we can't statically
	// determine, rather than let that error fail the whole rule.
	for _, traversal := range expr.Variables() {
		if root := traversal.RootName(); root == "each" || root == "count" {
			return "", false, nil
		}
	}

	err = runner.EvaluateExpr(expr, &value, nil)
	if err == nil {
		return value, true, nil
	}

	if errors.Is(err, tflint.ErrUnknownValue) || errors.Is(err, tflint.ErrNullValue) || errors.Is(err, tflint.ErrSensitive) {
		return "", false, nil
	}

	return "", false, err
}
