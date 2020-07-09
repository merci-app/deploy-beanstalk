package config

import (
	"time"

	"awsutils/pkg/env"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

func New() *aws.Config {
	return &aws.Config{
		Retryer: client.DefaultRetryer{
			NumMaxRetries: env.Int("MAX_RETRIES", 20),
			MinRetryDelay: env.Duration("MIN_RETRY_DELAY", time.Second),
			MaxRetryDelay: env.Duration("MAX_RETRY_DELAY", time.Minute),
		},
		Region: aws.String(env.String("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			env.String("AWS_ACCESS_KEY"),
			env.String("AWS_SECRET_KEY"),
			"",
		),
	}
}
