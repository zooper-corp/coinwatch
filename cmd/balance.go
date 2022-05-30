package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zooper-corp/CoinWatch/client"
	"github.com/zooper-corp/CoinWatch/display"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "balance",
	Short: "Update and dump current balance to std out",
	Run: func(cmd *cobra.Command, args []string) {
		configPath, _ := cmd.Flags().GetString("config")
		dbPath, _ := cmd.Flags().GetString("db-path")
		skipUpdate, _ := cmd.Flags().GetBool("skip-update")
		minUpdate, _ := cmd.Flags().GetInt("min-update")
		c, err := client.New(configPath, dbPath)
		if err != nil {
			fatal("Unable to create client: %v\n", err)
		}
		if !skipUpdate {
			err = c.UpdateBalance(int64(minUpdate) * 60)
			if err != nil {
				fatal("Unable to update: %v", err)
			}
		}
		style := display.GetDefaultAsciiTableStyle()
		style.Style = display.Wide
		style.Borders = true
		table, err := display.SummaryAsciiTable(&c, 7, style)
		if err != nil {
			fatal("Unable to dump table: %v", err)
		}
		fmt.Println(table)
		graph, err := display.TotalAsciiGraph(&c, 7, display.GetDefaultAsciiGraphStyle())
		if err != nil {
			fatal("Unable to dump graph: %v", err)
		}
		fmt.Println(graph)
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().BoolP("skip-update", "s", false, "Do not update balances and prices")
	dumpCmd.Flags().IntP("min-update", "m", 15, "Minimum time in minutes between updates")
}
