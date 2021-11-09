package main

import (
	"github.com/enriquebris/goconcurrentqueue"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	"inventory-watcher/cron"
	"inventory-watcher/telegram"
	"inventory-watcher/web"
	"log"
	"os"
)

var TELEGRAM_BOT_TOKEN = os.Getenv("TELEGRAM_BOT_TOKEN")
var TELEGRAM_CHAT_ID = os.Getenv("TELEGRAM_CHAT_ID")

func main() {

	bot, err := tgbotapi.NewBotAPI(TELEGRAM_BOT_TOKEN)
	//bot.Debug = true
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	queue := goconcurrentqueue.NewFIFO()

	var sendMessage = make(map[string]interface{})
	cronActions := make(map[string]cron.ICronAction)
	cronActions["ARDEN"] = web.NewArdenbike(queue, &sendMessage)
	cronActions["SAMG"] = web.NewSamg(queue, &sendMessage)
	cronActions["PPOMPPU"] = web.NewPpomppu(queue, &sendMessage)
	cronActions["RULIWEB"] = web.NewRuliweb(queue, &sendMessage)
	go telegram.UpdateChannel(bot, &cronActions)
	go telegram.Message(bot, queue, &TELEGRAM_CHAT_ID)

	for _, v := range cronActions {
		v.Start()
	}

	<-gocron.Start()

}
