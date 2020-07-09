package main

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"time"

	"awsutils/pkg/config"
	"awsutils/pkg/env"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	application := env.String("AWS_APPLICATION")
	environment := env.String("AWS_ENVIRONMENT")
	bucket := env.String("AWS_BUCKET")
	bucketKey := env.String("AWS_BUCKET_KEY")
	version := env.String("AWS_VERSION")
	autoCreate := os.Getenv("AWS_AUTO_CREATE") == "true"
	upload := os.Getenv("AWS_UPLOAD") == "true"
	checkStatusTimeout := env.Duration("AWS_CHECK_STATUS_TIMEOUT", time.Minute*15)
	checkInterval := env.Duration("AWS_CHECK_STATUS_INTERVAL", time.Second*5)
	degradedTimeout := env.Duration("AWS_DEGRADED_STATUS_TIMEOUT", time.Minute*15)

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("[Session] err; %v\n", err)
	}

	conf := config.New()

	// Upload
	if upload {
		log.Println("[Upload]", bucketKey)
		body, err := os.Open(bucketKey)
		if err != nil {
			log.Fatalf("[Upload] Bucket Key err; %v\n", err)
		}
		defer body.Close()

		s3Client := s3.New(sess, conf)

		resp, err := s3Client.PutObject(&s3.PutObjectInput{
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

	client := elasticbeanstalk.New(sess, conf)

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
	}

	var degradedStart time.Time

	start := time.Now()

	// check status
	for time.Since(start) < checkStatusTimeout {
		time.Sleep(checkInterval)

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
				if degradedStart.IsZero() {
					degradedStart = time.Now()
				} else if time.Since(degradedStart) >= degradedTimeout {
					log.Fatalf("[Status] Degraded timeout after %v", time.Since(degradedStart))
					os.Exit(2)
				}
			} else {
				degradedStart = time.Time{}
			}
		}
	}

	log.Fatalf("[Status] Timeout after %v", time.Since(start))
}

func contentType(path string) string {
	ext := filepath.Ext(path)
	typ := mime.TypeByExtension(ext)

	if typ == "" {
		typ = "application/octet-stream"
	}

	return typ
}
