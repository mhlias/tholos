package main

import (
	"fmt"
	"testing"
)

const fixtures_dir = "./test-fixtures"

func TestProjectConfigA(t *testing.T) {

	project_config, loaded := load_config(fmt.Sprintf("%s/project.yaml", fixtures_dir))

	if !loaded {
		t.Fatal("Project config file failed to load.")
	}

	if project_config.Parallelism != 4 || project_config.Project != "test" ||
		project_config.Region != "eu-west-1" || !project_config.Encrypt_s3_state ||
		len(project_config.Accounts) != 2 ||
		project_config.Accounts[fmt.Sprintf("%s-dev", project_config.Project)].AccountID != "1001" ||
		project_config.Accounts[fmt.Sprintf("%s-dev", project_config.Project)].RoamRole != "roam-role-dev" ||
		project_config.Accounts[fmt.Sprintf("%s-dev", project_config.Project)].Profile != "nproot" ||
		project_config.Accounts[fmt.Sprintf("%s-prd", project_config.Project)].Profile != "proot" ||
		project_config.Accounts[fmt.Sprintf("%s-dev", project_config.Project)].Secondary.Account_id != "2001" ||
		project_config.Accounts[fmt.Sprintf("%s-dev", project_config.Project)].Secondary.Role != "secondary-role-dev" ||
		project_config.Accounts[fmt.Sprintf("%s-dev", project_config.Project)].Secondary.Region != "eu-west-2" {
		t.Fatal("Project configuration parameters in fixtures don't match expected values when parsed.")
	}

}
