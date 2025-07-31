package aws

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type Services struct {
	SNS SNSService
}

func NewServices(cfg aws.Config) *Services {
	var client *sns.Client
	if snsEndpoint := os.Getenv("SNS_ENDPOINT"); snsEndpoint != "" {
		client = sns.NewFromConfig(cfg, func(o *sns.Options) {
			o.BaseEndpoint = aws.String(snsEndpoint)
		})
	} else {
		client = sns.NewFromConfig(cfg)
	}

	return &Services{
		SNS: SNSService{Client: client},
	}
}
