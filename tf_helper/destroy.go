package tf_helper

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func (c *Config) Destroy(parallelism int16) {

	//We want to really ensure destroy is what you want to do, therefore we double check here, requiring an input from CLI

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("You are running Destroy! Are you sure you want to proceed? Confirm by entering Destroy: ")

	input, _ := reader.ReadString('\n')

	if strings.TrimRight(strings.ToLower(input), "\n") == "destroy" {

		log.Println("[INFO] Proceeding with Terraform destroy.")

	} else {

		log.Println("[INFO] Not proceeding with Terraform destroy.")
		os.Exit(1)

	}

	cmd_name := "terraform"

	exec_args := []string{"destroy", fmt.Sprintf("-parallelism=%d", parallelism)}

	exec_args = append(exec_args, []string{"-refresh=true", "-var-file=params/env.tfvars"}...)

	if len(c.TargetsTF) > 0 {
		for _, t := range c.TargetsTF {
			exec_args = append(exec_args, fmt.Sprintf("-target=%s", t))
		}
	}

	exec_args = append(exec_args, "-auto-approve")

	if !ExecCmd(cmd_name, exec_args) {
		log.Fatal("[ERROR] Failed to run Terraform destroy. Aborting.")
	}

}
