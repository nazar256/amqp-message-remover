package remover

import (
	"encoding/json"
	"log"
	"regexp"

	"github.com/streadway/amqp"
)

type MatchType int

const (
	MatchBody MatchType = iota
	MatchHeaders
)

const (
	statusChannelBuffer = 2
)

type Config struct {
	Nack          bool
	Continuous    bool
	PrefetchCount uint16
	Dsn           string
	QueueName     string
	Regexp        *regexp.Regexp
	MatchType     MatchType
}

type Status struct {
	Processed int
	Removed   int
	Finished  bool
}

type StatusChannel <-chan Status

type msgChan <-chan amqp.Delivery

func RemoveMessages(config Config) StatusChannel {
	messages := initConsumer(config.Dsn, config.QueueName, config.PrefetchCount)
	statusCh := make(chan Status, statusChannelBuffer)
	status := new(Status)

	go func() {
		for m := range messages {
			var subject []byte

			switch config.MatchType {
			case MatchBody:
				subject = m.Body
			case MatchHeaders:
				headersJSON, err := json.Marshal(m.Headers)
				failOnError(err, "Failed to stringify message headers")

				subject = headersJSON
			}

			if config.Regexp.Match(subject) {
				removeMessage(m, config.Nack)
				status.Removed++
			} else if config.Continuous {
				err := m.Nack(false, true)
				failOnError(err, "Failed on returning not matched message to queue")
			}

			status.Processed++
			statusCh <- *status
		}

		status.Finished = true
		statusCh <- *status
		close(statusCh)
	}()

	return statusCh
}

func removeMessage(m amqp.Delivery, isNack bool) { // nolint:gocritic
	var err error
	if isNack {
		err = m.Nack(false, false)
	} else {
		err = m.Ack(false)
	}

	failOnError(err, "Failed to remove message")
}

func initConsumer(dsn, queueName string, prefetch uint16) msgChan {
	conn, err := amqp.Dial(dsn)
	failOnError(err, "Could not connect to rabbitmq with "+dsn)

	ch, err := conn.Channel()
	failOnError(err, "Could not create amqp channel")

	err = ch.Qos(int(prefetch), 0, false)
	failOnError(err, "Could not set prefetch")

	consumerCh, err := ch.Consume(
		queueName,         // queue
		"message-remover", // consumer
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	failOnError(err, "Could not start consumer")

	return consumerCh
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
