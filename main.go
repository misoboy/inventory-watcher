package main

import (
	"github.com/enriquebris/goconcurrentqueue"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
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

	go telegram.UpdateChannel(bot)
	go telegram.Message(bot, queue, &TELEGRAM_CHAT_ID)

	var sendedMessage = make(map[string]interface{})
	gocron.Every(30).Seconds().Do(web.ArdenShop, queue, &sendedMessage)
	gocron.Every(30).Seconds().Do(web.SamgShop, queue, &sendedMessage)
	gocron.Every(30).Seconds().Do(web.BF_Ppomppu, queue, &sendedMessage)

	<-gocron.Start()

}
