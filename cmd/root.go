package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "CoinWatch",
	Short:            "Monitor different crypto addresses at once",
	PersistentPreRun: initLog,
	Long: `CryptoWatch will monitor balances of different crypto addresses on different chains at once, store past
data, calculate portfolio gains and losses. It can be either used from a terminal or serve directly via HTTP
both APIs and a minimal JS client. Check sub commands help for help.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func fatal(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Log to stderr")
	rootCmd.PersistentFlags().StringP("config", "c", "./config.yml", "Config file path")
	rootCmd.PersistentFlags().StringP("db-path", "d", "~/.coinwatch.db", "DB file path")
}

func initLog(cmd *cobra.Command, args []string) {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if !verbose {
		log.SetOutput(ioutil.Discard)
	}
}
