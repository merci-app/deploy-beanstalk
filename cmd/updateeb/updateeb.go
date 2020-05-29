package main

import (
	"flag"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
)

func Env(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("[Requirements] Env %s is empty", key)
	}

	return value
}

func DurationFromEnv(key string, defaultValue time.Duration) time.Duration {
	if raw := os.Getenv(key); raw != "" {
		raw = strings.ToLower(strings.TrimSpace(raw))

		format := time.Second
		switch raw[len(raw)-1] {
		case 'm':
			format = time.Minute
			fallthrough
		case 's':
			raw = raw[:len(raw)-1]
		}

		n, _ := strconv.ParseInt(raw, 10, 64)
		if n <= 0 {
			log.Fatalf("%s must be greater than 0", key)
		}

		return format * time.Duration(n)
	}

	return defaultValue
}

func panicIfEmpty(label string, val *string) {
	if *val == "" {
		log.Fatalf("argument \"%s\" must not be empty", label)
	}
}

func main() {
	region := Env("AWS_REGION")
	accessKey := Env("AWS_ACCESS_KEY")
	secretKey := Env("AWS_SECRET_KEY")

	application := flag.String("application", "", "application name")
	environment := flag.String("environment", "", "environment name")
	version := flag.String("version", "", "unique version name")
	src := flag.String("src", "", "s3 src; -src=bucket:path/file")

	flag.Parse()

	pieces := strings.Split(*src, ":")
	if len(pieces) != 2 {
		log.Fatal("invalida -src; must be; -src=bucket:file")
	}

	panicIfEmpty("application", application)
	panicIfEmpty("environment", environment)
	panicIfEmpty("version", version)

	bucket := pieces[0]
	bucketKey := pieces[1]

	checkStatusTimeout := DurationFromEnv("AWS_CHECK_STATUS_TIMEOUT", time.Minute*15)
	checkInterval := DurationFromEnv("AWS_CHECK_STATUS_INTERVAL", time.Second*5)
	degradedTimeout := DurationFromEnv("AWS_DEGRADED_STATUS_TIMEOUT", time.Minute*15)

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("[Session] err; %v\n", err)
	}

	conf := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}
	client := elasticbeanstalk.New(sess, conf)

	{
		ap, err := client.CreateApplicationVersion(
			&elasticbeanstalk.CreateApplicationVersionInput{
				VersionLabel:          aws.String(*version),
				ApplicationName:       aws.String(*application),
				AutoCreateApplication: aws.Bool(false),
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
				VersionLabel:    aws.String(*version),
				ApplicationName: aws.String(*application),
				EnvironmentName: aws.String(*environment),
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
			EnvironmentName: aws.String(*environment),
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
