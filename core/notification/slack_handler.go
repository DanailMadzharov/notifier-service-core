package notification

import (
	"context"
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
	"net/http"
	. "sumup-notifier-service/core/errors"
	"sumup-notifier-service/core/kafkamessaging"
)

type SlackHandler struct {
	kafkaWriter *kafka.Writer
}

type SlackNotification struct {
	SlackChannel string `json:"slackChannel" validate:"required,lt=200"`
	Message      string `json:"message" validate:"required,lt=200"`
}

func NewSlackHandler(kafkaWriter *kafka.Writer) *SlackHandler {
	return &SlackHandler{
		kafkaWriter: kafkaWriter,
	}
}

func (s *SlackHandler) ParseData(rawData *json.RawMessage) (interface{}, *Error) {
	var slackData SlackNotification
	err := json.Unmarshal(*rawData, &slackData)
	if err != nil {
		return nil, &Error{
			StatusCode: http.StatusInternalServerError, ErrorCode: "501",
			Message: "An error occurred while raw data serialization",
		}
	}

	return slackData, nil
}

func (s *SlackHandler) ValidateExpectedFields(parsedData interface{}) *Error {
	data, ok := parsedData.(SlackNotification)
	if !ok {
		return &Error{
			StatusCode: http.StatusInternalServerError, ErrorCode: "501",
			Message: "An error occurred while raw data serialization",
		}
	}

	var fieldErrors *[]FieldError
	fieldErrors = &[]FieldError{}

	s.validateRequest(&data, fieldErrors)

	if len(*fieldErrors) > 0 {
		return &Error{
			StatusCode:  http.StatusBadRequest,
			ErrorCode:   "401",
			Message:     "Validation Error",
			FieldErrors: *fieldErrors,
		}
	}

	return nil
}

func (s *SlackHandler) SendNotification(parsedData interface{}) *Error {
	data, ok := parsedData.(SlackNotification)
	if !ok {
		return &Error{
			StatusCode: http.StatusInternalServerError, ErrorCode: "501",
			Message: "An error occurred while raw data serialization",
		}
	}

	jsonData, _ := json.Marshal(data)
	kafkamessaging.Produce(context.Background(), &jsonData, s.kafkaWriter)

	return nil
}

func (s *SlackHandler) validateRequest(request *SlackNotification, fieldErrors *[]FieldError) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(*request)

	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			*fieldErrors = append(*fieldErrors, FieldError{
				FieldName: e.Field(),
				Message:   e.Error(),
			})
		}
	}
}
