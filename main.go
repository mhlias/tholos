package main



import (
 "fmt"
 "os"
 "strings"
 "log"
 "io/ioutil"
 "path/filepath"
 "runtime"
 "time"

 "gopkg.in/yaml.v2"

 "github.com/mhlias/tholos/aws_helper"
 "github.com/mhlias/tholos/tf_helper"


)

type conf struct {
  Project string
  Region  string
  Root_profile string                 `yaml:"root-profile"`
  Roam_role string                    `yaml:"roam-role"`
  Accounts_mapping map[string] string `yaml:"accounts-mapping"`
  environment string
  account string
}


func main() {


  use_mfa := true
  retries := 3

  project_config := new(conf)

  configFile, _ := filepath.Abs("../../../project.yaml")
  yamlConf, file_err := ioutil.ReadFile(configFile)

  if file_err != nil {
    log.Fatalf("[ERROR] File does not exist or not accessible: ", file_err)
  }

  yaml_err := yaml.Unmarshal(yamlConf, &project_config)

  if yaml_err != nil {
    log.Fatal(yaml_err)
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
  project_config.account     = tmp[len(tmp)-2]

  mfa_device_id := os.Getenv("MFA_DEVICE_ID")

  if len(mfa_device_id) <= 0 {
    log.Println("[INFO] No mfa device id is set in the env. Set `MFA_DEVICE_ID` in your environment if you want to use one.")
    use_mfa = false
  }


  if len(project_config.Project) <= 0 {
    log.Fatal("[ERROR] No project is set in your project.yaml configuration.")
  }

  accounts := map[string] bool {fmt.Sprintf("%s-dev", project_config.Project): true, 
                            fmt.Sprintf("%s-prd", project_config.Project): true,
                           }

  state_config := &tf_helper.Config{ Bucket_name: fmt.Sprintf("%s-%s-%s-tfstate", project_config.Project, project_config.account, project_config.environment),
                                     State_filename: fmt.Sprintf("%s-%s-%s.tfstate", project_config.Project, project_config.account, project_config.environment),
                                     Versioning: true,
                                   }


  if _, ok := accounts[project_config.account]; !ok {
    log.Fatalf("[ERROR] Account directories do not match project name. Name found: %s, expected %s-dev or %s-prd\n", project_config.account, project_config.Project, project_config.Project)
  }


  awsconf := &aws_helper.Config{ Region: project_config.Region, 
                                 Profile: project_config.Root_profile, 
                                 Role: project_config.Roam_role, 
                                 Account_id: project_config.Accounts_mapping[project_config.account],
                                 Use_mfa: use_mfa,
                               }

  client := awsconf.Connect()


  bucket_created := false

  for i:=1; i<=retries; i++ {

    if !state_config.Create_bucket(client) {
      log.Printf("[WARN] S3 Bucket %s failed to be created. Retrying.\n", state_config.Bucket_name)
    } else {
      log.Printf("[INFO] S3 Bucket %s created and versioning enabled.\n", state_config.Bucket_name)
      bucket_created = true
      break
    }

    time.Sleep(time.Duration(i)*time.Second)

  }

  if bucket_created {
    state_config.Setup_remote_state(client)   
  } else {
    log.Fatalf("[ERROR] S3 Bucket failed to be created after %d retries. Aborting.\n", retries)
  }



  
  
}