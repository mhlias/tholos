package aws_helper


import (
    "log"
    "fmt"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/service/sts"
    "github.com/aws/aws-sdk-go/service/s3"


)


type AWSClient struct {
  stsconn            *sts.STS
  s3conn             *s3.S3
  region             string
}


type Config struct {
  Region 		 string
  Profile 	 string
  Role 			 string
  Account_id string
  Use_mfa		 bool
}




func (c *Config) Connect() interface{} {

  var client, project_client AWSClient

  screds := &credentials.SharedCredentialsProvider{Profile: c.Profile}

  awsConfig := &aws.Config{
    Credentials: credentials.NewCredentials(screds),
    Region:      aws.String(c.Region),
    MaxRetries:  aws.Int(3),
  }


  sess := session.New(awsConfig)

  log.Println("[INFO] Initializing STS Connection")
  client.stsconn = sts.New(sess)

  params := &sts.AssumeRole{}

  if c.Use_mfa {

  	params = &sts.AssumeRoleInput{
	      RoleArn:         aws.String(fmt.Sprintf("arn:aws:iam::%s:role/%s", c.Account_id, c.Role)),
	      RoleSessionName: aws.String(fmt.Sprintf("%s-%s", c.Account_id, c.Role)),
	      DurationSeconds: aws.Int64(3600),
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

  return c.assumeConnect(sts_resp)

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
  client.s3conn = s3.New(sess)

  return &client

}