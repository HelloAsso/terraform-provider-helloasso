# Terraform Provider HelloAsso

This repository is built on the [Terraform Plugin Framework](https://github.com/hashicorp/ From template [Terraform Provider Scaffolding](https://github.com/hashicorp/terraform-provider-helloasso)


## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.18
- An app registration in Azure AD having role permission Azure Devops

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

Run "go generate" to format example terraform files and generate the docs for the registry/website

If you do not have terraform installed, you can remove the formatting command, but its suggested to ensure the documentation is formatted properly.
`go:generate terraform fmt -recursive ./examples/`

Run the docs generation tool, check its repository for more information on how it works and how docs can be customized.
`go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs`
