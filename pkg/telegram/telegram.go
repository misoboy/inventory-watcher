package telegram

import (
	"errors"
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	"inventory-watcher/pkg/cron"
	"log"
	"strconv"
	"time"
)

type ITelegram interface {
	Message()
	UpdateChannel()
}

type telegram struct {
	bot         *tgbotapi.BotAPI
	queue       *goconcurrentqueue.FIFO
	cron        *gocron.Scheduler
	cronActions *map[string]cron.ICronAction
	chatId      *string
}

func NewTelegram(
	bot *tgbotapi.BotAPI,
	queue *goconcurrentqueue.FIFO,
	subCron *gocron.Scheduler,
	cronActions *map[string]cron.ICronAction,
	chatId *string,
) ITelegram {
	return &telegram{
		bot:         bot,
		queue:       queue,
		cron:        subCron,
		cronActions: cronActions,
		chatId:      chatId,
	}
}

func (s *telegram) Message() {

	defer func() {
		s.recover(recover(), func(err error) {
			messageMap := map[string]string{}
			messageMap["title"] = "[Error] Message 에러 발생"
			messageMap["text"] = err.Error()
			s.queue.Enqueue(messageMap)
		})
	}()

	for true {

		if s.queue.GetLen() != 0 {

			item, err := s.queue.Dequeue()
			if err != nil {
				fmt.Println(err)
				return
			}

			v := item.(map[string]string)

			channelId, _ := strconv.ParseInt(*s.chatId, 10, 64)
			msg := tgbotapi.NewMessage(channelId, "")
			msg.ParseMode = "html"
			msg.Text = fmt.Sprintf("<strong>%s</strong>\n%s", v["title"], v["text"])
			if val, ok := v["imgUrl"]; ok && val != "" {
				msg.Text += fmt.Sprintf("\n<a href=\"%s\"> </a>\n[첨부이미지]", v["imgUrl"])
			}

			if val, ok := v["linkUrl"]; ok && val != "" {

				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL("링크 바로가기", v["linkUrl"]),
					),
				)
			}

			_, err = s.bot.Send(msg)
			if err == nil {
				log.Println("Telegram Send Message Completed")
			} else {
				log.Println("Telegram Send Message Failed")
				log.Panic(err)
			}
		}
	}
}

func (s *telegram) UpdateChannel() {

	defer func() {
		s.recover(recover(), func(err error) {
			messageMap := map[string]string{}
			messageMap["title"] = "[Error] UpdateChannel 에러 발생"
			messageMap["text"] = err.Error()
			s.queue.Enqueue(messageMap)
		})
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := s.bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			var fromId int64
			var text = ""
			var markup interface{}
			if update.CallbackQuery != nil {
				text, markup = s.commandCallback(update.CallbackQuery.Data)
				fromId = int64(update.CallbackQuery.From.ID)

			} else if update.Message != nil {
				fromId = update.Message.Chat.ID

				//if update.Message.NewChatMembers != nil {
				//	for _, v := range *update.Message.NewChatMembers {
				//		msg.Text += fmt.Sprintf("%s님", v.UserName)
				//	}
				//	msg.Text = "어서오세요. 반갑습니다132."
				//}

				if update.Message.IsCommand() {
					text, markup = s.commandAction(update.Message.Command())
				}
			}

			if text != "" {
				msg := tgbotapi.NewMessage(fromId, "")
				msg.ParseMode = "html"
				msg.ReplyMarkup = markup
				msg.Text = text

				_, err := s.bot.Send(msg)
				if err == nil {
					log.Println("Telegram Send Command Message Completed")
				} else {
					log.Println("Telegram Send Command Message Failed")
					log.Panic(err)
				}
			}
		}
	}
}

