package web

import (
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/gocolly/colly/v2"
	"github.com/jasonlvhit/gocron"
	"github.com/misoboy/inventory-watcher/pkg/cron"
	"log"
	"strings"
)

type IRuliweb interface {
	cron.ICronAction
	hotdeal_ruliweb(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{})
}

type ruliweb struct {
	cron        *gocron.Scheduler
	queue       *goconcurrentqueue.FIFO
	sendMessage *map[string]interface{}
}

func NewRuliweb(c *gocron.Scheduler, q *goconcurrentqueue.FIFO, m *map[string]interface{}) cron.ICronAction {
	return &ruliweb{
		cron:        c,
		queue:       q,
		sendMessage: m,
	}
}

func (s *ruliweb) Start() {
	// 예판/핫딜
	job := s.cron.Every(30).Seconds()
	job.Tag("ruliweb", "루리웹 > 예판/핫딜")
	job.Do(s.hotdeal_ruliweb, s.queue, s.sendMessage)
	log.Println("Cron Start : Hotdeal ruliweb")
}

func (s *ruliweb) Stop() {
	s.cron.Remove(s.hotdeal_ruliweb)
	log.Println("Cron Stop : Hotdeal ruliweb")
}

func (s *ruliweb) hotdeal_ruliweb(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{}) {

	c := colly.NewCollector(
		//colly.Debugger(&debug.LogDebugger{}),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"),
		colly.AllowedDomains("bbs.ruliweb.com"),
	)

	WEB_URL := "https://bbs.ruliweb.com"

	log.Println("[hotdeal_ruliweb] Crawling...")
	var requestUrl string
	c.OnRequest(func(r *colly.Request) {
		requestUrl = r.URL.String()
	})

	c.OnHTML("table.board_list_table", func(e *colly.HTMLElement) {

		e.ForEach("tr.table_body", func(i int, tre *colly.HTMLElement) {
			if !strings.Contains(tre.Attr("class"), "best") && !strings.Contains(tre.Attr("class"), "notice") {
				elem := tre.DOM.Find("div.flex_wrapper > .flex_item").Eq(0)
				idx, _ := elem.Find("input[name=article_id]").Attr("value")
				idx = strings.TrimSpace(idx)
				imgSrc, _ := elem.Find("a.thumbnail").Attr("style")
				imgSrc = strings.Replace(strings.TrimSpace(imgSrc), "background-image: url(", "", 1)
				imgSrc = strings.Replace(imgSrc, ");", "", 1)
				hrefSrc, _ := elem.Find("a.thumbnail").Attr("href")
				title := strings.TrimSpace(elem.Find("a.deco").Text())

				if v, _ := (*sendMessage)["HOTDEAL_RULIWEB_IDX"].(string); v != idx {

					messageMap := map[string]string{}
					messageMap["title"] = fmt.Sprintf("[루리웹 > 예판/핫딜]")
					messageMap["text"] = title
					messageMap["imgUrl"] = imgSrc
					messageMap["linkUrl"] = hrefSrc
					queue.Enqueue(messageMap)
				}
				(*sendMessage)["HOTDEAL_RULIWEB_IDX"] = idx
			}
		})
	})

	c.Visit(fmt.Sprintf("%s/nin/board/1020?view=gallery", WEB_URL))

	log.Println("[hotdeal_ruliweb] Crawling end")
}

func (s *ruliweb) IsRunning() bool {
	for _, v := range s.cron.Jobs() {
		if "ruliweb" == (*v).Tags()[0] {
			return true
		}
	}

	return false
}
