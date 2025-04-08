package sqsworker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go/aws"
)

type Config struct {
	SqsQueueName             string
	Concurrency              int
	VisibilityTimeoutSeconds int
	WaitTimeSeconds          int
	TimeoutSeconds           int
}

type Handler interface {
	Handle(ctx context.Context, message types.Message) error
}

type SqsWorker struct {
	config    *Config
	handler   Handler
	messages  chan types.Message
	sqsClient *sqs.Client
}

func NewSqsWorkers(config *Config, sqsClient *sqs.Client, handler Handler) *SqsWorker {
	return &SqsWorker{
		config:    config,
		handler:   handler,
		messages:  make(chan types.Message),
		sqsClient: sqsClient,
	}
}

func (s *SqsWorker) Start(ctx context.Context) error {
	slog.Info("running sqs worker", "queue_name", s.config.SqsQueueName)

	output, err := s.sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(s.config.SqsQueueName),
	})
	if err != nil {
		return fmt.Errorf("failed to get queue url for %s: %w", s.config.SqsQueueName, err)
	}

	for i := range s.config.Concurrency {
		go func(goruntineId int) {
			if err := s.consumeMessages(ctx, goruntineId, output.QueueUrl); err != nil {
				slog.Error("failed to process messages", "error", err)
			}
		}(i)
	}

	for {
		receiveOutput, err := s.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:                    output.QueueUrl,
			VisibilityTimeout:           int32(s.config.VisibilityTimeoutSeconds),
			WaitTimeSeconds:             int32(s.config.WaitTimeSeconds),
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{"ApproximateReceiveCount"},
		})
		if err != nil {
			slog.Error("failed to receive messages form sqs", "queue_name", s.config.SqsQueueName, "queur_url", *output.QueueUrl, "err", err)
			continue
		}
		for _, message := range receiveOutput.Messages {
			s.messages <- message
		}
	}
}

func (s *SqsWorker) consumeMessages(ctx context.Context, goroutineId int, queueUrl *string) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("done consuming messages for goroutine %d: %w", goroutineId, ctx.Err())
		case message := <-s.messages:
			slog.Info("received message", "goroutine_id", goroutineId, "message_id", *message.MessageId)
			s.processMessage(ctx, goroutineId, message, queueUrl)
		}
	}
}

func (s *SqsWorker) processMessage(ctx context.Context, goroutineId int, message types.Message, queurUrl *string) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("failed to process sqs message due to panic", "goroutine_id", goroutineId, "message_id", *message.MessageId, "err", r)
		}
	}()

	handleCtx, handleCancel := context.WithTimeout(ctx, time.Duration(s.config.TimeoutSeconds)*time.Second)
	defer handleCancel()
	if err := s.handler.Handle(handleCtx, message); err != nil {
		slog.Error("failed to process sqs message", "goroutine_id", goroutineId, "message_id", *message.MessageId, "err", err)
	} else {
		if _, err := s.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      queurUrl,
			ReceiptHandle: message.ReceiptHandle,
		}); err != nil {
			slog.Error("failed to delete sqs message", "goroutine_id", goroutineId, "message_id", *message.MessageId, "err", err)
		}
	}
}
