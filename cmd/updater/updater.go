package main

import (
	"os"

	"github.com/alexlast/ecr-credential-updater/internal/kube"
	"github.com/alexlast/ecr-credential-updater/internal/updater"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyLevel: "level",
			log.FieldKeyMsg:   "message",
		},
	})
}

func main() {
	// Load config
	config := new(updater.Config)
	err := envconfig.Process("ecr", config)

	if err != nil {
		log.Fatalln(err)
	}

	// Create a new AWS session
	session := session.New(
		&aws.Config{
			Region: aws.String(config.RegistryRegion),
		},
	)

	// Build kubernetes client
	client, err := kube.BuildClient()

	if err != nil {
		log.Fatalln(err)
	}

	// Build new context
	c := &updater.Context{
		Config:     config,
		Kubernetes: client,
		ECR:        ecr.New(session),
	}

	c.Start()
}
