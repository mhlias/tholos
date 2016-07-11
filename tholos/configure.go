package tholos

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Tholos_config struct {
	Levels              int    `yaml:"levels"`
	Tf_modules_dir      string `yaml:"tf_modules_dir"`
	Project_config_file string `yaml:"project_config_file"`
}

func (t *Tholos_config) Configure(force bool) *Tholos_config {

	tholos_config := &Tholos_config{}

	UserHome := os.Getenv("HOME")

	if len(UserHome) <= 0 {
		log.Fatal("[ERROR] 'HOME' environment variable is not set. Aborting.\n")
	}

	tholos_config_fullpath := fmt.Sprintf("%s/.tholos.yaml", UserHome)

	_, fileexists_err := os.Stat(tholos_config_fullpath)

	if force || fileexists_err != nil {

		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Configuring Tholos for the first time.")
		fmt.Print("Number of directory levels from root of project to where your terraform templates will reside: ")
		levels, _ := reader.ReadString('\n')
		levelsint, _ := strconv.ParseInt(strings.TrimRight(levels, "\n"), 10, 32)
		tholos_config.Levels = int(levelsint)

		fmt.Print("Name of yaml project config file (including .yaml) in project root directory: ")
		project_config_file, _ := reader.ReadString('\n')
		tholos_config.Project_config_file = strings.TrimRight(project_config_file, "\n")

		fmt.Print("Name of directory your terraform remote modules will be stored in: ")
		tf_modules_dir, _ := reader.ReadString('\n')
		tholos_config.Tf_modules_dir = strings.TrimRight(tf_modules_dir, "\n")

		yaml_out, marshal_err := yaml.Marshal(tholos_config)

		if marshal_err != nil {
			log.Fatal("[ERROR] Failed marshalling struct to yaml: ", marshal_err)
		}

		save_err := ioutil.WriteFile(tholos_config_fullpath, yaml_out, 0600)

		if save_err != nil {
			log.Fatal("[ERROR] Failed to save tholos configuration to %s. Aborting with error: %s", tholos_config_fullpath, save_err)
		}

	} else {

		configFile, _ := filepath.Abs(tholos_config_fullpath)
		yamlConf, file_err := ioutil.ReadFile(configFile)

		if file_err != nil {
			log.Fatalf("[ERROR] File does not exist or not accessible: ", file_err)
		}

		yaml_err := yaml.Unmarshal(yamlConf, &tholos_config)

		if yaml_err != nil {
			log.Fatal(yaml_err)
		}

	}

	return tholos_config

}
