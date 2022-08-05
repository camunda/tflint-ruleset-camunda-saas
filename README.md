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
}
```

## Rules

Open the [documentation](./docs/README.md) to get the list of rules.


## Development


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
