package telegram

import (
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	"inventory-watcher/cron"
	"log"
	"strconv"
	"strings"
)

func Message(bot *tgbotapi.BotAPI, queue *goconcurrentqueue.FIFO, chatId *string) {

	for true {

		if queue.GetLen() != 0 {

			item, err := queue.Dequeue()
			if err != nil {
				fmt.Println(err)
				return
			}

			v := item.(map[string]string)

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
			msg.Text = fmt.Sprintf("<b>%s</b>\n%s\n<a href=\"%s\">링크 바로가기</a>", v["title"], v["text"], v["linkUrl"])

			_, err = bot.Send(msg)
			if err == nil {
				log.Println("Telegram Send Message Completed")
			} else {
				log.Println("Telegram Send Message Failed")
			}
		}
	}
}

func UpdateChannel(bot *tgbotapi.BotAPI, cronActions *map[string]cron.ICronAction) {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.ParseMode = "html"
			arg := strings.TrimSpace(update.Message.CommandArguments())
			switch update.Message.Command() {
			case "help":
				msg.Text = "배치 실행 : /start {all,samg,arden,ppomppu,ruliweb}\n배치 중지 : /stop {all,samg,arden,ppomppu,ruliweb}\n상태 조회 : /health\n도움말 : /help"
			case "stop":
				if arg == "" {
					msg.Text = "Required Argument /stop {all,samg,arden,ppomppu,ruliweb}"
				} else if arg == "all" {
					gocron.Clear()
					msg.Text = "[All] 배치 모두 정지"
				} else if arg == "samg" {
					(*cronActions)["SAMG"].Stop()
					msg.Text = "[삼진샵] 배치 정지"
				} else if arg == "arden" {
					(*cronActions)["ARDEN"].Stop()
					msg.Text = "[아덴바이크] 배치 정지"
				} else if arg == "ppomppu" {
					(*cronActions)["PPOMPPU"].Stop()
					msg.Text = "[뽐뿌] 배치 정지"
				} else if arg == "ruliweb" {
					(*cronActions)["RULIWEB"].Stop()
					msg.Text = "[루리웹] 배치 정지"
				}
			case "start":
				if arg == "" {
					msg.Text = "Required Argument /start {all,samg,arden,ppomppu,ruliweb}"
				} else if arg == "all" {
					(*cronActions)["SAMG"].Start()
					(*cronActions)["ARDEN"].Start()
					(*cronActions)["PPOMPPU"].Start()
					(*cronActions)["RULIWEB"].Start()
					msg.Text = "[All] 배치 모두 시작"
				} else if arg == "samg" {
					(*cronActions)["SAMG"].Start()
					msg.Text = "[삼진샵] 배치 시작"
				} else if arg == "arden" {
					(*cronActions)["ARDEN"].Start()
					msg.Text = "[아덴바이크] 배치 시작"
				} else if arg == "ppomppu" {
					(*cronActions)["PPOMPPU"].Start()
					msg.Text = "[뽐뿌] 배치 시작"
				} else if arg == "ruliweb" {
					(*cronActions)["RULIWEB"].Start()
					msg.Text = "[루리웹] 배치 시작"
				}
			case "health":
				msg.Text = fmt.Sprintf("Healthy is Job Running : %d \n", len(gocron.Jobs()))
				for i, job := range gocron.Jobs() {
					msg.Text += fmt.Sprintf("%d) %s\n", i+1, job.Tags()[0])
				}
			default:
				msg.Text = "I don't know that command"
			}
			bot.Send(msg)
		}

	}
}
