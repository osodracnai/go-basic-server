package cmd

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/osodracnai/go-basic-server/pkg/consumer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func ConsumerCmd() *cobra.Command {
	command := &cobra.Command{
		Use:           "consumer",
		Short:         "Start the consumer process",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			var configCmd ConfigConsumerCmd
			if err = viper.Unmarshal(&configCmd, DecoderConfigOptions); err != nil {
				return fmt.Errorf("parse config: %v", err)
			}
			configCmd.Jaeger.Name = appName
			//jaeger
			jaegerCloser, err := configCmd.Jaeger.Config()
			if err != nil {
				logrus.Fatal(err)
			}
			defer func() {
				if err := jaegerCloser.Close(); err != nil {
					logrus.Fatal(err)
				}
			}()

			//kafka
			deferFunc, err := configCmd.Kafka.GetDefaultConfig()
			if err != nil {
				logrus.Fatal(err)
			}
			defer deferFunc()
			consumerGroup, err := sarama.NewConsumerGroupFromClient(configCmd.Kafka.GroupID, configCmd.Kafka.Client)
			if err != nil {
				logrus.Fatal(err)
			}
			topicsToRead := configCmd.Topics
			consumerHandler := consumer.Consumer{
				TopicRead: topicsToRead,
				Ready:     make(chan bool),
			}
			ctx, cancel := context.WithCancel(context.Background())
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					// `Consume` should be called inside an infinite loop, when a
					// server-side rebalance happens, the consumer session will need to be
					// recreated to get the new claims
					if err := consumerGroup.Consume(ctx, topicsToRead, &consumerHandler); err != nil {
						logrus.Fatalf("Error from consumer: %v", err)
					}
					// check if context was cancelled, signaling that the consumer should stop
					if ctx.Err() != nil {
						return
					}
					consumerHandler.Ready = make(chan bool)
				}
			}()

			<-consumerHandler.Ready // Await till the consumer has been set up
			logrus.Info("Sarama consumer up and running! ", configCmd.Topics)

			sigterm := make(chan os.Signal, 1)
			signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
			select {
			case <-ctx.Done():
				logrus.Info("terminating: context cancelled")
			case <-sigterm:
				logrus.Info("terminating: via signal")
			}
			cancel()
			wg.Wait()
			if err != nil {
				return err
			}
			return nil
		},
	}

	f := command.Flags()
	k := Kafka{GroupID: appName}
	j := Jaeger{Name: appName}
	j.Flags(f)
	k.Flags(f)
	consumerFlags(f)

	return command
}
