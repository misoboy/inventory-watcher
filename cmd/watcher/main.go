package main

import (
	"github.com/enriquebris/goconcurrentqueue"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	"github.com/misoboy/inventory-watcher/pkg/cron"
	"github.com/misoboy/inventory-watcher/pkg/telegram"
	web "github.com/misoboy/inventory-watcher/pkg/web"
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
	subCron := gocron.NewScheduler()

	sendMessage := make(map[string]interface{})
	cronActions := make(map[string]cron.ICronAction)
	//cronActions["ARDEN"] = web.NewArdenbike(subCron, queue, &sendMessage)
	//cronActions["SAMG"] = web.NewSamg(subCron, queue, &sendMessage)
	//cronActions["PPOMPPU"] = web.NewPpomppu(subCron, queue, &sendMessage)
	cronActions["RULIWEB"] = web.NewRuliweb(subCron, queue, &sendMessage)
	cronActions["GANGNAM"] = web.NewGangnam(subCron, queue, &sendMessage)
	//cronActions["HIMART"] = web.NewHimart(subCron, queue, &sendMessage)

	mainCron := cron.NewCron(subCron, &cronActions, queue)
	// 오전 7시 (UTC + 9Hour)
	job1 := gocron.Every(1).Day().At("22:00:00")
	job1.Tag("startat", "Root StartAt Job")
	job1.Do(mainCron.Start)
	// 오후 10시 (UTC + 9Hour)
	job2 := gocron.Every(1).Day().At("13:00:00")
	job2.Tag("endat", "Root EndAt Job")
	job2.Do(mainCron.Stop)

	go func() {
		<-subCron.Start()
	}()

	tg := telegram.NewTelegram(bot, queue, subCron, &cronActions, &TELEGRAM_CHAT_ID)
	go tg.UpdateChannel()
	go tg.Message()

	<-gocron.Start()
}
