## Overview

A simple terraform wrapper for AWS use in Go inspired by [domed-city] (https://github.com/ITV/domed-city).
WARNING Version v1.0 has breaking changes due to restructure of configuration syntax.

*Warning from Tholos version v1.0.0 config backwards compatibility is broken*

### Features

This tool wraps terraform execution forcing a specific structure while providing some helpful features:

	- Detects AWS account to use based on current working directory.
	- Gets STS tokens and uses them for the current account.
	- Creates an S3 bucket in the current account and enables versioning.
	- Configures remote terraform state on the created S3 bucket.
	- Creates a DynamoDB lock table and uses it to lock remote S3 state (Terraform 0.9.x only)
	- Provides management of remote git/github terraform modules using the Terrafile concept.
	- Provides plan and apply functionality with resources target support and keeps the local & remote states in sync.
	- Support for Terraform 0.9.x and legacy mode with version autodetection
	- Support for Terraform state environments (created if not exist already) with Terraform 0.9.x versions


### Setup Requirements

This tool will ask you for input on the first run to configure itself for your user.
Configuration input required:

	- Name of your project config yaml file: This file will reside on a directory up to 3 levels back from the current working directory and needs to be `%name.yaml` which `%name` you specify in this stage.
	- Directory name of your terraform modules: This will be always created 1 directory level before your current working directory and will be used by the Terrafile concept to store your project's modules. Will also be the modules source in your terraform templates.
	- With Terraform 0.9.x you need to include in a .tf file the following terraform block:

	```
	terraform {
	    required_version = ">= 0.9.0"
	    backend "s3" {}
	}

	```

The configured file is in your `$HOME/.tholos.yaml`
Example contents:

```
tf_modules_dir: tfmodules
project_config_file: project.yaml

```



From the files mentioned above here are some examples of what their contents need to be:

`project.yaml`:

```
project: testproject
region: eu-west-1
encrypt-s3-state: true
accounts:
  test-dev:
    profile: nproot
    account_id: 1001
    roaming-role: roam-role-dev
    secondary:
      id: 2001
      role: secondary-role-dev
      region: eu-west-2
  test-prd:
    profile: proot
parallelism: 4

```
- `project` should match the name of your AWS account with any -suffix allowed
- `region` is the AWS region your project will be deployed into
- `profiles` is the AWS shared credentials profile to use to assume the role for each account or authenticate if sts is not in use
- `roam-role` is the AWS IAM role that you can assume in the project's AWS accounts *1
- `use-sts` is a boolean value that enables or disables STS authentication. If not enabled a profile name matching project-%suffix% is expected to be found in your AWS shared credentials file with access and secret keys.
- `encrypt-s3-state` is a boolean value that enables or disables S3 remote state server side encryption.
- `accounts-mapping` is a hash mapping your account-%suffix% used in the project to their AWS account IDS which is needed to assume roles and get STS tokens
- `parallelism` is the Terraform parallelism setting for refresh of the plan defaults to 10 if omitted
- `secondary` is for secondary account resources with all of `id`, `role` and `region` required for each secondary account.

An example configuration with a single AWS account configured:

```
---
project: PROJECT_NAME
region: eu-west-1
encrypt-s3-state: true
accounts:
  test-dev:
    profile: js-mgmt
    account_id: ACCOUNT_ID
    roam-role: ROLE_NAME
```


So based on the above your working directories structure would look like:

```
.
├── project.yaml
└── test-dev <- account name specified in project.yaml
    └── test <- state/environment name can be anything and where you run tholos into
        ├── main.tf
        ├── params <- name of the directory where your state/env tfvars file will be in
        │   └── env.tfvars <- needs to be named that
        ├── plans <- needs to be created the first time
            └── plan.tfplan

```

*1 More information on how to setup AWS assume roles can be found here: [tutorial] (http://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html) [To create a role for cross-account access (AWS CLI)] (http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user.html#roles-creatingrole-user-cli)

### Cross-account resources

From v0.9.0 there is support for a secondary account per primary account to allow to provision AWS resources across different AWS accounts. Can only be used with STS tokens. A requirement is the AWS shared credentials profile that is used on the primary account to be able to assume the role specified on the secondary account. `project.yaml` configuration example follows:
See the config example above on what you need to provide for a secondary account.

Then at runtime tholos will export the following environment variables that Terraform can pick up:

- TF_VAR_secondary_access_key_id
- TF_VAR_secondary_secret_access_key
- TF_VAR_secondary_security_token
- TF_VAR_secondary_session_token
- TF_VAR_secondary_region

From there in your Terraform code you can have a provider entry for the secondary account like:

```
provider "aws" {
  alias = "secondary"
  region = "${var.secondary_region}"
  access_key = "${var.secondary_access_key_id}"
  secret_key = "${var.secondary_secret_access_key}"
  token = "${var.secondary_session_token}"
}

```

`Terrafile`:

```
tf_aws_asg_elb:
	source: git@github.com:terraform-community-modules/tf_aws_asg_elb.git
	version: v0.1.5
tf_aws_asg:
	source: git@github.com:terraform-community-modules/tf_aws_asg.git
	version: v0.2.1

```
Terrafile Needs to be residing 2 directory levels before your current working directory/environment.

Then on your terraform templates you could use a module like:

```

module "my_asg" {
  source = "../../%name_of_tfmodule_directory/tf_aws_asg_elb"

  params.....

}

```

The tool expects to find a `params/env.tfvars` file containing your environment's tfvars to be passed to terraform.




### Beginning with tholos

## Usage

When you start working on a new environment you need to run `tholos -init` for the first time, then you can start plan and applying with `tholos -p` and `tholos -a` respectively.

The tool accepts the following parameters:

```
  -a	Terraform Apply Plan
  -c	Force reconfiguration of Tholos
  -e  Terraform state environment to use
  -o	Display Terraform outputs
  -p	Terraform Plan
  -init	Initialize remote S3 bucket and state
  -t  Terraform resources to target only, (-t resourcetype.resource resourcetype2.resource2)
  -u	Fetch and update modules from remote repo

```

If you have setup MFA access to your AWS accounts as a requirement, then you can set environment variables `MFA_DEVICE_ID` and `MFA_TOKEN` to the mfa device id registered with your iam user and the current mfa token respectively.


### Limitations

This tool is heavily opinionated in trying to enforce a structure in the way terraform is used. It may not be useful to anyone else if you do not want to conform to that structure.

### Acknowledgements

Credits and thanks to the following people and old colleagues from ITV:

- [Efstathios Xagoraris] (https://github.com/xiii) for coming up with `Terrafile` which tholos implements.
- [Ben Snape] (https://github.com/bsnape) for the `domed-city` implementation which tholos is inspired from.
- [Stefan Coccora] (https://github.com/stefancocora) & [Andrew Stangl] (https://github.com/madandroid) for their original concept and implementation that evolved to `domed-city`.
