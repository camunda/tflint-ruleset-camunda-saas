# TFLint Ruleset for Camunda SaaS

This is a set of custom [TFLint](https://github.com/terraform-linters/tflint)
rules to run against the Terraform configuration for Camunda SaaS.


## Installation

You can install the plugin with `make install`, or by copying the binary
`tflint-ruleset-camunda-saas` into `~/.tflint.d/plugins`.

Then, enable it with:

```hcl
plugin "camunda-saas" {
  enabled = true
  version = "v1.1.0"
  source  = "github.com/camunda-cloud/tflint-ruleset-camunda-saas"
}
```

## Rules

Open the [documentation](./docs/README.md) to get the list of rules.


## Development

### Making a release

1. Create a new Git tag, using the `vX.Y.Z` format:

   ```
   git tag --annotate --sign --message "Release vX.Y.Z" vX.Y.Z
   ```

1. Push the new tag to GitHub:

   ```
   git push origin vX.Y.Z
   ```

1. GitHub Actions should take care of creating the artifacts and creating the GitHub Releases.
1. Update the TFLint configuration file to use the new version:

   ```hcl
   plugin "camunda-saas" {
     enabled = true
     version = "vX.Y.Z"
     source  = "github.com/camunda-cloud/tflint-ruleset-camunda-saas"
   }
   ```



### Building the plugin

Clone the repository locally and run the following command:

```
$ make
```

You can easily install locally the built plugin with the following:

```
$ make install
```


### Testing the plugin

Run `make install` then run `tflint` from one of the `tests` subdirectory.
