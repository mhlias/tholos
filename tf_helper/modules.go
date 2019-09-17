package tf_helper

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/JSainsburyPLC/tholos/tholos"
)

type Modules struct {
	Name map[string]struct {
		Source  string
		Version string
	}
}

func (m *Modules) Fetch_modules(tholos_conf *tholos.Tholos_config) {

	modulesFile, _ := filepath.Abs(fmt.Sprintf("../../Terrafile"))
	yamlModules, file_err := ioutil.ReadFile(modulesFile)

	if file_err != nil {
		log.Fatal("[ERROR] File does not exist or not accessible: ", file_err)
	}

	yaml_err := yaml.Unmarshal(yamlModules, &m.Name)

	if yaml_err != nil {
		log.Fatal("[ERROR] Failed to parse Terrafile yaml: ", yaml_err)
	}

	cmd_name := "rm"

	exec_args := []string{"-rf", fmt.Sprintf("../../%s", tholos_conf.Tf_modules_dir)}

	if len(tholos_conf.Tf_modules_dir) > 0 && !strings.Contains(tholos_conf.Tf_modules_dir, "*") {

		log.Println("[INFO] Cleaning up old Terraform modules.")

		if !ExecCmd(cmd_name, exec_args) {
			log.Fatal("[ERROR] Failed to clean up old Terraform modules. Aborting.")
		}

	} else {
		log.Println("[WARN] Didn't clean up old terraform modules directory as directory name not set or contains wildcard character.")
	}

	cmd_name = "mkdir"

	exec_args = []string{"-p", fmt.Sprintf("../../%s", tholos_conf.Tf_modules_dir)}

	log.Println("[INFO] Creating Terraform modules directory (if not present already).")

	if !ExecCmd(cmd_name, exec_args) {
		log.Fatal("[ERROR] Failed to create Terrform modules directory. Aborting.")
	}

	log.Println("[INFO] Fetching Terraform modules and updating existing ones.")

	for name, module := range m.Name {

		cmd_name := "git"

		exec_args := []string{"clone",
			"-b",
			module.Version,
			module.Source,
			fmt.Sprintf("../../%s/%s", tholos_conf.Tf_modules_dir, name),
		}

		if !ExecCmd(cmd_name, exec_args) {
			log.Fatal("[ERROR] Failed to fetch Terraform modules from remote. Aborting.")
		}

	}

}
