package web

import (
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/gocolly/colly/v2"
	"inventory-watcher/telegram"
	"log"
	"regexp"
	"strings"
)

func SamgShop(queue *goconcurrentqueue.FIFO, sendedMessage *map[string]interface{}) {

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
				(*sendedMessage)["SAMG_"+productId] = false
			} else if data == "F" {
				// 재고 있음
				log.Println(fmt.Sprintf("[SamgShop] %s (재고 O)", title))
				if v, ok := (*sendedMessage)["SAMG_"+productId].(bool); ok && !v {
					queue.Enqueue(telegram.MessageForm{
						Title: fmt.Sprintf("[%s]", title), Text: "구매 가능..!!!", LinkUrl: requestUrl,
					})
				}
				(*sendedMessage)["SAMG_"+productId] = true
			}
		}
	})

	for _, product_id := range PRODUCT_IDS {
		productId = product_id
		c.Visit(fmt.Sprintf("%s/product/detail.html?product_no=%s&cate_no=82&display_group=1", WEB_URL, product_id))
	}

	log.Println("[SamgShop] Crawling end")
}
