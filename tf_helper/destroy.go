package tf_helper

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const CLR_0 = "\x1b[30;1m"
const CLR_R = "\x1b[31;1m"
const CLR_G = "\x1b[32;1m"
const CLR_Y = "\x1b[33;1m"
const CLR_B = "\x1b[34;1m"
const CLR_M = "\x1b[35;1m"
const CLR_C = "\x1b[36;1m"
const CLR_W = "\x1b[37;1m"
const CLR_N = "\x1b[0m"

func (c *Config) Destroy(parallelism int16) {

	//We want to really ensure destroy is what you want to do, therefore we double check here, requiring an input from CLI

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%sYou are running Destroy! Are you sure you want to proceed? Confirm by entering \"destroy\": ", CLR_R)

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
