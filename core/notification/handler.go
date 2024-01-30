package notification

import (
	"encoding/json"
	"sumup-notifier-service/core/errors"
)

type INotificationHandler interface {
	ParseData(rawData *json.RawMessage) (interface{}, *errors.Error)
	ValidateExpectedFields(parsedData interface{}) *errors.Error
	SendNotification(parsedData interface{}) *errors.Error
}

type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (service *NotificationService) HandleNotification(handler INotificationHandler, data *json.RawMessage) *errors.Error {
	parsedData, httpError := handler.ParseData(data)
	if httpError != nil {
		return httpError
	}
	httpError = handler.ValidateExpectedFields(parsedData)
	if httpError != nil {
		return httpError
	}
	httpError = handler.SendNotification(parsedData)
	if httpError != nil {
		return httpError
	}

	return nil
}
