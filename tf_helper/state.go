package tf_helper


import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mhlias/tholos/aws_helper"
)



type config struct {
  bucket_name string
  state_filename string
  versioning bool
}


func (c *config) Create_bucket(client interface{}) bool {

	params := new(s3.ListBucketsInput)
	resp, err := client.(AWSClient).s3conn.ListBuckets(params)

	if err != nil {
		log.Println("[ERROR] Failed to check if bucket exists: ", err)
		return false
	}


	for _, b := range resp.Buckets {

		if b.Name == c.bucket_name {

			 if c.enable_versioning() {

			 	} else {

			 	}

			return true
		}

	}

	params := &s3.CreateBucketInput{
	    Bucket: aws.String(c.bucket_name),
	}
	
	resp, err := client.(AWSClient).s3conn.CreateBucket(params)


	if c.enable_versioning() {

 	} else {
 		
 	}



	return true

}


func (c *config) enable_versioning(client interface{}) bool {

	params := &s3.GetBucketVersioningInput{
		Bucket: aws.String(c.bucket_name),
	}

	resp, err := client.(AWSClient).s3conn.GetBucketVersioning(params)

	if err != nil {

	}

	if *resp.VersioningConfiguration {
		return true
	} else {

		params2 := &s3.PutBucketVersioningInput{
		    Bucket: aws.String(c.bucket_name), // Required
		    VersioningConfiguration: &s3.VersioningConfiguration{ // Required
		        MFADelete: aws.String("MFADelete"),
		        Status:    aws.String("BucketVersioningStatus"),
		    },
		}

		resp2, err2 := client.(AWSClient).s3conn.PutBucketVersioning(params2)

		if err2 != nil {
		    log.Println("[ERROR] Failed to enable versioning on S3 bucket %s: ", c.bucket_name, err)
		    return false
		}

	}

	return true

}


func (c *config) setup_remote_state(client interface{}) {

	os.Exec()

}