package aws

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSService struct {
	Client *sns.Client
}

func (s *SNSService) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.Client.ListTopics(ctx, &sns.ListTopicsInput{})
	return err
}

func (s *SNSService) Publish(msg, topic string) error {
	topicPrefix := os.Getenv("SNS_TOPIC_PREFIX")
	if topicPrefix == "" {
		return errors.New("no topic prefix")
	}
	topicSufix := os.Getenv("SNS_TOPIC_SUFIX")
	topicArn := fmt.Sprintf("%s:%s_%s", topicPrefix, topic, topicSufix)

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
