# `camunda_tsuga_monitor_required_tags`

Require every `tsuga_monitor` resource to set the full tagging convention, and to constrain
certain tags to a fixed set of allowed values.

| Tag | Required | Allowed values |
|-----|----------|-----------------|
| `env` | Yes | non-empty string |
| `team` | Yes | non-empty string |
| `service` | Yes | non-empty string |
| `security` | Yes | `true`, `false` |
| `managed-by` | Yes | `terraform` |

A value is only checked (against its allowed list, or for non-emptiness) when it can be statically
determined. If a tag's value comes from a variable or another resource's attribute that isn't
known yet (for example, linting `modules/tsuga/monitors` on its own, without a calling workspace
assigning `environment_tag`), or from an `each.key`/`each.value`/`count.index` reference, the value
check is skipped rather than failed - only the *presence* of the tag is still enforced in that
case.

## Example

```hcl
resource "tsuga_monitor" "test" {
  name        = "test-monitor"
  owner       = "some-team-id"
  permissions = "all"
  priority    = "3"

  tags = [
    { key = "team", value = "sre" },
    { key = "managed-by", value = "terraform" },
  ]

  # ...
}
```

```
$ tflint
3 issue(s) found:

Error: 'tsuga_monitor.test' is missing the required 'env' tag (camunda_tsuga_monitor_required_tags)

  on test.tf line 7:
   7:   tags = [
   8:     { key = "team", value = "sre" },
   9:     { key = "managed-by", value = "terraform" },
  10:   ]

Error: 'tsuga_monitor.test' is missing the required 'security' tag (camunda_tsuga_monitor_required_tags)

  on test.tf line 7:
   7:   tags = [
   8:     { key = "team", value = "sre" },
   9:     { key = "managed-by", value = "terraform" },
  10:   ]

Error: 'tsuga_monitor.test' is missing the required 'service' tag (camunda_tsuga_monitor_required_tags)

  on test.tf line 7:
   7:   tags = [
   8:     { key = "team", value = "sre" },
   9:     { key = "managed-by", value = "terraform" },
  10:   ]
```

An allowed-value violation looks like this instead:

```hcl
tags = [
  { key = "security", value = "yes" },
  # ...
]
```

```
$ tflint
Error: 'tsuga_monitor.test' tag 'security' is "yes", must be one of: true, false (camunda_tsuga_monitor_required_tags)

  on test.tf line 8:
   8:     { key = "security", value = "yes" },
```

An empty-string violation (for `env`, `team`, or `service`) looks like this:

```hcl
tags = [
  { key = "env", value = "" },
  # ...
]
```

```
$ tflint
Error: 'tsuga_monitor.test' tag 'env' must not be an empty string (camunda_tsuga_monitor_required_tags)

  on test.tf line 8:
   8:     { key = "env", value = "" },
```

## Why

Tsuga monitors (`camunda/terraform`'s `modules/tsuga/monitors`) don't notify anyone directly -
the sibling `modules/tsuga/notifications` module matches on the monitor's `env`/`team` tags to
decide whether, and where, to post a Slack message. A monitor without an `env` tag - or with an
empty one - matches no notification rule's `query_string` and is silently never delivered
anywhere, even if it fires.

Because this only affects delivery - not the monitor's own validity - a plain code review is easy
to miss, and a Terraform `lifecycle.precondition` would have to be copy-pasted into every monitor
resource to have the same effect (and could just as easily be forgotten as the tag itself). This
rule checks every `tsuga_monitor` resource unconditionally, so a new monitor can't skip a required
tag, or misspell a constrained one, without either fixing it or explicitly overriding the rule (see
below).

See `modules/tsuga/monitors/README.md#tagging-convention` in `camunda/terraform` for the full
tagging convention, including what each tag is actually consumed by.

## How To Fix

Add the missing tag(s) to the resource's `tags` list, or fix the value to one of the allowed ones,
e.g.:

```hcl
tags = [
  { key = "env", value = var.environment_tag },
  { key = "team", value = "sre" },
  { key = "service", value = "my-service" },
  { key = "security", value = "false" },
  { key = "managed-by", value = "terraform" },
]
```

If a specific monitor genuinely needs to deviate (there's currently no known legitimate case), you
can either:

* Add a [`tflint-ignore` annotation](https://github.com/terraform-linters/tflint/blob/master/docs/user-guide/annotations.md)
  to skip the rule for this resource.
* Completely disable the rule in the `.tflint.hcl` file with:

  ```hcl
  rule "camunda_tsuga_monitor_required_tags" {
    enabled = false
  }
  ```
