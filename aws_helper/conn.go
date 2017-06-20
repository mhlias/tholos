package aws_helper

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
)

type AWSClient struct {
	stsconn *sts.STS
	S3conn  *s3.S3
	Dynconn *dynamodb.DynamoDB
	region  string
}

type Config struct {
	Region        string
	Profile       string
	Role          string
	Account_id    string
	Use_mfa       bool
	Use_sts       bool
	Mfa_device_id string
	Mfa_token     string
}

func (c *Config) Connect() interface{} {

	var client AWSClient

	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_SECURITY_TOKEN")
	os.Unsetenv("AWS_SESSION_TOKEN")
	os.Unsetenv("AWS_DEFAULT_REGION")

	screds := &credentials.SharedCredentialsProvider{Profile: c.Profile}

	log.Printf("[INFO] Using aws shared credentials profile: %s\n", c.Profile)

	awsConfig := &aws.Config{
		Credentials: credentials.NewCredentials(screds),
		Region:      aws.String(c.Region),
		MaxRetries:  aws.Int(3),
	}

	sess := session.New(awsConfig)

	if c.Use_sts {

		log.Println("[INFO] Initializing STS Connection")
		client.stsconn = sts.New(sess)

		params := &sts.AssumeRoleInput{}

		if c.Use_mfa {

			params = &sts.AssumeRoleInput{
				RoleArn:         aws.String(fmt.Sprintf("arn:aws:iam::%s:role/%s", c.Account_id, c.Role)),
				RoleSessionName: aws.String(fmt.Sprintf("%s-%s", c.Account_id, c.Role)),
				DurationSeconds: aws.Int64(3600),
				SerialNumber:    aws.String(c.Mfa_device_id),
				TokenCode:       aws.String(c.Mfa_token),
			}

		} else {

			params = &sts.AssumeRoleInput{
				RoleArn:         aws.String(fmt.Sprintf("arn:aws:iam::%s:role/%s", c.Account_id, c.Role)),
				RoleSessionName: aws.String(fmt.Sprintf("%s-%s", c.Account_id, c.Role)),
				DurationSeconds: aws.Int64(3600),
			}

		}

		sts_resp, sts_err := client.stsconn.AssumeRole(params)

		if sts_err != nil {
			log.Fatalf("Unable to assume role: %v", sts_err.Error())
		}

		os.Setenv("AWS_ACCESS_KEY_ID", *sts_resp.Credentials.AccessKeyId)
		os.Setenv("AWS_SECRET_ACCESS_KEY", *sts_resp.Credentials.SecretAccessKey)
		os.Setenv("AWS_SECURITY_TOKEN", *sts_resp.Credentials.SessionToken)
		os.Setenv("AWS_SESSION_TOKEN", *sts_resp.Credentials.SessionToken)
		os.Setenv("AWS_DEFAULT_REGION", c.Region)

		return c.assumeConnect(sts_resp)

	} else {

		profile_creds := credentials.Value{}
		var profile_err error

		if profile_creds, profile_err = screds.Retrieve(); profile_err != nil {
			log.Fatalf("[ERROR] Failed to get aws credentials for profile: %s with error: %s\n", c.Profile, profile_err.Error())
		}

		os.Setenv("AWS_ACCESS_KEY_ID", profile_creds.AccessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", profile_creds.SecretAccessKey)
		os.Setenv("AWS_DEFAULT_REGION", c.Region)
	}

	log.Println("[INFO] Initializing S3 Connection")
	client.S3conn = s3.New(sess)
	client.Dynconn = dynamodb.New(sess)

	return &client

}

func (c *Config) assumeConnect(sts *sts.AssumeRoleOutput) interface{} {

	var client AWSClient

	awsConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials(*sts.Credentials.AccessKeyId, *sts.Credentials.SecretAccessKey, *sts.Credentials.SessionToken),
		Region:      aws.String(c.Region),
		MaxRetries:  aws.Int(3),
	}

	sess := session.New(awsConfig)

	log.Println("[INFO] Initializing S3 Connection")
	client.S3conn = s3.New(sess)
	client.Dynconn = dynamodb.New(sess)

	return &client

}
