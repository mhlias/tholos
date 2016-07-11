package tf_helper

import (
	"log"

	"github.com/mhlias/tholos/tholos"
)

func (c *Config) Plan(tholos_conf *tholos.Tholos_config) {

	cmd_name := "rm"

	exec_args := []string{"-rf", ".terraform"}

	log.Println("[INFO] Deleting Terraform cache directory.")

	if !ExecCmd(cmd_name, exec_args) {
		log.Fatal("[ERROR] Failed to delete Terraform cache directory. Aborting.")
	}

	cmd_name = "rm"

	exec_args = []string{"-f", "plans/plan.tfplan"}

	log.Println("[INFO] Deleting Terraform old plan.")

	if !ExecCmd(cmd_name, exec_args) {
		log.Fatal("[ERROR] Failed to delete Terraform old plan. Aborting.")
	}

	cmd_name = "terraform"

	exec_args = []string{"get", "-update=true"}

	log.Println("[INFO] Fetching Terraform modules and updating existing ones.")

	if !ExecCmd(cmd_name, exec_args) {
		log.Fatal("[ERROR] Failed to fetch Terraform modules. Aborting.")
	}

	c.Setup_remote_state()

	exec_args = []string{"plan", "-module-depth=1", "-refresh=true", "-out=plans/plan.tfplan", "-var-file=params/env.tfvars"}

	log.Println("[INFO] Running Terraform plan.")

	if !ExecCmd(cmd_name, exec_args) {
		log.Fatal("[ERROR] Failed to run Terraform plan. Aborting.")
	}

}
