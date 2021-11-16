package web

import (
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/gocolly/colly/v2"
	"github.com/jasonlvhit/gocron"
	"inventory-watcher/pkg/cron"
	"log"
	"regexp"
	"strings"
)

type ISamg interface {
	cron.ICronAction
	samgShop(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{})
}

type samg struct {
	cron        *gocron.Scheduler
	queue       *goconcurrentqueue.FIFO
	sendMessage *map[string]interface{}
}

func NewSamg(c *gocron.Scheduler, q *goconcurrentqueue.FIFO, m *map[string]interface{}) cron.ICronAction {
	return &samg{
		cron:        c,
		queue:       q,
		sendMessage: m,
	}
}

func (s *samg) Start() {
	// 해외뽐뿌
	job := s.cron.Every(30).Seconds()
	job.Tag("samg", "삼진샵 > 반짝반짝캐치티니핑")
	job.Do(s.samgShop, s.queue, s.sendMessage)
	log.Println("Cron Start : SamgShop")

}

func (s *samg) Stop() {
	s.cron.Remove(s.samgShop)
	log.Println("Cron Stop : SamgShop")
}

func (s *samg) samgShop(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{}) {

	c := colly.NewCollector(
		//colly.Debugger(&debug.LogDebugger{}),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"),
		colly.AllowedDomains("samgshop.com"),
	)

	WEB_URL := "https://samgshop.com"
	PRODUCT_IDS := []string{"164"}

	log.Println("[SamgShop] Crawling...")
	var requestUrl string
	var productId string
	c.OnRequest(func(r *colly.Request) {
		requestUrl = r.URL.String()
	})

	c.OnHTML("body > script", func(e *colly.HTMLElement) {

		title := e.DOM.Closest("body").Find("div.infoArea > .headingArea > h2").Text()

		var text = e.Text
		if strings.Contains(text, "var is_soldout_icon") {
			//fmt.Println(e.Text)
			pattern := "is_soldout_icon([0-9]){0,}\\s=\\s\\'.*?(\\'\\;)"
			r := regexp.MustCompile(pattern)
			data := r.FindString(text)
			data = strings.Replace(data, "is_soldout_icon = '", "", 1)
			data = strings.Replace(data, "';", "", 1)

			// Soldout 여부 확인 (is_soldout_icon = T, F)
			if data == "T" {
				// 재고 없음
				log.Println(fmt.Sprintf("[SamgShop] %s (재고 X)", title))
				(*sendMessage)["SAMG_"+productId] = false
			} else if data == "F" {
				// 재고 있음
				log.Println(fmt.Sprintf("[SamgShop] %s (재고 O)", title))
				if (*sendMessage)["SAMG_"+productId] == nil {
					(*sendMessage)["SAMG_"+productId] = false
				}

				if v, ok := (*sendMessage)["SAMG_"+productId].(bool); ok && !v {
					messageMap := map[string]string{}
					messageMap["title"] = "[삼진샵 > 티니핑]"
					messageMap["text"] = title
					messageMap["linkUrl"] = requestUrl
					queue.Enqueue(messageMap)
				}
				(*sendMessage)["SAMG_"+productId] = true
				fmt.Println()
			}
		}
	})

	for _, product_id := range PRODUCT_IDS {
		productId = product_id
		c.Visit(fmt.Sprintf("%s/product/detail.html?product_no=%s&cate_no=82&display_group=1", WEB_URL, product_id))
	}

	log.Println("[SamgShop] Crawling end")
}

func (s *samg) IsRunning() bool {
	for _, v := range s.cron.Jobs() {
		if "samg" == (*v).Tags()[0] {
			return true
		}
	}

	return false
}