func (s *telegram) commandAction(command string) (string, interface{}) {

	var text = "/start or /help 도움말을 통해 버튼 기능을 이용 해주세요."
	var helpBtn interface{}
	if command == "start" || command == "help" {
		helpBtn = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("모두 실행", "start_all"),
				tgbotapi.NewInlineKeyboardButtonData("삼진 ✔", "start_samg"),
				tgbotapi.NewInlineKeyboardButtonData("아덴 ✔", "start_arden"),
				tgbotapi.NewInlineKeyboardButtonData("뽐뿌 ✔", "start_ppomppu"),
				tgbotapi.NewInlineKeyboardButtonData("루리웹 ✔", "start_ruliweb"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("모두 중지", "stop_all"),
				tgbotapi.NewInlineKeyboardButtonData("삼진 ✖", "stop_samg"),
				tgbotapi.NewInlineKeyboardButtonData("아덴 ✖", "stop_arden"),
				tgbotapi.NewInlineKeyboardButtonData("뽐뿌 ✖", "stop_ppomppu"),
				tgbotapi.NewInlineKeyboardButtonData("루리웹 ✖", "stop_ruliweb"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("상태 조회 ❤", "health"),
				tgbotapi.NewInlineKeyboardButtonData("도움말 ❔", "help"),
			),
		)

		text = "시작 및 도움말 안내.\n기능은 다음 버튼을 통해 실행 가능해요."

	}
	return text, helpBtn
}

func (s *telegram) commandCallback(commandData string) (string, interface{}) {

	var text string
	switch commandData {
	case "start_all":
		for _, v := range *s.cronActions {
			v.Start()
		}
		text = "[All] 배치 모두 시작 완료"
	case "start_samg":
		(*s.cronActions)["SAMG"].Start()
		text = "[삼진샵] 배치 시작 완료"
	case "start_arden":
		(*s.cronActions)["ARDEN"].Start()
		text = "[아덴바이크] 배치 시작 완료"
	case "start_ppomppu":
		(*s.cronActions)["PPOMPPU"].Start()
		text = "[뽐뿌] 배치 시작 완료"
	case "start_ruliweb":
		(*s.cronActions)["RULIWEB"].Start()
		text = "[루리웹] 배치 시작 완료"
	case "stop_all":
		for _, v := range *s.cronActions {
			v.Stop()
		}
		text = "[All] 배치 모두 중지 완료"
	case "stop_samg":
		(*s.cronActions)["SAMG"].Stop()
		text = "[삼진샵] 배치 중지 완료"
	case "stop_arden":
		(*s.cronActions)["ARDEN"].Stop()
		text = "[아덴바이크] 배치 중지 완료"
	case "stop_ppomppu":
		(*s.cronActions)["PPOMPPU"].Stop()
		text = "[뽐뿌] 배치 중지 완료"
	case "stop_ruliweb":
		(*s.cronActions)["RULIWEB"].Stop()
		text = "[루리웹] 배치 중지 완료"
	case "health":
		kst, _ := time.LoadLocation("Asia/Seoul")
		rootTxt := fmt.Sprintf("Root Cron Job Running : %d \n", len(gocron.Jobs()))
		for i, job := range gocron.Jobs() {
			rootTxt += fmt.Sprintf("%d) %s (%s)\n", i+1, job.Tags()[1], job.NextScheduledTime().In(kst).Format("2006-01-02 15:04:05"))
		}
		subTxt := fmt.Sprintf("Sub Cron Job Running : %d \n", len(s.cron.Jobs()))
		for i, job := range s.cron.Jobs() {
			subTxt += fmt.Sprintf("%d) %s (%s)\n", i+1, job.Tags()[1], job.NextScheduledTime().In(kst).Format("2006-01-02 15:04:05"))
		}
		text = rootTxt + "\n" + subTxt
	case "help":
		return s.commandAction("help")
	default:
		text = ""
	}

	return text, nil
}

func (s *telegram) recover(r interface{}, callback func(err error)) {
	var err error
	if r != nil {
		log.Println("Recovered in f", r)
		switch x := r.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = errors.New("unknown panic")
		}

		callback(err)
	}
}
