package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-redis/redis/v8"
	tb "gopkg.in/tucnak/telebot.v2"
)

const VACCINE_KEY = "vaccine-open"

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

var ctx = context.Background()

func main() {
	chatId, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)

	redisHost := os.Getenv("REDIS_HOST")
	redisPass := os.Getenv("REDIS_PASSWORD")

	if err != nil {
		log.Fatalln("CHAT_ID not present")
	}

	bot := createBot()
	log.Println("Bot created")

	go bot.Start()
	log.Println("Bot started")

	log.Println("Starting redis client")
	db := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPass,
		DB:       0,
	})

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
		items := doc.Find("#reservlista-covid-19").Nodes

		if len(items) > 0 && strings.Contains(items[0].FirstChild.Data, "stängd") {
			log.Println("list closed")

			err := db.Set(ctx, VACCINE_KEY, false, 0).Err()
			if err != nil {
				bot.Send(&tb.Chat{ID: chatId}, "Failed to set redis key on list close")
			}

			continue
		}

		log.Println("list open")

		key, err := db.Get(ctx, VACCINE_KEY).Result()
		if err != nil {
			bot.Send(&tb.Chat{ID: chatId}, "Failed to get redis key on list open")
		}

		found, _ := strconv.ParseBool(key)

		if !found {
			err := db.Set(ctx, VACCINE_KEY, true, 0).Err()
			if err != nil {
				bot.Send(&tb.Chat{ID: chatId}, "Failed to set redis key on list open")
			}

			log.Println("sending message")
			bot.Send(&tb.Chat{ID: chatId}, "List open!\nhttps://doktor24.se/vaccin/covid-vaccin/region-stockholm/")
		}

	}

}
