package tf_helper

import (
	"log"
)

func (c *Config) Apply() {

	cmd_name := "terraform"

	exec_args := []string{"apply", "plans/plan.tfplan"}

	log.Println("[INFO] Applying Terraform plan.")

	if !ExecCmd(cmd_name, exec_args) {
		log.Fatal("[ERROR] Failed to apply Terraform plan. Aborting.")
	}

}
