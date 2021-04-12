package cmd

import (
	"fmt"
	"github.com/osodracnai/go-basic-server/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/url"
)

func ServerCmd() *cobra.Command {
	command := &cobra.Command{
		Use:           "server",
		Short:         "Start the server process",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			var configCmd ConfigServerCmd
			if err = viper.Unmarshal(&configCmd, DecoderConfigOptions); err != nil {
				return fmt.Errorf("parse config: %v", err)
			}
			s, err := server.New()
			if err != nil {
				return err
			}
			listenURL, err := url.Parse(configCmd.Listen)
			if err != nil {
				logrus.Fatal(err)
			}

			r := s.NewEngine()
			logrus.Infof("Running server on %s", listenURL.Host)
			err = r.Run(listenURL.Host)
			if err != nil {
				return err
			}
			return nil
		},
	}
	f := command.Flags()
	serverFlags(f)

	return command
}
