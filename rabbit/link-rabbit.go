package main

import (
	"bytes"
	"curtis.io/link-checker/crawl"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	u "net/url"
	"strconv"
	s "strings"
)

func fail(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

var defaultServer = "amqp://guest:guest@localhost:5672/"
var numWorkers = 10

// TODO: will only process one message at a time,
// there should be no need for multiple crawler services,
// this should just serve multiple requesets in parallel,
// can use prefetch to limit amount going at once
func main() {
	conn, err := amqp.Dial(defaultServer)
	fail(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	forever := make(chan bool)
	for i := 0; i <= numWorkers; i++ {
		go listenOnQueue(conn, "rpc_queue", i)
	}
	fmt.Println("Awaiting rpc requests")
	<-forever

	fmt.Println("done")
}

func listenOnQueue(conn *amqp.Connection, queueName string, index int) {
	ch, err := conn.Channel()
	fail(err, "Failed to open channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	fail(err, "Failed to declare a queue")

	// what do these settings do?
	err = ch.Qos(1, 0, false)
	fail(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name,
		strconv.Itoa(index), // consumer
		false,
		false,
		false,
		false,
		nil,
	)
	fail(err, "Failed to register a consumer")

	for d := range msgs {
		url, err := u.Parse(string(d.Body))
		fail(err, "Failed to parse url.")

		log.Printf(" [.] crawl(%s)", url)

		crawlResults := crawl.Crawl(*url)

		i := 0
		var buffer bytes.Buffer
		for link := range crawlResults {
			i = (i + 1) % 10
			fmt.Print("appending " + s.Repeat(".", i) + "\r")
			buffer.WriteString(fmt.Sprint("%s, %d\n", link.Url, link.Status))
		}
		fmt.Println("")

		err = ch.Publish(
			"",
			d.ReplyTo,
			false,
			false,
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: d.CorrelationId,
				Body:          []byte(buffer.String()),
			},
		)
		fail(err, "Failed to publish a message")
		fmt.Println("Response sent")

		d.Ack(false)
	}
}
