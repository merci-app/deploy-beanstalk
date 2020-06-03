package main

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func Env(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("[Requirements] Env %s is empty", key)
	}

	return value
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

	if len(os.Args) != 3 {
		log.Fatal("uploads3 file bucket:file")
	}

	src := os.Args[1]

	pieces := strings.Split(os.Args[2], ":")
	if len(pieces) != 2 {
		log.Fatal("please provide bucket:/path")
	}

	bucket := pieces[0]
	dest := pieces[1]

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("[Session] err; %v\n", err)
	}

	conf := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}

	// Upload
	body, err := os.Open(src)
	if err != nil {
		log.Fatalf("[Upload S3] Could not open source file; %v\n", err)
	}
	defer body.Close()

	resp, err := s3.New(sess, conf).PutObject(&s3.PutObjectInput{
		Body:        body,
		Bucket:      aws.String(bucket),
		Key:         aws.String(dest),
		ContentType: aws.String(contentType(dest)),
	})
	if err != nil {
		log.Fatalf("[Upload S3] Could not upload file; %v\n", err)
	}

	log.Println("[Upload S3]", resp)
}

func contentType(path string) string {
	ext := filepath.Ext(path)
	typ := mime.TypeByExtension(ext)

	if typ == "" {
		typ = "application/octet-stream"
	}

	return typ
}
