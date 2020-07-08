package main

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"awsutils/pkg/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("uploads3 file bucket:file")
	}

	src := os.Args[1]

	pieces := strings.Split(os.Args[2], ":")
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
