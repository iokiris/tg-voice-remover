package main

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type Broker struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      amqp.Queue
}

func connectRabbitMQ() (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	maxRetries := 15
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial("amqp://guest:guest@rabbitmq-tvr:5672/")
		if err == nil {
			log.Println("Successfully connected to RabbitMQ")
			return conn, nil
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}
	return nil, err
}

func NewBroker() *Broker {
	conn, err := connectRabbitMQ()
	if err != nil {
		log.Fatalf("RabbitMQ connection error %v", err)
	}
	log.Println("Connected to RabbitMQ")
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Cannot open connection channel")
	}
	q, err := ch.QueueDeclare(
		"task_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Cannot declare queue %v", err)
	}
	return &Broker{
		connection: conn,
		channel:    ch,
		queue:      q,
	}
}

func (broker *Broker) ensureChannel() error {
	if broker.channel != nil {
		if _, err := broker.channel.QueueInspect(broker.queue.Name); err == nil {
			return nil
		}
	}
	newCh, err := broker.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w\n", err)
	}
	broker.channel = newCh
	return nil
}

func (broker *Broker) ReceiveMessage() (<-chan amqp.Delivery, error) {
	if err := broker.ensureChannel(); err != nil {
		return nil, err
	}
	messages, err := broker.channel.Consume(
		broker.queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (broker *Broker) PushMessage(data []byte) error {

	err := broker.channel.Publish(
		"",
		broker.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         data,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		return err
	}
	return nil
}

func (broker *Broker) Close() {
	err := broker.channel.Close()
	if err != nil {
		return
	}
	err = broker.connection.Close()
	if err != nil {
		return
	}
}
