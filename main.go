package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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
	Encrypt_s3_state bool  `yaml:"encrypt-s3-state"`
	Parallelism      int16 `yaml:"parallelism"`
	environment      string
	account          string
	Accounts         map[string]aws_helper.Account `yaml:"accounts"`
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

const CLR_0 = "\x1b[30;1m"
const CLR_R = "\x1b[31;1m"
const CLR_G = "\x1b[32;1m"
const CLR_Y = "\x1b[33;1m"
const CLR_B = "\x1b[34;1m"
const CLR_M = "\x1b[35;1m"
const CLR_C = "\x1b[36;1m"
const CLR_W = "\x1b[37;1m"
const CLR_N = "\x1b[0m"

func main() {

	use_mfa := true
	retries := 3

	planPtr := flag.Bool("p", false, "Terraform Plan")
	applyPtr := flag.Bool("a", false, "Terraform Apply Plan")
	initPtr := flag.Bool("init", false, "Initialize project S3 bucket state")
	modulesPtr := flag.Bool("u", false, "Fetch and update modules from remote repo")
	outputsPtr := flag.Bool("o", false, "Display Terraform outputs")
	configPtr := flag.Bool("c", false, "Force reconfiguration of Tholos")
	envPtr := flag.String("e", "", "Terraform state environment to use")
	flag.Var(&targetsTF, "t", "Terraform resources to target only, (-t resourcetype.resource resourcetype2.resource2)")

	flag.Parse()

	tholos := &tholos.Tholos_config{}

	tholos_conf := tholos.Configure(*configPtr)

	if !*planPtr && !*initPtr && !*modulesPtr && !*outputsPtr && !*applyPtr {
		fmt.Println("Please provide one of the following parameters:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	config_loaded, found := false, false
	project_config := &conf{}

	for l := 1; l < 4; l++ {

		dir_levels := strings.Repeat("../", l)

		project_config, found = load_config(fmt.Sprintf("%s%s", dir_levels, tholos_conf.Project_config_file))

		if found {
			config_loaded = true
			break
		}

	}

	if !config_loaded {
		log.Fatalf("Project config file with name %s couldn't be found up to 3 levels back from current directory. Aborting..", tholos_conf.Project_config_file)
	}

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

	fmt.Printf("%sWORKING ON AWS account: %s and environment: %s\nUsing shared credentials profile %s and assuming IAM role %s on AWS account with ID %s\n", CLR_G, project_config.account, project_config.environment, project_config.Accounts[project_config.account].Profile, project_config.Accounts[project_config.account].RoamRole, project_config.Accounts[project_config.account].AccountID)

	time.Sleep(3 * time.Second)

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

	tf_version := get_tf_version()

	tf_legacy := true

	tf_lock_legacy := true

	log.Printf("[INFO] Terraform Version found: %s\n", tf_version)

	ver_int, _ := strconv.Atoi(strings.Split(tf_version, ".")[1])
	ver_patch_int, _ := strconv.Atoi(strings.Split(tf_version, ".")[2])

	if ver_int > 8 {
		tf_legacy = false
		if len(*envPtr) > 0 {
			log.Printf("[INFO] Will be working on STATE ENVIRONMENT: %s", *envPtr)
			// Sleep for 5 seconds let the user stop execution if wrong state environment
			time.Sleep(5 * time.Second)
		}
	} else {
		log.Printf("[WARN] Running in legacy mode, current Terraform version: %s, install >=0.9.x for full features.\n", tf_version)
	}

	if ver_int >= 10 || (ver_int > 9 && ver_patch_int >= 7) {
		tf_lock_legacy = false
	}

	state_config := &tf_helper.Config{Bucket_name: fmt.Sprintf("%s-%s-%s-tfstate", project_config.Project, project_config.account, project_config.environment),
		State_filename:   fmt.Sprintf("%s-%s-%s.tfstate", project_config.Project, project_config.account, project_config.environment),
		Lock_table:       fmt.Sprintf("%s-%s-%s-locktable", project_config.Project, project_config.account, project_config.environment),
		Versioning:       true,
		Region:           project_config.Region,
		Encrypt_s3_state: project_config.Encrypt_s3_state,
		TargetsTF:        targetsTF,
		TFlegacy:         tf_legacy,
		TFLockLegacy:     tf_lock_legacy,
		TFenv:            *envPtr,
	}

	var tf_parallelism int16 = 10

	if &project_config.Parallelism != nil && project_config.Parallelism > 0 {
		tf_parallelism = project_config.Parallelism
	}

	modules := &tf_helper.Modules{}

	var client interface{}

	if !*modulesPtr {

		awsconf := &aws_helper.Config{
			Region:        project_config.Region,
			Use_mfa:       use_mfa,
			Mfa_device_id: mfa_device_id,
			Mfa_token:     mfa_token,
			AWSAccount:    project_config.Accounts[project_config.account],
		}

		client = awsconf.Connect()

	}

	if *initPtr || *planPtr || *applyPtr {

		if !state_config.TFlegacy {

			if len(*envPtr) > 0 {
				state_config.Switch_env()
			}

			for j := 1; j <= retries; j++ {

				if !state_config.Create_locktable(client) {
					log.Printf("[WARN] DynamoDB table %s failed to be created. Retrying.\n", state_config.Lock_table)
				} else {
					log.Printf("[INFO] DynamoDB table %s created.\n", state_config.Lock_table)
					break
				}

				time.Sleep(time.Duration(j) * time.Second)

			}
		}

	}

	if *initPtr {

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
		state_config.Plan(tholos_conf, tf_parallelism)
	} else if *modulesPtr {
		modules.Fetch_modules(tholos_conf)
	} else if *outputsPtr {
		state_config.Outputs()
	} else if *applyPtr {
		state_config.Apply()
	}

}

func load_config(project_config_file string) (*conf, bool) {

	project_config := &conf{}

	configFile, _ := filepath.Abs(project_config_file)
	yamlConf, file_err := ioutil.ReadFile(configFile)

	if file_err != nil {
		return project_config, false
	}

	yaml_err := yaml.Unmarshal(yamlConf, &project_config)

	if yaml_err != nil {
		log.Fatal(yaml_err)
	}

	return project_config, true

}

func get_tf_version() string {

	cmd := exec.Command("terraform", "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal("Failed to get Terraform version, please make sure terraform is installed and in the path", err)
	}

	out_str := out.String()

	start := strings.Index(out_str, "v")

	ver := out_str[start : start+7]

	return ver
}
