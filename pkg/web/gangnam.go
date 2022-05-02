package web

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/gocolly/colly/v2"
	"github.com/jasonlvhit/gocron"
	"github.com/misoboy/inventory-watcher/pkg/cron"
	"log"
)

type IGangnam interface {
	cron.ICronAction
	gangnam(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{})
}

type gangnam struct {
	cron        *gocron.Scheduler
	queue       *goconcurrentqueue.FIFO
	sendMessage *map[string]interface{}
}

func NewGangnam(c *gocron.Scheduler, q *goconcurrentqueue.FIFO, m *map[string]interface{}) cron.ICronAction {
	return &gangnam{
		cron:        c,
		queue:       q,
		sendMessage: m,
	}
}

func (s *gangnam) Start() {
	job := s.cron.Every(30).Seconds()
	job.Tag("gannam", "강남구청 > 미미위클린 놀이터")
	job.Do(s.gangnam, s.queue, s.sendMessage)
	log.Println("Cron Start : Gangnam")
}

func (s *gangnam) Stop() {
	s.cron.Remove(s.gangnam)
	log.Println("Cron Stop : Gangnam")
}

func (s *gangnam) gangnam(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{}) {

	c := colly.NewCollector(
		//colly.Debugger(&debug.LogDebugger{}),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"),
		colly.AllowedDomains("www.gangnam.go.kr"),
	)

	WEB_URL := "https://www.gangnam.go.kr"

	log.Println("[memeweclean] Crawling...")
	var requestUrl string
	c.OnRequest(func(r *colly.Request) {
		requestUrl = r.URL.String()
	})

	c.OnHTML("div[id=calendar] > div.calendar_tbl_wrap > table > tbody", func(e *colly.HTMLElement) {

		e.ForEach("tr", func(i int, tre *colly.HTMLElement) {

			tre.DOM.Find("td").Each(func(i int, s *goquery.Selection) {
				day := s.Find("div.calendar_day > span.pl5").Text()
				if day == "14" || day == "15" {
					isOpen := s.Find("div.fc-content").Find("p.mb5").Eq(1).Find("button")
					if isOpen.Length() > 0 {
						if v, ok := (*sendMessage)[fmt.Sprintf("GANGNAM_MEMEWE_IDX_%s", day)].(string); !ok || v != "true" {
							link, _ := s.Find("div.fc-content").Find("p.mb5").Eq(1).Find("button").Attr("href")

							messageMap := map[string]string{}
							messageMap["title"] = fmt.Sprintf("[강남구청 > 미미위클린 놀이터]")
							messageMap["text"] = fmt.Sprintf("미미위 클린 (%s일) 13:30 예약 가능", day)
							messageMap["imgUrl"] = ""
							messageMap["linkUrl"] = link
							queue.Enqueue(messageMap)
						}
						(*sendMessage)[fmt.Sprintf("GANGNAM_MEMEWE_IDX_%s", day)] = "true"
					}

				}
			})
		})
	})

	c.Visit(fmt.Sprintf("%s/resv/apply/memewe_clean_playground/list.do?mid=ID04_02071902", WEB_URL))

	log.Println("[memeweclean] Crawling end")
}

func (s *gangnam) IsRunning() bool {
	for _, v := range s.cron.Jobs() {
		if "gangnam" == (*v).Tags()[0] {
			return true
		}
	}

	return false
}
