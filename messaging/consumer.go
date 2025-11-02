package messaging

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	queueName = "canvas_creation_queue"
)

type CanvasCreationMessage struct {
	CanvasID string `json:"canvas_id"`
	UserID   string `json:"user_id"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func StartConsumer(dbpool *pgxpool.Pool) {
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}

	var conn *amqp.Connection
	var err error
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(rabbitmqURL)
		if err == nil {
			break
		}
		log.Printf("No se pudo conectar a RabbitMQ, reintentando en 5 segundos... (%v)", err)
		time.Sleep(5 * time.Second)
	}
	failOnError(err, "Failed to connect to RabbitMQ after several retries")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(1, 0, false)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (false, lo haremos manualmente)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			var msg CanvasCreationMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("Error decoding JSON: %s", err)
				d.Nack(false, false)
				continue
			}

			_, err := dbpool.Exec(context.Background(), "INSERT INTO canvas (id, user_id) VALUES ($1, $2)", msg.CanvasID, msg.UserID)
			if err != nil {
				log.Printf("Failed to create canvas in DB: %s", err)
				d.Nack(false, true)
				continue
			}

			log.Printf("Canvas created successfully with ID: %s", msg.CanvasID)
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
