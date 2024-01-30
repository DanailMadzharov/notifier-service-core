package kafkamessaging

import (
	"context"
	"github.com/segmentio/kafka-go"
	"net/http"
	"sumup-notifier-service/core/errors"
)

func Produce(ctx context.Context, message *[]byte, writer *kafka.Writer) *errors.Error {
	err := writer.WriteMessages(ctx, kafka.Message{
		Value: *message,
	})
	if err != nil {
		return &errors.Error{
			StatusCode: http.StatusInternalServerError,
			Message:    "The message could not be sent",
			ErrorCode:  "515",
		}
	}

	return nil
}
