package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	tb "gopkg.in/tucnak/telebot.v2"
)

func createBot() *tb.Bot {
	botToken := os.Getenv("BOT_TOKEN")

	bot, err := tb.NewBot(tb.Settings{
		Token:  botToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		panic(err)
	}

	bot.Handle(tb.OnText, handleMe)

	return bot
}

func handleMe(m *tb.Message) {
	log.Println(m.Sender.ID, m.Text)
}

func main() {
	chatId, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)

	if err != nil {
		log.Fatalln("CHAT_ID not present")
	}

	bot := createBot()
	log.Println("Bot created")

	go bot.Start()
	log.Println("Bot started")

	found := false

	log.Println("Starting loop")
	for {
		time.Sleep(1 * time.Minute)

		// Request the HTML page.
		res, err := http.Get("https://doktor24.se/vaccin/covid-vaccin/region-stockholm/")

		if err != nil {
			log.Println(err)
			continue
		}

		if res.StatusCode != 200 {
			log.Println("status code error:", res.StatusCode, res.Status)
			continue
		}

		// Load the HTML document
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			log.Println(err)
			continue
		}

		// Find the items
		items := doc.Find("#main-content > div.block.block--yellow.block--mobile-left.block--desktop-left.block--text > div.block__content.container > div.row.align-items-center > div > div > p:nth-child(2) > strong").Nodes

		if len(items) > 0 && items[0].FirstChild.Data == "stängd" {
			log.Println("list closed")

			found = false

			continue
		}

		log.Println("list open")

		if !found {
			found = true

			log.Println("sending message")
			bot.Send(&tb.Chat{ID: chatId}, "List open!\nhttps://doktor24.se/vaccin/covid-vaccin/region-stockholm/")
		}

	}

}
