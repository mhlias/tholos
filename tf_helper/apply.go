package tf_helper

import (
	"fmt"
	"log"
)

func (c *Config) Apply() {

	cmd_name := "terraform"

	exec_args := []string{"apply", "plans/plan.tfplan"}

	if len(c.TargetsTF) > 0 {
		for _, t := range c.TargetsTF {
			exec_args = append(exec_args, fmt.Sprintf("-target=%s", t))
		}
	}

	log.Println("[INFO] Applying Terraform plan.")

	if !ExecCmd(cmd_name, exec_args) {
		log.Fatal("[ERROR] Failed to apply Terraform plan. Aborting.")
	}

}
