package main

import (
	"fmt"
	"testing"
)

const fixtures_dir = "./test-fixtures"

func TestProjectConfig(t *testing.T) {

	project_config := load_config(fmt.Sprintf("%s/project.yaml", fixtures_dir))

	if project_config.Parallelism != 4 || project_config.Project != "test" ||
		project_config.Region != "eu-west-1" || !project_config.Use_sts ||
		!project_config.Encrypt_s3_state || len(project_config.Roam_roles[fmt.Sprintf("%s-dev", project_config.Project)]) <= 0 ||
		len(project_config.Accounts_mapping[fmt.Sprintf("%s-dev", project_config.Project)]) <= 0 ||
		len(project_config.Accounts_mapping[fmt.Sprintf("%s-prd", project_config.Project)]) <= 0 ||
		len(project_config.Secondary) != 1 || project_config.Secondary[fmt.Sprintf("%s-dev", project_config.Project)].Account_id != "2001" ||
		project_config.Secondary[fmt.Sprintf("%s-dev", project_config.Project)].Role != "secondary-role-dev" ||
		project_config.Secondary[fmt.Sprintf("%s-dev", project_config.Project)].Region != "eu-west-2" ||
		len(project_config.Profiles) != 2 || project_config.Profiles[fmt.Sprintf("%s-dev", project_config.Project)] != "nproot" {
		t.Fatal("Project configuration parameters in fixtures don't match expected values when parsed.")
	}

}
