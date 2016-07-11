package tf_helper

import (
	"log"
)

func (c *Config) Outputs() {

	c.Setup_remote_state()

	cmd_name := "terraform"

	exec_args := []string{"output"}

	log.Println("[INFO] Displaying Terraform outputs.")

	if !ExecCmd(cmd_name, exec_args) {
		log.Fatal("[ERROR] Failed to get Terraform outputs. Aborting.")
	}

}
