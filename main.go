package main

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/aws/aws-sdk-go/service/s3"
)

func Env(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("[Requirements] Env %s is empty", key)
	}

	return value
}

func main() {
	application := Env("AWS_APPLICATION")
	environment := Env("AWS_ENVIRONMENT")
	region := Env("AWS_REGION")
	accessKey := Env("AWS_ACCESS_KEY")
	secretKey := Env("AWS_SECRET_KEY")
	bucket := Env("AWS_BUCKET")
	bucketKey := Env("AWS_BUCKET_KEY")
	version := Env("AWS_VERSION")
	autoCreate := os.Getenv("AWS_AUTO_CREATE") == "true"
	upload := os.Getenv("AWS_UPLOAD") == "true"

	maxChecks := 40
	if raw := os.Getenv("AWS_MAX_CHECKS"); raw != "" {
		n, _ := strconv.ParseInt(raw, 10, 64)
		if n <= 0 {
			log.Fatal("AWS_MAX_CHECKS must be greater than 0")
		}

		maxChecks = int(n)
	}

	checkInterval := time.Second * 2
	if raw := os.Getenv("AWS_CHECK_STATUS_INTERVAL"); raw != "" {
		n, _ := strconv.ParseInt(raw, 10, 64)
		if n <= 0 {
			log.Fatal("AWS_CHECK_STATUS_INTERVAL must be greater than 0")
		}

		checkInterval = time.Second * time.Duration(n)
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("[Session] err; %v\n", err)
	}

	conf := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}
	client := elasticbeanstalk.New(sess, conf)

	// Upload
	if upload {
		body, err := os.Open(bucketKey)
		if err != nil {
			log.Fatalf("[Upload] Bucket Key err; %v\n", err)
		}
		defer body.Close()

		client := s3.New(sess, conf)

		resp, err := client.PutObject(&s3.PutObjectInput{
			Body:        body,
			Bucket:      aws.String(bucket),
			Key:         aws.String(bucketKey),
			ContentType: aws.String(contentType(bucketKey)),
		})
		if err != nil {
			log.Fatalf("[Upload] err; %v\n", err)
		}

		log.Println("[Upload]", resp)
	}

	{
		ap, err := client.CreateApplicationVersion(
			&elasticbeanstalk.CreateApplicationVersionInput{
				VersionLabel:          aws.String(version),
				ApplicationName:       aws.String(application),
				AutoCreateApplication: aws.Bool(autoCreate),
				SourceBundle: &elasticbeanstalk.S3Location{
					S3Bucket: aws.String(bucket),
					S3Key:    aws.String(bucketKey),
				},
			},
		)
		if err != nil {
			log.Fatalf("[Create] err; %v\n", err)
		}

		log.Println("[Create]", ap)
	}

	{
		up, err := client.UpdateEnvironment(
			&elasticbeanstalk.UpdateEnvironmentInput{
				VersionLabel:    aws.String(version),
				ApplicationName: aws.String(application),
				EnvironmentName: aws.String(environment),
			},
		)
		if err != nil {
			log.Fatalf("[Update] err; %v\n", err)
		}

		log.Println("[Update]", up)
		time.Sleep(checkInterval)
	}

	// check status
	for i := 0; i < maxChecks; i++ {
		out, err := client.DescribeEnvironmentHealth(&elasticbeanstalk.DescribeEnvironmentHealthInput{
			AttributeNames:  []*string{aws.String(elasticbeanstalk.EnvironmentHealthAttributeAll)},
			EnvironmentName: aws.String(environment),
		})
		if err != nil {
			log.Fatalf("[Status] err; %v\n", err)
		}

		log.Printf("[Status] %s/%s\n", *out.Status, *out.HealthStatus)
		if out.Status != nil && out.HealthStatus != nil && *out.Status == elasticbeanstalk.EnvironmentStatusReady {
			if *out.HealthStatus == elasticbeanstalk.EnvironmentHealthStatusOk {
				os.Exit(0)
				break
			} else if *out.HealthStatus == elasticbeanstalk.EnvironmentHealthStatusDegraded {
				os.Exit(2)
				break
			}
		}

		time.Sleep(checkInterval)
	}

	log.Fatalf("[Status] Timeout after %d seconds", maxChecks*int(checkInterval/time.Second))
}

func contentType(path string) string {
	ext := filepath.Ext(path)
	typ := mime.TypeByExtension(ext)

	if typ == "" {
		typ = "application/octet-stream"
	}

	return typ
}
