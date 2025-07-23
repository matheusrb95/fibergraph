package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSService struct {
	Client *sns.Client
}

func (s *SNSService) Publish(msg, topicArn string) error {
	input := &sns.PublishInput{
		Message:  aws.String(msg),
		TopicArn: aws.String(topicArn),
	}

	_, err := s.Client.Publish(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("publish sns message. %w", err)
	}

	return nil
}
