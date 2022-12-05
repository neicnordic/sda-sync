package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

var (
	users      []string
	onemessage []interface{}
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// Function for generating accession ids
func generateIds(queue string) string {
	egaInt := 12000000000 + rand.Intn(1000)
	strNumber := strconv.Itoa(egaInt)
	id := ""
	if queue == "verified" {
		id = "EGAF" + strNumber
	} else {
		id = "EGAD" + strNumber
	}
	return id
}

// Function for checking if a string exists in an array
func contains(list []string, str string) bool {
	for _, value := range list {
		if value == str {
			return true
		}
	}
	return false
}

// Function for getting all the messages from the "completed" queue
// and create one message with all of them included
func getAllMessages(msgs <-chan amqp.Delivery, channel *amqp.Channel) {
	// Consume messages from the queue and create one message
	for delivered := range msgs {
		var message map[string]interface{}

		err := json.Unmarshal(delivered.Body, &message)
		failOnError(err, "Failed to unmarshal the message")
		log.Printf("Received a message from completed queue: %s", delivered.Body)

		// Append the delivered message to the one big message
		onemessage = append(onemessage, message)

		// Check if the user exists in the the users list and if
		// it is not then add the user
		exists := contains(users, message["user"].(string))
		if !exists {
			users = append(users, message["user"].(string))
		}

		// When the number of messages received from the "completed" queue
		// is equal to the number we want then create the new messages for mapping
		if len(onemessage) == 1 {
			dataSetMsgs(onemessage, users, channel, delivered.CorrelationId)
		}
	}
}

// Function for creating messages for mapping from the one big message.
// The number of the new messages is equal to the number of different
// users (the same dataset id will be given to all files that this user uploaded)
func dataSetMsgs(unMarBody []interface{}, users []string, channel *amqp.Channel, corrid string) {
	// Loop over the array of different users
	for _, user := range users {
		message := make(map[string]interface{})
		var ids []string
		// Loop through all the messages and add in an array all the accessions ids
		// from the user
		for _, dataset := range unMarBody {
			ds := dataset.(map[string]interface{})
			if user == ds["user"] {
				ids = append(ids, ds["accession_id"].(string))
			}
		}

		// Create a dataset id
		datasetID := generateIds("completed")

		// Add the necessary info to the new message
		message["type"] = "mapping"
		message["dataset_id"] = datasetID
		message["accession_ids"] = ids

		// Marshal the new body whith all the information
		createdBody, err := json.Marshal(message)
		failOnError(err, "Failed to marshal the new message for mapping")

		// Send the message to the files queue
		go sendMessage(createdBody, corrid, channel, "completed")
	}
}

// Fuction for consuming the messages in the queue
func consumeFromQueue(msgs <-chan amqp.Delivery, channel *amqp.Channel, queue string) {
	// For "completed" queue do not consume every incoming message.
	// Wait until all the messages are in the queue
	if queue == "completed" {
		getAllMessages(msgs, channel)
	}
	// Check the queue for messages
	for delivered := range msgs {
		//TODO: add json validation before calling the function
		log.Printf("Received a message from %v queue: %s", queue, delivered.Body)
		sendMessage(delivered.Body, delivered.CorrelationId, channel, queue)
	}
}

// Function for sending message to the file queue
//   - If the message comes with queue name "completed" then it is
//     already modified ready so no further information is needed.
//   - If the message comes from inbox queue: only the type is added.
//   - If the message comes from verified queue: type and accession id are added
//   - If the message comes from stableIDs queue: type and dataset id are added
func sendMessage(body []byte, corrid string, channel *amqp.Channel, queue string) {
	var newBody []byte
	if queue != "completed" {
		var message map[string]interface{}

		// Unmarshal the message
		// TODO: remove the error if json validation is implemented
		err := json.Unmarshal(body, &message)
		failOnError(err, "Failed to unmarshal the message")

		// Add the type in the received message depending on the queue
		if queue == "inbox" {
			delete(message, "filesize")
			delete(message, "operation")
			message["type"] = "ingest"
			log.Print(message)
		} else if queue == "verified" {
			message["type"] = "accession"
			accessionid := generateIds(queue)
			message["accession_id"] = accessionid
		} else if queue == "stableIDs" {
			message["type"] = "mapping"
			datasetid := generateIds(queue)
			message["dataset_id"] = datasetid
		}

		// Marshal the new body where the type is included
		newBody, err = json.Marshal(message)
		failOnError(err, "Failed to marshal the new message")
	} else {
		newBody = body
	}

	// Maybe move the context to the main
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publish message to the files queue
	err := channel.PublishWithContext(ctx,
		"localega.v1", // exchange
		"files",       // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentEncoding: "UTF-8",
			ContentType:     "application/json",
			DeliveryMode:    amqp.Persistent,
			CorrelationId:   corrid,
			Priority:        0, // 0-9
			Body:            []byte(newBody),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf("Send a message from %v queue to files: %s", queue, []byte(newBody))
}

// This function is using a channel to get the messages from a given queue
// and returns the messages
func messages(queue string, channel *amqp.Channel) <-chan amqp.Delivery {
	queueFullname := ""
	if queue == "stableIDs" {
		queueFullname = "v1." + queue
	} else {
		queueFullname = "v1.files." + queue
	}
	log.Printf("Consuming messages from %v queue", queueFullname)
	// Receive messages from the files.inbox queue
	messages, err := channel.Consume(
		queueFullname, // queue
		"",            // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	failOnError(err, "Failed to register a consumer")

	return messages
}

func main() {
	// Connect to the mock cega server
	conn, err := amqp.Dial("amqp://test:test@cegamq:5672/lega")
	failOnError(err, "Failed to connect to CEGA MQ")
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Queues that are checked for messages
	queues := []string{"inbox", "verified", "completed"}

	var forever chan struct{}

	// Loop over the given queues
	for _, queue := range queues {
		// Get the message from the queue
		msgs := messages(queue, ch)

		// Consume messages from specific queue
		go consumeFromQueue(msgs, ch, queue)
	}
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
