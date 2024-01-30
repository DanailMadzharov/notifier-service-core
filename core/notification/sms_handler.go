package notification

import (
	"context"
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"net/http"
	"regexp"
	. "sumup-notifier-service/core/errors"
	"sumup-notifier-service/core/kafkamessaging"
)

type SMSHandler struct {
	kafkaWriter *kafka.Writer
}

type SMSNotification struct {
	TelephoneNumber string `json:"telephoneNumber" validate:"required,len=13"`
	Message         string `json:"message" validate:"required,lt=200"`
}

func NewSMSHandler(kafkaWriter *kafka.Writer) *SMSHandler {
	return &SMSHandler{
		kafkaWriter: kafkaWriter,
	}
}

func (s *SMSHandler) ParseData(rawData *json.RawMessage) (interface{}, *Error) {
	var smsData SMSNotification
	err := json.Unmarshal(*rawData, &smsData)
	if err != nil {
		return nil, &Error{
			StatusCode: http.StatusInternalServerError, ErrorCode: "501",
			Message: "An error occurred while raw data serialization",
		}
	}

	return smsData, nil
}
func (s *SMSHandler) ValidateExpectedFields(parsedData interface{}) *Error {
	data, handlerError := s.getSmsNotification(parsedData)
	if handlerError != nil {
		return handlerError
	}

	var fieldErrors *[]FieldError
	fieldErrors = &[]FieldError{}

	s.validateRequest(&data, fieldErrors)
	validateTelephoneNumber(&data.TelephoneNumber, fieldErrors)

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

func (s *SMSHandler) SendNotification(parsedData interface{}) *Error {
	data, handlerError := s.getSmsNotification(parsedData)
	if handlerError != nil {
		return handlerError
	}

	jsonData, _ := json.Marshal(data)
	kafkamessaging.Produce(context.Background(), &jsonData, s.kafkaWriter)

	return nil
}

func (s *SMSHandler) validateRequest(request *SMSNotification, fieldErrors *[]FieldError) {
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

func validateTelephoneNumber(number *string, fieldErrors *[]FieldError) {
	bulgarianNumberPattern := `^\+359\d{9}$`
	if matched, _ := regexp.MatchString(bulgarianNumberPattern, *number); !matched {
		*fieldErrors = append(*fieldErrors, FieldError{FieldName: "telephoneNumber",
			Message: "The telephone number is not valid, ex: +359111111111"})
	}
}

func (s *SMSHandler) getSmsNotification(parsedData interface{}) (SMSNotification, *Error) {
	data, ok := parsedData.(SMSNotification)
	if !ok {
		log.Error().Msg("parsedData is not of type email, aborting...")
		return SMSNotification{}, &Error{
			StatusCode: http.StatusInternalServerError,
			ErrorCode:  "501",
			Message:    "The specified object is not of type Sms",
		}
	}

	return data, nil
}
