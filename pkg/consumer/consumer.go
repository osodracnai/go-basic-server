package consumer

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/osodracnai/go-basic-server/utils"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	TopicRead []string
	Ready     chan bool
}

func (ch *Consumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}
func (ch *Consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}
func (ch *Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var err error

		log := logrus.WithFields(logrus.Fields{
			utils.KeyKafkaTopic.String():  msg.Topic,
			utils.KeyKafkaOffset.String(): msg.Offset,
		})
		log.Info("Processing Message")
		if msg.Topic == "topic1" {
			err = ch.readMessage(msg)
		}

		if err != nil {
			return err
		}
		sess.MarkMessage(msg, "")
		log.Info("Message processed")
	}
	return nil
}

func (ch *Consumer) readMessage(msg *sarama.ConsumerMessage) error {

	var msgInterface interface{}
	err := json.Unmarshal(msg.Value, &msgInterface)
	if err != nil {
		logrus.Error(err)
		return err
	}
	fmt.Println("Reading msg:")

	fmt.Println(string(msg.Value))

	fmt.Println("Msg readed:")

	return nil
}
