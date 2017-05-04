package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/mhlias/tholos/aws_helper"
	"github.com/mhlias/tholos/tf_helper"
	"github.com/mhlias/tholos/tholos"
)

type conf struct {
	Project          string
	Region           string
	Roam_roles       map[string]string `yaml:"roam-roles"`
	Accounts_mapping map[string]string `yaml:"accounts-mapping"`
	Use_sts          bool              `yaml:"use-sts"`
	Encrypt_s3_state bool              `yaml:"encrypt-s3-state"`
	environment      string
	account          string
}

type multiflag []string

func (d *multiflag) String() string {
	return fmt.Sprintf("%d", *d)
}

func (d *multiflag) Set(value string) error {
	*d = append(*d, value)
	return nil
}

var targetsTF multiflag

func main() {

	use_mfa := true
	retries := 3

	planPtr := flag.Bool("p", false, "Terraform Plan")
	applyPtr := flag.Bool("a", false, "Terraform Apply Plan")
	syncPtr := flag.Bool("s", false, "Sync remote S3 state")
	modulesPtr := flag.Bool("u", false, "Fetch and update modules from remote repo")
	outputsPtr := flag.Bool("o", false, "Display Terraform outputs")
	configPtr := flag.Bool("c", false, "Force reconfiguration of Tholos")
	flag.Var(&targetsTF, "t", "Terraform resources to target only, (-t resourcetype.resource resourcetype2.resource2)")

	flag.Parse()

	tholos := &tholos.Tholos_config{}

	tholos_conf := tholos.Configure(*configPtr)

	if !*planPtr && !*syncPtr && !*modulesPtr && !*outputsPtr && !*applyPtr {
		fmt.Println("Please provide one of the following parameters:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	dir_levels := strings.Repeat("../", tholos_conf.Levels)

	project_config := load_config(fmt.Sprintf("%s%s", dir_levels, tholos_conf.Project_config_file))

	curr_dir, err := os.Getwd()

	if err != nil {
		log.Fatal("[ERROR] Failed to get current working directory. Aborting with error: ", err)
	}

	dir_separator := ""

	if runtime.GOOS == "windows" {
		dir_separator = "\\"
	} else {
		dir_separator = "/"
	}

	tmp := strings.Split(curr_dir, dir_separator)

	project_config.environment = tmp[len(tmp)-1]
	project_config.account = tmp[len(tmp)-2]

	mfa_device_id := os.Getenv("MFA_DEVICE_ID")
	mfa_token := ""

	if len(mfa_device_id) <= 0 {
		log.Println("[INFO] No mfa device id is set in the env. Set `MFA_DEVICE_ID` in your environment if you want to use one.")
		use_mfa = false
	} else {
		log.Printf("[INFO] MFA device with id: %s was detected in the environment. Using it.\n", mfa_device_id)
		mfa_token = os.Getenv("MFA_TOKEN")
		if len(mfa_token) >= 0 {
			use_mfa = true
		} else {
			use_mfa = false
			log.Println("[INFO] No mfa token was provided in the env. Set `MFA_TOKEN` in your environment if you want to use one.")
		}
	}

	if len(project_config.Project) <= 0 {
		log.Fatal("[ERROR] No project is set in your project.yaml configuration.")
	}

	accounts := map[string]bool{fmt.Sprintf("%s-dev", project_config.Project): true,
		fmt.Sprintf("%s-prd", project_config.Project): true,
	}

	state_config := &tf_helper.Config{Bucket_name: fmt.Sprintf("%s-%s-%s-tfstate", project_config.Project, project_config.account, project_config.environment),
		State_filename:   fmt.Sprintf("%s-%s-%s.tfstate", project_config.Project, project_config.account, project_config.environment),
		Versioning:       true,
		Encrypt_s3_state: project_config.Encrypt_s3_state,
		TargetsTF:        targetsTF,
	}

	modules := &tf_helper.Modules{}

	if _, ok := accounts[project_config.account]; !ok {
		log.Fatalf("[ERROR] Account directories do not match project name. Name found: %s, expected %s-dev or %s-prd\n", project_config.account, project_config.Project, project_config.Project)
	}

	var client interface{}

	if !*modulesPtr {

		profile := project_config.account

		if project_config.Use_sts {
			profile = tholos_conf.Root_profile
		}

		awsconf := &aws_helper.Config{
			Region:        project_config.Region,
			Profile:       profile,
			Role:          project_config.Roam_roles[project_config.account],
			Account_id:    project_config.Accounts_mapping[project_config.account],
			Use_mfa:       use_mfa,
			Use_sts:       project_config.Use_sts,
			Mfa_device_id: mfa_device_id,
			Mfa_token:     mfa_token,
		}

		client = awsconf.Connect()

	}

	if *syncPtr {

		bucket_created := false

		for i := 1; i <= retries; i++ {

			if !state_config.Create_bucket(client) {
				log.Printf("[WARN] S3 Bucket %s failed to be created. Retrying.\n", state_config.Bucket_name)
			} else {
				log.Printf("[INFO] S3 Bucket %s created and versioning enabled.\n", state_config.Bucket_name)
				bucket_created = true
				break
			}

			time.Sleep(time.Duration(i) * time.Second)

		}

		if bucket_created {
			state_config.Setup_remote_state()
		} else {
			log.Fatalf("[ERROR] S3 Bucket failed to be created after %d retries. Aborting.\n", retries)
		}

	} else if *planPtr {
		state_config.Plan(tholos_conf)
	} else if *modulesPtr {
		modules.Fetch_modules(tholos_conf)
	} else if *outputsPtr {
		state_config.Outputs()
	} else if *applyPtr {
		state_config.Apply()
	}

}

func load_config(project_config_file string) *conf {

	project_config := &conf{}

	configFile, _ := filepath.Abs(project_config_file)
	yamlConf, file_err := ioutil.ReadFile(configFile)

	if file_err != nil {
		log.Fatalln("[ERROR] File does not exist or not accessible: ", file_err)
	}

	yaml_err := yaml.Unmarshal(yamlConf, &project_config)

	if yaml_err != nil {
		log.Fatal(yaml_err)
	}

	return project_config

}
