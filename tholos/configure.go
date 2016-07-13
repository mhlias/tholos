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
	Root_profile        string `yaml:"root_profile"`
}

func (t *Tholos_config) Configure(force bool, inputs ...string) *Tholos_config {

	tholos_config := &Tholos_config{}

	UserHome := os.Getenv("HOME")

	if len(UserHome) <= 0 {
		log.Fatal("[ERROR] 'HOME' environment variable is not set. Aborting.\n")
	}

	tholos_config_fullpath := fmt.Sprintf("%s/.tholos.yaml", UserHome)

	_, fileexists_err := os.Stat(tholos_config_fullpath)

	if force || fileexists_err != nil {

		if len(inputs) > 0 {

			tmp := strings.Split(inputs[0], ",")

			levelsint, _ := strconv.ParseInt(tmp[0], 10, 32)
			tholos_config.Levels = int(levelsint)

			tholos_config.Project_config_file = tmp[2]
			tholos_config.Tf_modules_dir = tmp[1]
			tholos_config.Root_profile = tmp[3]

		} else {

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

			fmt.Print("Name of your root AWS account profile in aws config/credentials: ")
			root_profile, _ := reader.ReadString('\n')
			tholos_config.Root_profile = strings.TrimRight(root_profile, "\n")

		}

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
