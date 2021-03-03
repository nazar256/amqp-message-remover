package remover

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
	"regexp"
)

// MatchType indicates what will be matched with regexp
type MatchType int

const (
	// MatchBody used for matching AMQP-message bodies with regexp
	MatchBody MatchType = iota
	// MatchHeaders used for matching message headers (pre-serialized as json)
	MatchHeaders
)

// Config contains parameters for message deletion process
type Config struct {
	Dsn           string
	QueueName     string
	Regexp        *regexp.Regexp
	PrefetchCount int
	Nack          bool
	MatchType     MatchType
	Continuous    bool
}

// Status contains metrics about current deletion process
type Status struct {
	Processed int
	Removed   int
	Finished  bool
}

// StatusChannel read-only channel which provides Status updates
type StatusChannel <-chan Status

type msgChan <-chan amqp.Delivery

// RemoveMessages will connect to AMQP broker and start removing messages by Config
func RemoveMessages(config Config) StatusChannel {
	messages := initConsumer(config.Dsn, config.QueueName, config.PrefetchCount)
	statusCh := make(chan Status, 2)
	status := new(Status)

	go func() {
		for m := range messages {
			var subject []byte

			switch config.MatchType {
			case MatchBody:
				subject = m.Body
			case MatchHeaders:
				headersJson, err := json.Marshal(m.Headers)
				failOnError(err, "Failed to stringify message headers")
				subject = headersJson
			}
			if config.Regexp.Match(subject) {
				removeMessage(m, config.Nack)
				status.Removed += 1
			} else if config.Continuous {
				err := m.Nack(false, true)
				failOnError(err, "Failed on returning not matched message to queue")
			}
			status.Processed += 1
			statusCh <- *status
		}

		status.Finished = true
		statusCh <- *status
		close(statusCh)
	}()

	return statusCh
}

func removeMessage(m amqp.Delivery, isNack bool) {
	var err error
	if isNack {
		err = m.Nack(false, false)
	} else {
		err = m.Ack(false)
	}
	failOnError(err, "Failed to remove message")
}

func initConsumer(dsn string, queueName string, prefetch int) msgChan {
	conn, err := amqp.Dial(dsn)
	failOnError(err, "Could not connect to rabbitmq with "+dsn)

	ch, err := conn.Channel()
	failOnError(err, "Could not create amqp channel")

	err = ch.Qos(prefetch, 0, false)
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
