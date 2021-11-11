package web

import (
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/gocolly/colly/v2"
	"github.com/jasonlvhit/gocron"
	"inventory-watcher/cron"
	"log"
	"strings"
)

type IPpomppu interface {
	cron.ICronAction
	bf_ppomppu(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{})
	shopping_ppomppu(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{})
}

type ppomppu struct {
	cron        *gocron.Scheduler
	queue       *goconcurrentqueue.FIFO
	sendMessage *map[string]interface{}
}

func NewPpomppu(c *gocron.Scheduler, q *goconcurrentqueue.FIFO, m *map[string]interface{}) cron.ICronAction {
	return &ppomppu{
		cron:        c,
		queue:       q,
		sendMessage: m,
	}
}

func (s *ppomppu) Start() {
	// 해외뽐뿌
	job := s.cron.Every(30).Seconds()
	job.Tag("ppomppu", "뽐뿌 > 해외뽐뿌")
	job.Do(s.bf_ppomppu, s.queue, s.sendMessage)
	log.Println("Cron Start : BF Ppomppu")
	// 쇼핑뽐뿌
	job1 := s.cron.Every(30).Seconds()
	job1.Tag("ppomppu", "뽐뿌 > 쇼핑뽐뿌")
	job1.Do(s.shopping_ppomppu, s.queue, s.sendMessage)
	log.Println("Cron Start : SHOPPING Ppomppu")
}

func (s *ppomppu) Stop() {
	s.cron.Remove(s.bf_ppomppu)
	log.Println("Cron Stop : BF Ppomppu")
	s.cron.Remove(s.shopping_ppomppu)
	log.Println("Cron Stop : SHOPPING Ppomppu")
}

func (s *ppomppu) bf_ppomppu(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{}) {

	c := colly.NewCollector(
		//colly.Debugger(&debug.LogDebugger{}),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"),
		colly.AllowedDomains("www.ppomppu.co.kr"),
	)

	WEB_URL := "https://www.ppomppu.co.kr"

	log.Println("[BF_Ppomppu] Crawling...")
	var requestUrl string
	c.OnRequest(func(r *colly.Request) {
		requestUrl = r.URL.String()
	})

	c.OnHTML("table[id=revolution_main_table] > tbody", func(e *colly.HTMLElement) {

		e.ForEach("tr", func(i int, tre *colly.HTMLElement) {
			if i == 6 {

				class := strings.TrimSpace(tre.Attr("class"))

				if class != "" && (class == "list1" || class == "list0") {

					idx := strings.TrimSpace(tre.DOM.Find("td").Eq(0).Text())
					//fmt.Println("idx : ", idx)

					imgSrc, _ := tre.DOM.Find("td").Eq(2).Find("tr").Find("td").Eq(0).Find("img").Attr("src")
					fullImgSrc := "https:" + imgSrc
					//fmt.Println("imgSrc : ", fullImgSrc)

					aTag := tre.DOM.Find("td").Eq(2).Find("tr").Find("td").Eq(1).Find("a")
					hrefSrc, _ := aTag.Attr("href")
					title := strings.TrimSpace(aTag.Text())
					//fmt.Println("title : ", title)

					if v, ok := (*sendMessage)["BF_PPOMPPU_IDX"].(string); !ok || v != idx {

						messageMap := map[string]string{}
						messageMap["title"] = fmt.Sprintf("[해외뽐뿌] %s", title)
						messageMap["text"] = fmt.Sprintf("<a href=\"%s\"> </a>", fullImgSrc)
						messageMap["linkUrl"] = fmt.Sprintf("%s/zboard/%s", WEB_URL, hrefSrc)
						queue.Enqueue(messageMap)
					}
					(*sendMessage)["BF_PPOMPPU_IDX"] = idx

				}
			}
		})
	})

	c.Visit(fmt.Sprintf("%s/zboard/zboard.php?id=ppomppu4", WEB_URL))

	log.Println("[BF_Ppomppu] Crawling end")
}

func (s *ppomppu) shopping_ppomppu(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{}) {

	c := colly.NewCollector(
		//colly.Debugger(&debug.LogDebugger{}),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"),
		colly.AllowedDomains("www.ppomppu.co.kr"),
	)

	WEB_URL := "https://www.ppomppu.co.kr"

	log.Println("[SHOPPING_Ppomppu] Crawling...")
	var requestUrl string
	c.OnRequest(func(r *colly.Request) {
		requestUrl = r.URL.String()
	})

	c.OnHTML("table[id=revolution_main_table] > tbody", func(e *colly.HTMLElement) {

		e.ForEach("tr", func(i int, tre *colly.HTMLElement) {
			if i == 6 {

				class := strings.TrimSpace(tre.Attr("class"))

				if class != "" && (class == "list1" || class == "list0") {

					idx := strings.TrimSpace(tre.DOM.Find("td").Eq(0).Text())
					//fmt.Println("idx : ", idx)

					imgSrc, _ := tre.DOM.Find("td").Eq(2).Find("tr").Find("td").Eq(0).Find("img").Attr("src")
					fullImgSrc := "https:" + imgSrc
					//fmt.Println("imgSrc : ", fullImgSrc)

					aTag := tre.DOM.Find("td").Eq(2).Find("tr").Find("td").Eq(1).Find("a")
					hrefSrc, _ := aTag.Attr("href")
					title := strings.TrimSpace(aTag.Text())
					//fmt.Println("title : ", title)

					if v, ok := (*sendMessage)["SHOPPING_PPOMPPU_IDX"].(string); !ok || v != idx {
						messageMap := map[string]string{}
						messageMap["title"] = fmt.Sprintf("[쇼핑뽐뿌] %s", title)
						messageMap["text"] = fmt.Sprintf("<a href=\"%s\"> </a>", fullImgSrc)
						messageMap["linkUrl"] = fmt.Sprintf("%s/zboard/%s", WEB_URL, hrefSrc)
						queue.Enqueue(messageMap)
					}
					(*sendMessage)["SHOPPING_PPOMPPU_IDX"] = idx

				}
			}
		})
	})

	c.Visit(fmt.Sprintf("%s/zboard/zboard.php?id=pmarket", WEB_URL))

	log.Println("[SHOPPING_Ppomppu] Crawling end")
}

func (s *ppomppu) IsRunning() bool {
	for _, v := range s.cron.Jobs() {
		if "ppomppu" == (*v).Tags()[0] {
			return true
		}
	}

	return false
}
