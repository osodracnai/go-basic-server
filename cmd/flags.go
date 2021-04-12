package cmd

import "github.com/spf13/pflag"

func serverFlags(flags *pflag.FlagSet) {
	flags.String("listen", "http://0.0.0.0:8080", "HTTP(S) address to listen at.")
}
func consumerFlags(flags *pflag.FlagSet) {
	flags.StringArray("topics", []string{"topic1"}, "topics to read")
}
