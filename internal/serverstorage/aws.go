package serverstorage

// export AWS_ACCESS_KEY_ID
// export AWS_SECRET_ACCESS_KEY

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	awsbucketName = "item-keeper-demo"
	awsregion     = "eu-central-1"
	s3client      *s3.S3
)

func createSession() error {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsregion),
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		fmt.Println("Failed to create session", err)
		return err
	}

	// create client
	s3client = s3.New(sess)
	return nil
}

func uploadFileS3(file *File) error {

	// Upload the data to S3
	_, err := s3client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(awsbucketName),
		Key:    aws.String(strconv.FormatInt(file.FileID, 10)),
		Body:   bytes.NewReader(file.Body),
	})
	if err != nil {
		fmt.Println("Failed to upload data", err)
		return err
	}
	return nil
}

func downloadS3(file *File) ([]byte, error) {

	// Download file from S3
	result, err := s3client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(awsbucketName),
		Key:    aws.String(strconv.FormatInt(file.FileID, 10)),
	})
	if err != nil {
		fmt.Println("Failed to download file", err)
		return nil, err
	}
	defer result.Body.Close()

	// Read the file content into a byte slice
	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		fmt.Println("Failed to read file content", err)
		return nil, err
	}
	return data, err
}

func deleteS3(files []File) error {

	// create a list of objects to be deleted
	objects := make([]*s3.ObjectIdentifier, 0)
	for _, file := range files {
		objects = append(objects, &s3.ObjectIdentifier{
			Key: aws.String(strconv.FormatInt(file.FileID, 10)),
		})
	}

	// input for the DeleteObjects API
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(awsbucketName),
		Delete: &s3.Delete{
			Objects: objects,
			Quiet:   aws.Bool(true), // Set to true to suppress the response
		},
	}

	// Delete the objects from the S3 bucket
	_, err := s3client.DeleteObjects(input)
	if err != nil {
		fmt.Println("Failed to delete objects from S3", err)
		return err
	}
	return nil
}
