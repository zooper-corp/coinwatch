package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zooper-corp/CoinWatch/bot"
	"github.com/zooper-corp/CoinWatch/client"
	"github.com/zooper-corp/CoinWatch/config"
	"log"
	"os"
	"strconv"
	"time"
)

// botCmd represents the bot command
var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Starts a Telegram bot",
	Run: func(cmd *cobra.Command, args []string) {
		configPath, _ := cmd.Flags().GetString("config")
		dbPath, _ := cmd.Flags().GetString("db-path")
		c, err := client.New(configPath, dbPath)
		if err != nil {
			fatal("Unable to create client: %v", err)
		}
		// Dump env
		log.Printf("BOT_TOKEN:%s CHAT_D:%s", os.Getenv("BOT_TOKEN"), os.Getenv("BOT_CHAT"))
		// Check bot config
		updateInterval, _ := cmd.Flags().GetInt("update-interval")
		botToken, _ := cmd.Flags().GetString("token")
		if botToken == "" {
			botToken = os.Getenv("BOT_TOKEN")
		}
		botChat, _ := cmd.Flags().GetInt64("chat-id")
		if botChat == 0 {
			botChatInt, err := strconv.Atoi(os.Getenv("BOT_CHAT"))
			if err == nil {
				botChat = int64(botChatInt)
			}
		}
		if botChat == 0 || botToken == "" {
			fatal("Bot token or chat ID not provided")
		}
		b := bot.New(&c, config.TelegramBotConfig{
			ApiToken:       botToken,
			ChatId:         botChat,
			UpdateInterval: time.Minute * time.Duration(updateInterval),
		})
		b.Start()
	},
}

func init() {
	rootCmd.AddCommand(botCmd)
	botCmd.Flags().String("token", "", "Telegram API token, overrides BOT_TOKEN env var")
	botCmd.Flags().Int64("chat-id", 0, "Telegram admin chat, overrides BOT_CHAT env var")
	botCmd.Flags().IntP("update-interval", "i", 15, "Time in minutes between updates")
}
