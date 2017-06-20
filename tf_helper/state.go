package tf_helper

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mhlias/tholos/aws_helper"
)

type Config struct {
	Bucket_name      string
	State_filename   string
	Lock_table       string
	Encrypt_s3_state bool
	Versioning       bool
	TargetsTF        []string
	TFlegacy         bool
	TFenv            string
	Region           string
}

func (c *Config) Create_bucket(client interface{}) bool {

	resp, err := client.(*aws_helper.AWSClient).S3conn.ListBuckets(&s3.ListBucketsInput{})

	if err != nil {
		log.Println("[ERROR] Failed to check if bucket exists: ", err)
		return false
	}

	for _, b := range resp.Buckets {

		if *b.Name == c.Bucket_name {

			if c.enable_versioning(client) {

			} else {

			}

			return true
		}

	}

	params := &s3.CreateBucketInput{
		Bucket: aws.String(c.Bucket_name),
	}

	_, err2 := client.(*aws_helper.AWSClient).S3conn.CreateBucket(params)

	if err2 != nil {
		log.Fatal("[ERROR] Failed to create bucket with name %s with error: %v\n", c.Bucket_name, err2)
	}

	if c.enable_versioning(client) {
		log.Printf("[INFO] Versioning was enabled on bucket %s.\n", c.Bucket_name)
	} else {
		log.Fatal("[ERROR] Versioning failed to be enabled in remote state S3 bucket.")
	}

	return true

}

func (c *Config) enable_versioning(client interface{}) bool {

	params := &s3.GetBucketVersioningInput{
		Bucket: aws.String(c.Bucket_name),
	}

	resp, err := client.(*aws_helper.AWSClient).S3conn.GetBucketVersioning(params)

	if err != nil {
		log.Println("[ERROR] Failed to enabled versioning in the remote state S3 bucket: ", err)

	}

	if resp.Status != nil && *resp.Status == "Enabled" {
		return true
	} else {

		params2 := &s3.PutBucketVersioningInput{
			Bucket: aws.String(c.Bucket_name), // Required
			VersioningConfiguration: &s3.VersioningConfiguration{ // Required
				Status: aws.String("Enabled"),
			},
		}

		_, err2 := client.(*aws_helper.AWSClient).S3conn.PutBucketVersioning(params2)

		if err2 != nil {
			log.Println("[ERROR] Failed to enable versioning on S3 bucket %s: ", c.Bucket_name, err)
			return false
		}

	}

	return true

}

func (c *Config) Create_locktable(client interface{}) bool {

	params := &dynamodb.ListTablesInput{}

	resp, err := client.(*aws_helper.AWSClient).Dynconn.ListTables(params)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	for _, dt := range resp.TableNames {
		if *dt == c.Lock_table {
			return true
		}
	}

	params2 := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("LockID"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("LockID"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(c.Lock_table),
	}
	_, err2 := client.(*aws_helper.AWSClient).Dynconn.CreateTable(params2)

	if err2 != nil {
		fmt.Println(err2.Error())
		return false
	}

	return true

}

func (c *Config) Switch_env() {

	var args []string
	env_exists := false

	cmdList := exec.Command("terraform", "env", "list")
	var out bytes.Buffer
	cmdList.Stdout = &out
	err := cmdList.Run()
	if err != nil {
		log.Fatal("Failed to get Terraform state environments list:", err)
	}

	out_str := out.String()

	tfenvs := strings.Split(out_str, "\n")

	for _, e := range tfenvs {
		if c.TFenv == strings.Trim(e, "* ") {
			env_exists = true
			break
		}
	}

	if !env_exists {

		cmdCreate := "terraform"

		args = []string{
			"env",
			"new",
			c.TFenv,
		}

		if ExecCmd(cmdCreate, args) {
			log.Printf("[INFO] Terraform state environment %s created.", c.TFenv)
		} else {
			log.Fatal("[ERROR] Failed create Terraform state environment. Aborting.\n")
		}

	}

	cmdSelect := "terraform"

	args = []string{
		"env",
		"select",
		c.TFenv,
	}

	if ExecCmd(cmdSelect, args) {
		log.Printf("[INFO] Terraform state environment %s selected.", c.TFenv)
	} else {
		log.Fatal("[ERROR] Failed select Terraform state environment. Aborting.\n")
	}

}

func (c *Config) Setup_remote_state() {

	//log.Printf("[INFO] Environment variables: %s, %s, %s, %s", os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), os.Getenv("AWS_SECURITY_TOKEN"), os.Getenv("AWS_DEFAULT_REGION") )

	cmdName := "terraform"

	var args []string

	if c.TFlegacy {

		args = []string{"remote",
			"config",
			"-backend=S3",
			fmt.Sprintf("-backend-config=bucket=%s", c.Bucket_name),
			fmt.Sprintf("-backend-config=key=%s", c.State_filename),
			fmt.Sprintf("-backend-config=encrypt=%t", c.Encrypt_s3_state),
		}

	} else {

		args = []string{"init",
			"-backend=true",
			fmt.Sprintf("-backend-config=bucket=%s", c.Bucket_name),
			fmt.Sprintf("-backend-config=key=%s", c.State_filename),
			fmt.Sprintf("-backend-config=region=%s", c.Region),
			fmt.Sprintf("-backend-config=lock_table=%s", c.Lock_table),
			fmt.Sprintf("-backend-config=encrypt=%t", c.Encrypt_s3_state),
			"-force-copy",
		}

	}

	if ExecCmd(cmdName, args) {
		log.Println("[INFO] Remote State was set up successfully.")
	} else {
		log.Fatal("[ERROR] Remote state failed to be set up. Aborting.\n")
	}

}
