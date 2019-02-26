package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/yowithus/guessit/common"
	"github.com/yowithus/guessit/models"
)

var qna models.QNA
var congratsText = "Selamat kamu udah berhasil jawab semua pertanyaan dengan benar!! Ketik /mulai kalo mau main lagi :)"
var giveupText = "Yah kok nyerah? Better luck next time yah, semangat :)"
var replyText = ""
var correctAnswers []string
var isStarted = false
var correct = 0
var blank = "______________________"

var event *linebot.Event

func Play(c *gin.Context) {
	bot := common.GetBot()
	events, err := bot.ParseRequest(c.Request)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	for _, event = range events {
		userID := event.Source.UserID
		groupID := event.Source.GroupID
		roomID := event.Source.RoomID

		log.Println("User:", userID, " Group:", groupID, " Room:", roomID)

		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				input := message.Text
				replyText = ""

				switch input {
				case "/mulai":
					startGame()
				case "/ulang":
					restartGame()
				case "/nyerah":
					endGame()
				case "/hint":
					hint()
				default:
					guess(input)
				}

				reply(replyText)
			}
		}
	}

	c.JSON(http.StatusOK, "OK")
}

func startGame() {
	if isStarted {
		return
	}

	reset()

	qnas := common.GetQNAs()
	random := rand.Intn(len(qnas))
	qna = qnas[random]

	for i := 0; i < len(qna.Answers); i++ {
		num := i + 1
		correctAnswers = append(correctAnswers, fmt.Sprintf("%d. %s", num, blank))
	}

	question := qna.Question
	correctAnswersString := strings.Join(correctAnswers[:], "\n")
	replyText = fmt.Sprintf("%s\n%s", question, correctAnswersString)
	isStarted = true
}

func restartGame() {
	if !isStarted {
		return
	}

	reset()

	qnas := common.GetQNAs()
	random := rand.Intn(len(qnas))
	qna = qnas[random]

	for i := 0; i < len(qna.Answers); i++ {
		num := i + 1
		correctAnswers = append(correctAnswers, fmt.Sprintf("%d. %s", num, blank))
	}

	question := qna.Question
	correctAnswersString := strings.Join(correctAnswers[:], "\n")
	replyText = fmt.Sprintf("%s\n%s", question, correctAnswersString)
	isStarted = true
}

func endGame() {
	if !isStarted {
		return
	}

	answers := qna.Answers
	question := qna.Question

	for i := 0; i < len(answers); i++ {
		num := i + 1
		answerText := answers[i].Text
		answerScore := answers[i].Score
		correctAnswers[i] = fmt.Sprintf("%d. %s (%d)", num, answerText, answerScore)
	}

	correctAnswersString := strings.Join(correctAnswers[:], "\n")
	replyText = fmt.Sprintf("%s\n%s\n\n%s", question, correctAnswersString, giveupText)
	reset()
}

func guess(input string) {
	if !isStarted {
		return
	}

	for i := 0; i < len(correctAnswers); i++ {
		num := i + 1

		if strings.EqualFold(fmt.Sprintf("%d. %s", num, input), correctAnswers[i]) {
			return
		}
	}

	answers := qna.Answers
	question := qna.Question

	for i := 0; i < len(answers); i++ {
		answerText := answers[i].Text
		answerScore := answers[i].Score
		num := i + 1

		if strings.EqualFold(input, answerText) {
			correctAnswers[i] = fmt.Sprintf("%d. %s (%d)", num, answerText, answerScore)
			correctAnswersString := strings.Join(correctAnswers[:], "\n")
			replyText = fmt.Sprintf("%s\n%s", question, correctAnswersString)
			correct++

			if correct == len(answers) {
				replyText = fmt.Sprintf("%s\n\n%s", replyText, congratsText)
				reset()
			}
			return
		}
	}
}

func hint() {
	if !isStarted {
		return
	}

	var hint = ""

	for i := 0; i < len(correctAnswers); i++ {
		num := i + 1

		if strings.EqualFold(fmt.Sprintf("%d. %s", num, blank), correctAnswers[i]) {
			answerText := qna.Answers[i].Text
			for j := 0; j < len(answerText); j++ {
				c := answerText[j]

				if j == 0 || j == len(answerText)-1 {
					hint = fmt.Sprintf("%s %c", hint, c)
				} else {
					hint = fmt.Sprintf("%s _", hint)
				}
			}
			replyText = fmt.Sprintf("Ngestuck yah? Ini aku kasih hint buat kamu\nHint:%s\nTetep semangat :)", hint)
			return
		}
	}
}

func reset() {
	correct = 0
	correctAnswers = nil
	qna = models.QNA{}
	isStarted = false
}

func reply(text string) {
	if text == "" {
		return
	}

	bot := common.GetBot()
	if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(text)).Do(); err != nil {
		log.Print(err)
	}
}

func PlayTest(c *gin.Context) {
	input := c.Query("input")
	replyText = ""

	switch input {
	case "mulai":
		startGame()
	case "ulang":
		restartGame()
	case "nyerah":
		endGame()
	case "hint":
		hint()
	default:
		guess(input)
	}

	c.String(http.StatusOK, replyText)
}
