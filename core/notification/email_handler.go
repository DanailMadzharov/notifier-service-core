package notification

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"net/http"
	. "sumup-notifier-service/core/errors"
	"sumup-notifier-service/core/kafkamessaging"
)

import (
	"encoding/json"
)

type EmailNotification struct {
	ToEmail   string `json:"toEmail" validate:"required,lt=50"`
	FromEmail string `json:"fromEmail" validate:"required,lt=50"`
	Message   string `json:"message" validate:"required,lt=200"`
	Subject   string `json:"subject" validate:"required,lt=100"`
}

type EmailHandler struct {
	kafkaWriter *kafka.Writer
}

func NewEmailHandler(kafkaWriter *kafka.Writer) *EmailHandler {
	return &EmailHandler{
		kafkaWriter: kafkaWriter,
	}
}

func (e *EmailHandler) ParseData(rawData *json.RawMessage) (interface{}, *Error) {
	var emailData EmailNotification
	err := json.Unmarshal(*rawData, &emailData)
	if err != nil {
		return nil, &Error{
			StatusCode: http.StatusInternalServerError, ErrorCode: "501",
			Message: "An error occurred while raw data serialization",
		}
	}

	return emailData, nil
}

func (e *EmailHandler) ValidateExpectedFields(parsedData interface{}) *Error {
	data, handlerError := e.getEmailNotification(parsedData)
	if handlerError != nil {
		return handlerError
	}

	var fieldErrors *[]FieldError
	fieldErrors = &[]FieldError{}

	e.validateRequest(&data, fieldErrors)

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

func (e *EmailHandler) SendNotification(parsedData interface{}) *Error {
	data, handlerError := e.getEmailNotification(parsedData)
	if handlerError != nil {
		return handlerError
	}

	jsonData, _ := json.Marshal(data)
	kafkamessaging.Produce(context.Background(), &jsonData, e.kafkaWriter)

	return nil
}

func (e *EmailHandler) validateRequest(request *EmailNotification, fieldErrors *[]FieldError) {
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

func (e *EmailHandler) getEmailNotification(parsedData interface{}) (EmailNotification, *Error) {
	data, ok := parsedData.(EmailNotification)
	if !ok {
		log.Error().Msg("parsedData is not of type email, aborting...")
		return EmailNotification{}, &Error{
			StatusCode: http.StatusInternalServerError,
			ErrorCode:  "501",
			Message:    "The specified object is not of type Email",
		}
	}

	return data, nil
}
