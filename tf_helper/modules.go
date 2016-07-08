package tf_helper

import (
	"fmt"
	"log"
	"io/ioutil"
  "path/filepath"

	"gopkg.in/yaml.v2"
)


type Modules struct {
	Name map[string] struct {
		Source string
		Version string
	}
}



func (m *Modules) Fetch_modules() {



	modulesFile, _ := filepath.Abs("../../Terrafile")
  yamlModules, file_err := ioutil.ReadFile(modulesFile)

  if file_err != nil {
    log.Fatalf("[ERROR] File does not exist or not accessible: ", file_err)
  }

  yaml_err := yaml.Unmarshal(yamlModules, &m.Name)

  if yaml_err != nil {
    log.Fatal("[ERROR] Failed to parse Terrafile yaml: ", yaml_err)
  }

  cmd_name := "rm"

  exec_args := []string { "-rf", "../../modules" }

  log.Println("[INFO] Cleaning up old Terraform modules.")

  if !ExecCmd(cmd_name, exec_args) {
  	log.Fatal("[ERROR] Failed to clean up old Terraform modules. Aborting.")
  }

  cmd_name = "mkdir"

  exec_args = []string { "-p", "../../modules" }

  log.Println("[INFO] Creating Terraform modules directory (if not present already).")

	if !ExecCmd(cmd_name, exec_args) {
		log.Fatal("[ERROR] Failed to create Terrform modules directory. Aborting.")
	}

	log.Println("[INFO] Fetching Terraform modules and updating existing ones.")

  for name, module := range m.Name {

	  cmd_name := "git"

	  exec_args := []string { "clone", 
	  												"-b", 
	  												module.Version, 
	  												module.Source, 
	  												fmt.Sprintf("../../modules/%s", name),
	  											}

	  if !ExecCmd(cmd_name, exec_args) {
	  	log.Fatal("[ERROR] Failed to fetch Terraform modules from remote. Aborting.")
	  }

  }




}