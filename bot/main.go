package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/zooper-corp/CoinWatch/client"
	"github.com/zooper-corp/CoinWatch/config"
	"github.com/zooper-corp/CoinWatch/display"
	"github.com/zooper-corp/CoinWatch/tools"
	"log"
	"strconv"
	"strings"
	"time"
)

type TelegramBot struct {
	config               config.TelegramBotConfig
	client               *client.Client
	bot                  *tgbotapi.BotAPI
	stopClientUpdateLoop chan struct{}
}

var mainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/sum 7"),
		tgbotapi.NewKeyboardButton("/sum 31"),
		tgbotapi.NewKeyboardButton("/allocation"),
		tgbotapi.NewKeyboardButton("/wallets"),
	),
)

func New(c *client.Client, cfg config.TelegramBotConfig) TelegramBot {
	bot, err := tgbotapi.NewBotAPI(cfg.ApiToken)
	if err != nil {
		log.Fatalf("Unable to start telegram bot: %v", err)
	}
	return TelegramBot{cfg, c, bot, make(chan struct{})}
}

func (b *TelegramBot) Start() {
	b.startClientUpdateLoop()
	b.sendTextMessage("Bot started")
	b.startPolling()
}

func (b *TelegramBot) startPolling() {
	b.bot.Debug = true
	log.Printf("Authorized on account %s", b.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Chat.ID != b.config.ChatId {
			log.Printf("Message from unauthorized user %s != %s", update.Message.From.UserName, b.config.ChatId)
			continue
		}
		// Valid
		b.onUpdate(update)
	}
}

func (b *TelegramBot) startClientUpdateLoop() {
	ticker := time.NewTicker(b.config.UpdateInterval)
	b.updateBalance()
	go func() {
		for {
			select {
			case <-ticker.C:
				b.updateBalance()
			case <-b.stopClientUpdateLoop:
				log.Printf("Stopping balance update ticker")
				ticker.Stop()
				return
			}
		}
	}()
}
func (b *TelegramBot) updateBalance() {
	log.Printf("Running ticker update")
	err := b.client.UpdateBalance(15)
	if err != nil {
		b.sendTextMessage(fmt.Sprintf("Balance update failed %v", err))
	}
}

func (b *TelegramBot) onUpdate(update tgbotapi.Update) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	cmd := strings.Split(strings.Trim(update.Message.Text, " "), " ")
	if len(cmd) == 0 {
		return
	}
	switch cmd[0] {
	case "/sum":
		days := getIntFromCmd(cmd, 1, 7)
		u := b.client.GetLastBalanceUpdate()
		// Table
		t, _ := display.SummaryAsciiTable(b.client, days, display.GetDefaultAsciiTableStyle())
		l := strings.Split(t, "\n")
		// Graph
		g, _ := display.TotalAsciiGraph(b.client, days, display.AsciiGraphStyle{Width: 25, Height: 8})
		// Dump
		b.sendHtmlMessage(fmt.Sprintf(
			"<b>Update</b>\n%s\n"+
				"<b>Balance</b>\n<pre>%s</pre>\n"+
				"<b>Summary</b>\n<pre>%s</pre>\n"+
				"<b>Performance</b>\n<pre>%s</pre>",
			u.Format(time.RFC822),
			strings.Join(l[0:len(l)-1], "\n"),
			l[len(l)-1:][0], g,
		))
	case "/allocation":
		days := getIntFromCmd(cmd, 1, 7)
		u := b.client.GetLastBalanceUpdate()
		t, _ := display.AllocationAsciiTable(b.client, days, display.GetDefaultAsciiTableStyle())
		b.sendHtmlMessage(fmt.Sprintf(
			"<b>Update</b>\n%s\n<b>Allocation</b>\n<pre>%s</pre>",
			u.Format(time.RFC822), t,
		))
	case "/wallets":
		balances := b.client.GetLastBalance()
		t := ""
		for _, wallet := range balances.Wallets() {
			t = t + fmt.Sprintf("<b>%s</b>\n", strings.ToUpper(wallet))
			for _, token := range balances.Tokens() {
				valid := false
				thead := fmt.Sprintf(" - <b>%s</b>\n", strings.ToUpper(token))
				for _, ba := range balances.Entries() {
					if ba.Token == token && ba.Wallet == wallet && ba.Balance != 0 {
						if !valid {
							valid = true
							t = t + thead
						}
						t = t + fmt.Sprintf(
							"   - %s [%s%s] <pre>%s</pre>\n",
							tools.HumanFloat64(ba.Balance),
							tools.HumanFloat64(ba.FiatValue),
							b.client.GetFiatSymbol(),
							ba.Address,
						)
					}
				}
			}
		}
		b.sendHtmlMessage(fmt.Sprintf(
			"<b>Wallets</b>\n%s\n%s",
			b.client.GetLastBalanceUpdate().Format(time.RFC822), t,
		))
	}
}

func (b *TelegramBot) sendTextMessage(text any) {
	b.sendMessage(text, "")
}

func (b *TelegramBot) sendHtmlMessage(text any) {
	b.sendMessage(text, tgbotapi.ModeHTML)
}

func (b *TelegramBot) sendMessage(text any, mode string) {
	msg := tgbotapi.NewMessage(b.config.ChatId, fmt.Sprintf("%v", text))
	msg.ParseMode = mode
	msg.ReplyMarkup = mainKeyboard
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("Unable to send message: %v", err)
		msg = tgbotapi.NewMessage(b.config.ChatId, fmt.Sprintf("%v", err))
		_, _ = b.bot.Send(msg)
	}
}

func getIntFromCmd(cmd []string, index int, defaultValue int) int {
	r := defaultValue
	if len(cmd) > index {
		i, err := strconv.Atoi(cmd[index])
		if err == nil {
			r = i
		}
	}
	return r
}
