package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/osodracnai/go-basic-server/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"time"
)

func (k *Kafka) GetDefaultConfig() (func(), error) {
	kfversion, err := sarama.ParseKafkaVersion(k.Version) // kafkaVersion is the version of kafka server like 0.11.0.2
	if err != nil {
		return nil, err
	}
	kafkaconfig := sarama.NewConfig()
	kafkaconfig.Version = kfversion
	kafkaconfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaconfig.Producer.Retry.Max = 5
	kafkaconfig.Producer.Return.Successes = true
	kafkaconfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	k.Client, err = sarama.NewClient(k.Brokers, kafkaconfig)

	if err != nil {
		return nil, err
	}

	k.Consumer, err = sarama.NewConsumerGroupFromClient(k.GroupID, k.Client)
	if err != nil {
		return nil, err
	}

	k.Producer, err = sarama.NewSyncProducerFromClient(k.Client)

	if err != nil {
		return nil, err
	}

	def := func() {
		go func() {
			if err := k.Client.Close(); err != nil {
				logrus.Fatal(err)
			}
		}()
		go func() {
			if err := k.Consumer.Close(); err != nil {
				logrus.Fatal(err)
			}
		}()
		go func() {
			if err := k.Producer.Close(); err != nil {
				logrus.Fatal(err)
			}
		}()
	}
	return def, nil
}

func (k *Kafka) SendMessage(ctx context.Context, message interface{}, topic string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "KafkaSendMessage")
	defer span.Finish()
	headers := make(map[string]string)
	_ = opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(headers))

	object, err := json.Marshal(message)
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag(utils.KeyErrorMessage.String(), fmt.Sprintf("error unsmarshalling message: %v", err))
		logrus.Errorf("error unsmarshalling message: %v", err)
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.StringEncoder(object),
		Timestamp: time.Now(),
	}
	for headerKey, headerValue := range headers {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(headerKey),
			Value: []byte(headerValue),
		})
	}
	partition, offset, err := k.Producer.SendMessage(msg)
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag(utils.KeyErrorMessage.String(), fmt.Sprintf("error sending message to kafka: %v", err))
		logrus.Errorf("error sending message to kafka: %s", err.Error())
		return err
	}
	span.SetTag(utils.KeyKafkaTopic.String(), msg.Topic)
	span.SetTag(utils.KeyKafkaMessageLength.String(), msg.Value.Length())
	span.SetTag(utils.KeyKafkaPartition.String(), partition)
	span.SetTag(utils.KeyKafkaOffset.String(), offset)
	span.SetTag("span.kind", "client")
	logrus.Infof("Message is stored in topic %s/partition(%d)/offset(%d)\n", topic, partition, offset)
	return nil
}

func (k *Kafka) Flags(flags *pflag.FlagSet) {
	flags.StringSlice("kafka.brokers", []string{"localhost:9092"}, "Kafka brokers")
	flags.String("kafka.group-id", "groud-id default", "Kafka group ID")
	flags.String("kafka.version", "0.11.0.2", "Kafka version")
}
func (k *Kafka) GenerateSpanFromMessage(message *sarama.ConsumerMessage, operationName string) (context.Context, opentracing.Span) {
	headers := make(map[string]string)
	for _, header := range message.Headers {
		headers[string(header.Key)] = string(header.Value)
	}
	spanContext, _ := opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.TextMapCarrier(headers))
	span := opentracing.StartSpan(operationName, opentracing.FollowsFrom(spanContext))
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	span.SetTag(utils.KeyKafkaTopic.String(), message.Topic)
	span.SetTag(utils.KeyKafkaPartition.String(), message.Partition)
	span.SetTag(utils.KeyKafkaOffset.String(), message.Offset)
	span.SetTag("span.kind", "client")

	return ctx, span
}
