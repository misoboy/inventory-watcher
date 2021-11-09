package web

import (
	"encoding/json"
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/gocolly/colly/v2"
	"github.com/jasonlvhit/gocron"
	"inventory-watcher/cron"
	"log"
	"regexp"
	"strings"
	_ "strings"
)

type IArdenbike interface {
	cron.ICronAction
	ardenShop(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{})
}

type ardenbike struct {
	queue       *goconcurrentqueue.FIFO
	sendMessage *map[string]interface{}
}

func NewArdenbike(q *goconcurrentqueue.FIFO, m *map[string]interface{}) IArdenbike {
	return &ardenbike{
		queue:       q,
		sendMessage: m,
	}
}

func (s *ardenbike) Start() {
	job := gocron.Every(30).Seconds()
	job.Tag("아덴바이크 > 프라임 하드쉘 의류")
	job.Do(s.ardenShop, s.queue, s.sendMessage)
	log.Println("Cron Start : ArdenShop")
}

func (s *ardenbike) Stop() {
	gocron.Remove(s.ardenShop)
	log.Println("Cron Stop : ArdenShop")
}

func (s *ardenbike) ardenShop(queue *goconcurrentqueue.FIFO, sendMessage *map[string]interface{}) {

	c := colly.NewCollector(
		//colly.Debugger(&debug.LogDebugger{}),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"),
		colly.AllowedDomains("ardenbike.co.kr"),
	)

	WEB_URL := "https://ardenbike.co.kr"
	PRODUCT_IDS := []string{"4141"}

	log.Println("[ArdenShop] Crawling...")
	var requestUrl string
	var productId string
	c.OnRequest(func(r *colly.Request) {
		requestUrl = r.URL.String()
	})

	c.OnHTML("body > script", func(e *colly.HTMLElement) {

		title := strings.TrimSpace(e.DOM.Closest("body").Find("div.infomation > h3").Text())

		var text = e.Text
		if strings.Contains(text, "var option_stock_data") {
			//fmt.Println(e.Text)
			pattern := "option_stock_data([0-9]){0,}\\s=\\s\\'.*?(\\'\\;)"
			r := regexp.MustCompile(pattern)
			jsonData := r.FindString(text)
			jsonData = strings.Replace(jsonData, "option_stock_data = '", "", 1)
			jsonData = strings.Replace(jsonData, "';", "", 1)
			jsonData = strings.ReplaceAll(jsonData, "\\", "")
			data := make(map[string]interface{})
			json.Unmarshal([]byte(jsonData), &data)

			for _, v := range data {
				detailMap := v.(map[string]interface{})
				stockNumber := (detailMap["stock_number"]).(float64)
				optionValue := (detailMap["option_value"]).(string)
				if optionValue == "XL" || optionValue == "XXL" {

					if stockNumber > 0 {
						log.Println(fmt.Sprintf("[ArdenShop] %s [사이즈 : %s] (재고 O)", title, optionValue))
						if v, ok := (*sendMessage)["ARDEN_"+productId].(bool); ok && !v {
							messageMap := map[string]string{}
							messageMap["title"] = fmt.Sprintf("[아덴바이크] %s [사이즈 : %s]", title, optionValue)
							messageMap["text"] = "구매 가능..!!!"
							messageMap["linkUrl"] = requestUrl
							queue.Enqueue(messageMap)
						}
						(*sendMessage)["ARDEN_"+productId] = true
					} else {
						log.Println(fmt.Sprintf("[ArdenShop] %s [사이즈 : %s] (재고 X)", title, optionValue))
						(*sendMessage)["ARDEN_"+productId] = false
					}
				}
			}
		}
	})

	for _, product_id := range PRODUCT_IDS {
		productId = product_id
		c.Visit(fmt.Sprintf("%s/product/detail.html?product_no=%s&cate_no=66&display_group=1#none", WEB_URL, product_id))
	}

	log.Println("[ArdenShop] Crawling end")
}
