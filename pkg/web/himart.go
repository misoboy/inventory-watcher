package web

import (
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/gocolly/colly/v2"
	"github.com/jasonlvhit/gocron"
	"github.com/misoboy/inventory-watcher/pkg/cron"
	"log"
	"strconv"
	"strings"
)

type IHimart interface {
	cron.ICronAction
	himart(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{})
}

type himart struct {
	cron        *gocron.Scheduler
	queue       *goconcurrentqueue.FIFO
	sendMessage *map[string]interface{}
}

func NewHimart(c *gocron.Scheduler, q *goconcurrentqueue.FIFO, m *map[string]interface{}) cron.ICronAction {
	return &himart{
		cron:        c,
		queue:       q,
		sendMessage: m,
	}
}

func (s *himart) Start() {
	job := s.cron.Every(30).Seconds()
	job.Tag("himart", "HIMART > 엘든링")
	job.Do(s.himart, s.queue, s.sendMessage)
	log.Println("Cron Start : himart")
}

func (s *himart) Stop() {
	s.cron.Remove(s.himart)
	log.Println("Cron Stop : himart")
}

func (s *himart) himart(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{}) {

	c := colly.NewCollector(
		//colly.Debugger(&debug.LogDebugger{}),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"),
		colly.AllowedDomains("www.e-himart.co.kr"),
	)

	WEB_URL := "https://www.e-himart.co.kr"

	log.Println("[himart] Crawling...")
	var requestUrl string
	c.OnRequest(func(r *colly.Request) {
		requestUrl = r.URL.String()
	})

	c.OnHTML("div.prdItemList > ul", func(e *colly.HTMLElement) {

		e.ForEach("li", func(i int, tre *colly.HTMLElement) {

			prdItem := tre.DOM.Find("div.prdItem")
			goodsNo, goodsNoExists := prdItem.Attr("goodsno")
			if goodsNoExists && goodsNo == "0016862655" {
				discountPrice := strings.Replace(strings.TrimSpace(prdItem.Find("div.priceBenefit > .discountPrice > strong").Text()), ",", "", 1)
				//println("discountPrice : ", discountPrice)
				if sv, _ := strconv.Atoi(discountPrice); sv <= 53000 {
					if v, ok := (*sendMessage)["ELDENRING_HIMART_IDX"].(string); !ok || v != "true" {

						messageMap := map[string]string{}
						messageMap["title"] = fmt.Sprintf("[Himart > 엘든링]")
						messageMap["text"] = fmt.Sprintf("엘든링 가격 변동 발생 (%s)", discountPrice)
						messageMap["imgUrl"] = ""
						messageMap["linkUrl"] = "https://www.e-himart.co.kr/app/goods/goodsDetail?goodsNo=0016862655"
						queue.Enqueue(messageMap)
					}
					(*sendMessage)["ELDENRING_HIMART_IDX"] = "true"

				}

			}
		})
	})

	c.Visit(fmt.Sprintf("%s/app/search/totalSearch?query=엘든링", WEB_URL))

	log.Println("[himart] Crawling end")
}

func (s *himart) IsRunning() bool {
	for _, v := range s.cron.Jobs() {
		if "himart" == (*v).Tags()[0] {
			return true
		}
	}

	return false
}
