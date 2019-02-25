package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

var question = "Siapa tokoh yang ada di anime Naruto??"
var answers = [5]string{"Naruto Uzumaki", "Sakura Haruno", "Sasuke Uchiha", "Kakashi Hatake", "Jiraiya"}
var answered = [5]string{"1. _________", "2. _________", "3. _________", "4. _________", "5. _________"}
var correct = 0
var isStarted = false

func main() {
	channelSecret := "fd273a9bbe0c96dd1679cdd18a758fd4"
	channelAccessToken := "ppy9xSN68/N+Gwgojzcuku2AP1IsF321+pM+at1LFE5C/KD+O5sWkAjwgiJyA9oiO3p6YTddIkgJoO0atvJd7EQ45qahjxl1oRydy1l+9Gto1cKx2VATWxSjLRdADMaiweIe4s6BGMuT5A8oAehXxgdB04t89/1O/w1cDnyilFU="
	bot, err := linebot.New(channelSecret, channelAccessToken)
	if err != nil {
		panic(err.Error())
	}

	r := gin.Default()

	r.POST("/callback", func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)

		log.Print(c.Request)

		if err != nil {
			if err == linebot.ErrInvalidSignature {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		for _, event := range events {
			// userID := event.Source.UserID
			// groupID := event.Source.GroupID
			// roomID := event.Source.RoomID
			//
			// log.Println("User:", userID, " Group:", groupID, " Room:", roomID)

			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					input := message.Text
					var replyText string

					if input == "/start" {
						isStarted = true
						answeredConcat := strings.Join(answered[:], "\n")
						replyText := fmt.Sprintf("%s\n%s", question, answeredConcat)

						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyText)).Do(); err != nil {
							log.Print(err)
						}

						return
					}

					if isStarted {
						if correct == 5 {
							replyText = "Congrats you have answered all correctly!! Please wait for the full features to release :)"

							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyText)).Do(); err != nil {
								log.Print(err)
							}

							return
						}

						for i := 0; i < 5; i++ {
							if strings.EqualFold(input, answers[i]) {
								correct++
								answered[i] = fmt.Sprintf("%d. %s", i+1, input)
								break
							}
						}

						answeredConcat := strings.Join(answered[:], "\n")

						if correct == 5 {
							replyText = fmt.Sprintf("%s\nCongrats you have answered all correctly!! Please wait for the full features to release :)", answeredConcat)

							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyText)).Do(); err != nil {
								log.Print(err)
							}

							return
						}

						replyText = answeredConcat
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyText)).Do(); err != nil {
							log.Print(err)
						}

						return
					}
				}
			}
		}

		c.JSON(http.StatusOK, "asd")
	})

	var port string
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	} else {
		port = "2205"
	}

	r.Run(fmt.Sprintf(":%s", port))
}
