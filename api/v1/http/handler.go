package http

import (
	"encoding/json"
	goerrors "errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"sumup-notifier-service/core/errors"
	"sumup-notifier-service/core/notification"
)

type Handler struct {
	NotificationService *notification.NotificationService
	smsHandler          *notification.SMSHandler
	slackHandler        *notification.SlackHandler
	emailHandler        *notification.EmailHandler
}

func NewHandler(notificationService *notification.NotificationService,
	smsHandler *notification.SMSHandler,
	slackHandler *notification.SlackHandler,
	emailHandler *notification.EmailHandler) *Handler {
	return &Handler{
		NotificationService: notificationService,
		smsHandler:          smsHandler,
		slackHandler:        slackHandler,
		emailHandler:        emailHandler,
	}
}

func (h *Handler) handleNotificationCreation(writer http.ResponseWriter, request *http.Request) {
	rawData, err := h.extractRawMessage(writer, request)
	if err != nil {
		log.Error().Msgf("an error occurred when extacting raw message from request: %s\n", err.Error())
		return
	}
	log.Info().Msg("Uri: " + request.RequestURI + ", requestBody: " + string(rawData))

	notificationHandler, err := h.determineHandler(writer, request)
	if err != nil {
		log.Error().Msgf("an error occurred when determine the handler: %s\n", err.Error())
		return
	}

	httpError := h.NotificationService.HandleNotification(notificationHandler, &rawData)
	if httpError != nil {
		h.returnErrorResponse(writer, httpError.ErrorCode, httpError.Message,
			httpError.StatusCode, &httpError.FieldErrors)
	}
}

func (h *Handler) StartServer(port string) error {
	addr := fmt.Sprintf(":%s", port)
	log.Info().Msgf("Server is running on http://localhost%s", addr)
	return http.ListenAndServe(addr, nil)
}

func (h *Handler) determineHandler(writer http.ResponseWriter, request *http.Request) (notification.INotificationHandler, error) {
	vars := mux.Vars(request)
	channel := vars["channel"]

	switch channel {
	case "email":
		return h.emailHandler, nil
	case "sms":
		return h.smsHandler, nil
	case "slack":
		return h.slackHandler, nil
	}

	h.returnErrorResponse(writer, "402", "Wrong notification channel type, correct"+
		" values are: 'email', 'slack', and 'sms'", http.StatusBadRequest, nil)

	missingHandlerError := goerrors.New("missing handler for channel type")
	return nil, missingHandlerError
}

func (h *Handler) extractRawMessage(writer http.ResponseWriter, request *http.Request) (json.RawMessage, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		h.returnErrorResponse(writer, "503", "An error occurred while reading request body",
			http.StatusInternalServerError, nil)

		return nil, err
	}

	return body, nil
}

func (h *Handler) returnErrorResponse(writer http.ResponseWriter, errorCode string, message string, statusCode int,
	fieldErrors *[]errors.FieldError) {
	errorBody := errors.Error{
		StatusCode: statusCode, ErrorCode: errorCode, Message: message, FieldErrors: *fieldErrors,
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	if err := json.NewEncoder(writer).Encode(errorBody); err != nil {
		log.Error().Msgf("error happened while encoding JSON response. Err: %s\n", err)
	}
}
