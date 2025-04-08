package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"sqsworkers/sqsworker"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type handler struct{}

var _ sqsworker.Handler = (*handler)(nil)

func (h *handler) Handle(ctx context.Context, message types.Message) error {
	if message.Body == nil {
		slog.Error("invalid sqs message", "message_id", *message.MessageId, "body", *message.Body)
		return nil
	}
	var data struct {
		DurationSeconds int `json:"duration_seconds"`
	}
	if err := json.Unmarshal([]byte(*message.Body), &data); err != nil {
		slog.Error("invalid sqs message payload", "message_id", *message.MessageId, "body", *message.Body, "error", err)
		return nil
	}
	slog.Info("received message", "message_id", *message.MessageId, "body", *message.Body, "attributes", message.Attributes)
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("sqs handler context done: %w", ctx.Err())
		case t := <-ticker.C:
			if int(t.Sub(start).Seconds()) > data.DurationSeconds {
				slog.Info("completed message processing", "message_id", *message.MessageId)
				return nil
			}
			slog.Info("still processing message", "message_id", *message.MessageId)
		}
	}
}

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	queueName := "main"

	client := sqs.NewFromConfig(cfg)

	sqsWorkerConfig := &sqsworker.Config{
		SqsQueueName:             queueName,
		Concurrency:              10,
		VisibilityTimeoutSeconds: 300,
		WaitTimeSeconds:          10,
		TimeoutSeconds:           60,
	}

	worker := sqsworker.NewSqsWorkers(sqsWorkerConfig, client, &handler{})
	if err := worker.Start(ctx); err != nil {
		log.Fatal("failed to run sqs worker: %w", err)
	}
}
