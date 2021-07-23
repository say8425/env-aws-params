package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
)

type SSMClient struct {
	client *ssm.SSM
}

func NewSSMClient(region string, profile string) (*SSMClient, error) {
	var config *aws.Config

	awsSession := session.Must(session.NewSession(
		&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewSharedCredentials("", profile),
		}))
	_, err := awsSession.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}
	config = nil

	endpoint := os.Getenv("SSM_ENDPOINT")
	if endpoint != "" {
		config = &aws.Config{
			Endpoint: &endpoint,
		}
	}

	client := ssm.New(awsSession, config)
	return &SSMClient{client}, nil
}

func (c *SSMClient) GetParametersByPath(path string) (map[string]string, error) {
	if strings.HasSuffix(path, "/") != true {
		path = fmt.Sprintf("%s/", path)
	}

	var nextToken *string
	parameters := make(map[string]string)

	for {
		params := &ssm.GetParametersByPathInput{
			Path:           aws.String(path),
			Recursive:      aws.Bool(true),
			WithDecryption: aws.Bool(true),
			MaxResults:     aws.Int64(10),
			NextToken:      nextToken,
		}
		response, err := c.client.GetParametersByPath(params)

		if err != nil {
			awsErr, _ := err.(awserr.Error)
			log.Errorf("Error Getting Parameters from SSM: %s", awsErr.Code())
			return nil, err
		}

		for _, p := range response.Parameters {
			parameters[strings.TrimPrefix(*p.Name, path)] = *p.Value
		}

		if response.NextToken == nil {
			break
		}
		nextToken = response.NextToken
	}
	return parameters, nil
}
