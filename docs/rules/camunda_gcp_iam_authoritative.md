# `camunda_gcp_iam_authoritative`

Disallow to use the authoritative IAM resources in Google Cloud Platform.


## Examples

### `google_project_iam_policy`

```hcl
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
```

```
$ tflint
1 issue(s) found:

Error: 'google_project_iam_policy' is a dangerous resource, are you sure you want to use it? (camunda_gcp_iam_authoritative)

  on test.tf line 8:
   8: resource "google_project_iam_policy" "test" {
```


### `google_project_iam_binding`

```hcl
resource "google_project_iam_binding" "test" {
  project = "your-project-id"

  role    = "roles/test"
  members = []
}
```

```
$ tflint
1 issue(s) found:

Error: 'google_project_iam_binding' is a dangerous resource, are you sure you want to use it? (camunda_gcp_iam_authoritative)

  on test2.tf line 1:
   1: resource "google_project_iam_binding" "test" {
```


## Why

The following Terraform resources for Google Cloud Platform are
**authoritative** resources:

* [`google_project_iam_policy`](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/google_project_iam)

  This is authoritative for **all the project**: all the permissions not
  configured with this resource will be removed.
* [`google_project_iam_binding`](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/google_project_iam)

  This is authoritative for **a specific role**: all the members not configured
  with this will have this role revoked from them.

These resources can be very distruptive if not used correctly, with their
expected effects clearly understood, and they can completely lock you out of a GCP project.


## How To Fix

Most of the time, you want to use
[`google_project_iam_member`](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/google_project_iam#google_project_iam_member) instead.

If you **really** want to use one of these resources, and you really know what
you are doing, you can either:

* Add a [`tflint-ignore` annotation](https://github.com/terraform-linters/tflint/blob/master/docs/user-guide/annotations.md)
  to skip the rule for this resource.
* Completely disable the rule in the `.tflint.hcl` file with:

  ```hcl
  rule "camunda_gcp_iam_authoritative" {
    enabled = false
  }
  ```
