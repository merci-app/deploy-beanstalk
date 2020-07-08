package main

import (
	"flag"
	"log"
	"strings"

	"awsutils/pkg/config"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
)

func panicIfEmpty(label string, val *string) {
	if *val == "" {
		log.Fatalf("argument \"%s\" must not be empty", label)
	}
}

func main() {
	rawEnvs := flag.String("envs", "", "list of environment variables")
	application := flag.String("application", "", "application name")
	environment := flag.String("environment", "", "environment name")

	flag.Parse()

	panicIfEmpty("application", application)
	panicIfEmpty("environment", environment)

	envs := strings.Split(*rawEnvs, " ")
	if len(envs) == 0 || *rawEnvs == "" {
		log.Println("[Check-Environment] argument \"envs\" is empty")
		return
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("[Session] err; %v\n", err)
	}

	client := elasticbeanstalk.New(sess, config.New())

	desc, err := client.DescribeConfigurationSettings(&elasticbeanstalk.DescribeConfigurationSettingsInput{
		ApplicationName: application,
		EnvironmentName: environment,
	})
	if err != nil {
		log.Fatalf("[Check-Environment] fail to fetch data; %v\n", err)
	}

	removeEnvs := make(map[string]struct{}, len(envs))
	for _, a := range desc.ConfigurationSettings {
		for _, cfg := range a.OptionSettings {
			if *cfg.Namespace == "aws:elasticbeanstalk:application:environment" {
				removeEnvs[*cfg.OptionName] = struct{}{}
			}
		}
	}

	fail := false
	for _, env := range envs {
		if _, has := removeEnvs[env]; !has {
			log.Printf("[Check-Environment] key \"%s\" not found\n", env)
			fail = true
		}
	}

	if fail {
		log.Fatal("[Check-Environment] Not all environment variables were found.")
		return
	}

	log.Println("[Check-Environment] All environment variables were found.")
}
