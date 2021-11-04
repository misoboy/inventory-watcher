package web

import (
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/gocolly/colly/v2"
	"inventory-watcher/telegram"
	"log"
	"strings"
)

func BF_Ppomppu(queue *goconcurrentqueue.FIFO, sendedMessage *map[string]interface{}) {

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

					if v, ok := (*sendedMessage)["BF_PPOMPPU_IDX"].(string); !ok || v != idx {
						queue.Enqueue(telegram.MessageForm{
							Title: fmt.Sprintf("[%s]", title), Text: fmt.Sprintf("<a href=\"%s\"> </a>", fullImgSrc), LinkUrl: fmt.Sprintf("%s/zboard/%s", WEB_URL, hrefSrc),
						})
					}
					(*sendedMessage)["BF_PPOMPPU_IDX"] = idx

				}
			}
		})
	})

	c.Visit(fmt.Sprintf("%s/zboard/zboard.php?id=ppomppu4", WEB_URL))

	log.Println("[BF_Ppomppu] Crawling end")
}
