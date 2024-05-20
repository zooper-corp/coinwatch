package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zooper-corp/CoinWatch/api"
	"github.com/zooper-corp/CoinWatch/client"
	"github.com/zooper-corp/CoinWatch/config"
	"os"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Starts the API server",
	Run: func(cmd *cobra.Command, args []string) {
		configPath, _ := cmd.Flags().GetString("config")
		dbPath, _ := cmd.Flags().GetString("db-path")
		c, err := client.New(configPath, dbPath)
		if err != nil {
			fatal("Unable to create client: %v", err)
		}
		apiKey, _ := cmd.Flags().GetString("api-key")
		if apiKey == "" {
			apiKey = os.Getenv("API_KEY")
		}
		if apiKey == "" {
			fatal("API key not provided")
		}
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		server := api.NewApiServer(&c, config.ApiServerConfig{
			Host:   host,
			Port:   port,
			ApiKey: apiKey,
		})
		server.Start()
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)
	apiCmd.Flags().StringP("config", "c", "", "Path to the configuration file")
	apiCmd.Flags().StringP("db-path", "d", "", "Path to the database")
	apiCmd.Flags().StringP("api-key", "a", "", "API key for authentication")
	apiCmd.Flags().StringP("host", "H", "0.0.0.0", "Host for the API server")
	apiCmd.Flags().IntP("port", "p", 8080, "Port for the API server")
}
