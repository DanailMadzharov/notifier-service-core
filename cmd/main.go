package main

import (
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	apihandler "sumup-notifier-service/api/v1/http"
	"sumup-notifier-service/config"
	"sumup-notifier-service/core/notification"
)

func main() {
	loadConfig()

	log.Info().Msgf("Application Version: %s", viper.GetString("version"))
	brokers := viper.GetStringSlice("kafka.bootstrap-servers")

	slackHandlerKafkaWriter := kafka.Writer{
		Addr:  kafka.TCP(brokers...),
		Topic: viper.GetString("topic.slack.topic-name"),
	}
	slackHandler := notification.NewSlackHandler(
		&slackHandlerKafkaWriter,
	)

	smsHandlerKafkaWriter := kafka.Writer{
		Addr:  kafka.TCP(brokers...),
		Topic: viper.GetString("topic.sms.topic-name"),
	}
	smsHandler := notification.NewSMSHandler(
		&smsHandlerKafkaWriter,
	)

	emailHandlerKafkaWriter := kafka.Writer{
		Addr:  kafka.TCP(brokers...),
		Topic: viper.GetString("topic.emails.topic-name"),
	}
	emailHandler := notification.NewEmailHandler(
		&emailHandlerKafkaWriter,
	)

	handler := apihandler.NewHandler(notification.NewNotificationService(),
		smsHandler, slackHandler, emailHandler)

	apihandler.SetupRoutes(handler)

	err := handler.StartServer(viper.GetString("port"))
	if err != nil {
		log.Fatal().Msgf("error starting HTTP server: %s", err.Error())
	}
}

func loadConfig() {
	if err := config.Init(); err != nil {
		log.Fatal().Msgf("%s", err.Error())
	}
}
