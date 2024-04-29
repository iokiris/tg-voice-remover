package main

import (
	"github.com/streadway/amqp"
	"log"
)

type Broker struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      amqp.Queue
}

func NewBroker() *Broker {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

func (broker *Broker) PushMessage(data []byte) error {

	err := broker.channel.Publish(
		"",
		broker.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
			//DeliveryMode: amqp.Persistent
		})
	if err != nil {
		return err
	}
	return nil
}

func (broker *Broker) ReceiveMessage() (<-chan amqp.Delivery, error) {
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
