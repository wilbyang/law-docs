package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// TODO: read from config
const (
	QueueURL = "http://sqs.eu-west-1.localhost.localstack.cloud:4566/000000000000/my-queue"
)

type Notifier struct {
	SQSClient *sqs.Client
}

func NewNotifier(ctx context.Context, cfg aws.Config) (*Notifier, error) {
	sqsClient := sqs.NewFromConfig(cfg)
	return &Notifier{SQSClient: sqsClient}, nil
}

func (notifier *Notifier) SendMessage(ctx context.Context, message string) error {
	queueURL, err := notifier.getQueueURL(ctx, QueueURL)
	if err != nil {
		return err
	}
	_, err = notifier.SQSClient.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(message),
		QueueUrl:    aws.String(queueURL),
	})
	return err
}

func (notifier *Notifier) getQueueURL(ctx context.Context, queueName string) (string, error) {
	return QueueURL, nil

	result, err := notifier.SQSClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	if err != nil {
		return "", err
	}
	return *result.QueueUrl, nil
}

// ReceiveMessage receives a message from the queue and processes it, it is blocking, use it in a separate goroutine
func (notifier *Notifier) ReceiveMessage(processor func(message string) error) {
	queueURL, err := notifier.getQueueURL(context.TODO(), QueueURL)
	if err != nil {
		slog.Error("Failed to get queue URL, exiting sqs listener", "error", err)
		return
	}
	for {
		msg, err := notifier.SQSClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl: aws.String(queueURL),
		})
		if err != nil {
			slog.Error("Failed to receive message", "error", err)
			continue
		}
		for _, message := range msg.Messages {

			// set visibility timeout to 10 seconds
			_, err := notifier.SQSClient.ChangeMessageVisibility(context.TODO(), &sqs.ChangeMessageVisibilityInput{
				QueueUrl:          aws.String(queueURL),
				ReceiptHandle:     message.ReceiptHandle,
				VisibilityTimeout: 10,
			})
			if err != nil {
				slog.Error("Failed to change message visibility", "error", err)
				continue
			}

			err = processor(*message.Body)
			if err != nil {
				slog.Error("Failed to process message", "error", err)
			}
			// delete message from queue
			_, err = notifier.SQSClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil {
				slog.Error("Failed to delete message", "error", err)
			}
		}

		// sleep for 1 second
		//TODO: read from config
		time.Sleep(1 * time.Second)

	}
}
