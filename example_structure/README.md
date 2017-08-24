## Overview

Example structure for testing.


### Usage of example


You need to put the following contents in your `$HOME/.tholos.yaml`


```
tf_modules_dir: tfmodules
project_config_file: project.yaml

```

Edit `project.yaml` and replace AWS shared credentials profile names and account id and IAM role to assume to match your own.


Then on each of the environments `myenv` and `otherenv` you need to run `tholos -init` for the first time. Then you can start plan and applying with `tholos -p` and `tholos -a` respectively.
