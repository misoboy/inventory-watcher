package telegram

import (
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	"log"
	"strconv"
)

type MessageForm struct {
	Title   string
	Text    string
	LinkUrl string
}

func Message(bot *tgbotapi.BotAPI, queue *goconcurrentqueue.FIFO, chatId *string) {

	for true {

		if queue.GetLen() != 0 {

			item, err := queue.Dequeue()
			if err != nil {
				fmt.Println(err)
				return
			}

			v := item.(MessageForm)

			//resp, _ := client.R().
			//	SetQueryParams(map[string]string{
			//		"chat_id": TELEGRAM_CHAT_ID,
			//		"text": fmt.Sprintf("<b>%s</b>\n%s\n<a href=\"%s\">링크 바로가기</a>", v.Title, v.Text,v.LinkUrl),
			//		"parse_mode": "HTML",
			//	}).
			//	EnableTrace().
			//	Get(fmt.Sprintf("%s/bot%s/sendMessage", TELEGRAM_API_URL, TELEGRAM_BOT_TOKEN))

			channelId, _ := strconv.ParseInt(*chatId, 10, 64)
			msg := tgbotapi.NewMessage(channelId, "")
			msg.ParseMode = "html"
			msg.Text = fmt.Sprintf("<b>%s</b>\n%s\n<a href=\"%s\">링크 바로가기</a>", v.Title, v.Text, v.LinkUrl)

			_, err = bot.Send(msg)
			if err == nil {
				log.Println("Telegram Send Message Completed")
			} else {
				log.Println("Telegram Send Message Failed")
			}
		}
	}
}

func UpdateChannel(bot *tgbotapi.BotAPI) {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.ParseMode = "html"
			arg := update.Message.CommandArguments()
			switch update.Message.Command() {
			case "stop":
				if arg == "" {
					msg.Text = "Required Argument /stop {all,samg}"
				} else if arg == "all" {
					gocron.Clear()
					msg.Text = "All Stop Complete"
				}
			default:
				msg.Text = "I don't know that command"
			}
			bot.Send(msg)
		}

	}
}
