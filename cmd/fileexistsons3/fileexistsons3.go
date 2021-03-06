package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"awsutils/pkg/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("fileexistsons3 bucket:file")
	}

	pieces := strings.Split(os.Args[1], ":")
	if len(pieces) != 2 {
		log.Fatal("please provide bucket:path")
	}

	bucket := pieces[0]
	dest := pieces[1]

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("[Session] err; %v\n", err)
	}

	conf := config.New()

	resp, err := s3.New(sess, conf).HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dest),
	})

	if err != nil {
		if awsErr, is := err.(awserr.RequestFailure); is {
			if awsErr.StatusCode() == http.StatusNotFound {
				log.Println("[File Exists on S3] err;", err)
				os.Exit(4)
			}
		}

		log.Fatalf("[File Exists on S3] Could check file; %v\n", err)
	}

	_, _ = os.Stdout.WriteString(resp.String())
}
