## Overview

A simple terraform wrapper in Go inspired by [domed-city] (https://github.com/ITV/domed-city).


### Features

This tool wraps terraform execution forcing a specific structure while providing some helpful features:

	- Detects AWS account to use based on current working directory.
	- Gets STS tokens and uses them for the current account.
	- Creates an S3 bucket in the current account and enables versioning.
	- Configures remote terraform state on the created S3 bucket.
	- Provides management of remote git/github terraform modules using the Terrafile concept.
	- Provides plan and apply functionality and keeps the local & remote states in sync.


### Setup Requirements

This tool will ask you for input on the first run to configure itself for your user.
Configuration input required:

	- Directory levels from root directory of your project: If your project root directory is `/home/user/projects/my-project` and your terraform templates reside in `/home/user/projects/my-project/terraform/my-account-dev/staging/` then you need to set it to `3`.
	- Name of your project config yaml file: This file will reside on the root directory of your project and needs to be `%name.yaml` which `%name` you specify in this stage.
	- Directory name of your terraform modules: This will be always created a level down of the root of your project and will be used by the Terrafile concept to store your project's modules. Will also be the modules source in your terraform templates.
	- Root profile, is your `$HOME/.aws/credentials` profile name that can assume roles on your AWS accounts


From the files mentioned above here are some examples of what their contents need to be:

`%name.yaml`:

```
project: name_of_your_project
region: eu-west-1
roam-role: roam-role
use-sts: true
encrypt-s3-state: true
accounts-mapping:
    project-dev: 100000000001
    project-prd: 100000000002

```
- `project` should match the name of your AWS accounts without -dev/prd
- `region` is the AWS region your project will be deployed into
- `roam-role` is the AWS IAM role that you can assume in the project's AWS accounts *1
- `use-sts` is a boolean value that enables or disables STS authentication. If not enabled a profile name matching project-dev/prd is expected to be found in your AWS shared credentials file with access and secret keys.
- `encrypt-s3-state` is a boolean value that enables or disables S3 remote state server side encryption.
- `accounts-mapping` is a hash mapping your account-dev/prd used in the project to their AWS account IDS which is needed to assume roles and get STS tokens


*1 More information on how to setup AWS assume roles can be found here: [tutorial] (http://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html) [To create a role for cross-account access (AWS CLI)] (http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user.html#roles-creatingrole-user-cli)


`Terrafile`:

```
tf_aws_asg_elb:
	source: git@github.com:terraform-community-modules/tf_aws_asg_elb.git
	version: v0.1.5
tf_aws_asg:
	source: git@github.com:terraform-community-modules/tf_aws_asg.git
	version: v0.2.1

```
Terrafile Needs to be in a directory level down from the root of your project.

Then on your terraform templates you could use a module like:

```

module "my_asg" {
  source = "%levels_of_dirs_configured - 1/%name_of_tfmodule_directory/tf_aws_asg_elb"

  params.....

}

```

The tool expects to find a `params/env.tfvars` file containing your environment's tfvars to be passed to terraform.




### Beginning with tholos

## Usage

The tool accepts the following parameters:

```
  -a	Terraform Apply Plan
  -c	Force reconfiguration of Tholos
  -o	Display Terraform outputs
  -p	Terraform Plan
  -s	Sync remote S3 state
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
