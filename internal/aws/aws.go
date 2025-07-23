package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type Services struct {
	SNS SNSService
}

func NewServices(cfg aws.Config) *Services {
	client := sns.NewFromConfig(cfg, func(o *sns.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})

	return &Services{
		SNS: SNSService{Client: client},
	}
}
