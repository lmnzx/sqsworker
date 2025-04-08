package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	queueName := "main"
	client := sqs.NewFromConfig(cfg)

	queue, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(`{"duration_seconds": 15}`),
		QueueUrl:    queue.QueueUrl,
	})
	if err != nil {
		log.Fatal(err)
	}
}
