package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/yowithus/guessit/common"
	"github.com/yowithus/guessit/models"
)

var qna models.QNA
var replyText = ""
var correctAnswers []string
var correctFullAnswers []string
var isStarted = false
var correct = 0
var blank = "______________________"
var scoreBoards []models.ScoreBoard

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
		profile, err := bot.GetProfile(userID).Do()
		if err != nil {
			log.Println(err)
		}
		name := profile.DisplayName

		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				input := message.Text
				replyText = ""
				game(input, name)
				reply(replyText)
			}
		}
	}

	c.JSON(http.StatusOK, replyText)
}

func PlayTest(c *gin.Context) {
	input := c.Query("input")
	name := c.Query("name")
	replyText = ""
	game(input, name)
	c.String(http.StatusOK, replyText)
}

func game(input string, name string) {
	switch input {
	case "/mulai":
		start()
	case "/ganti":
		restart()
	case "/nyerah":
		end()
	case "/hint":
		hint()
	case "/score":
		score()
	default:
		guess(input, name)
	}
}

func start() {
	if isStarted {
		return
	}

	reset()

	qnas := common.GetQNAs()
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(len(qnas))
	qna = qnas[random]

	answers := qna.Answers
	question := qna.Question

	for i := range answers {
		num := i + 1
		correctAnswers = append(correctAnswers, "")
		correctFullAnswers = append(correctFullAnswers, fmt.Sprintf("%d. %s", num, blank))
	}

	correctFullAnswersString := strings.Join(correctFullAnswers[:], "\n")
	replyText = fmt.Sprintf("%s\n%s", question, correctFullAnswersString)
	isStarted = true
}

func restart() {
	if !isStarted {
		return
	}

	reset()

	qnas := common.GetQNAs()
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(len(qnas))
	qna = qnas[random]

	answers := qna.Answers
	question := qna.Question

	for i := range answers {
		num := i + 1
		correctAnswers = append(correctAnswers, "")
		correctFullAnswers = append(correctFullAnswers, fmt.Sprintf("%d. %s", num, blank))
	}

	correctFullAnswersString := strings.Join(correctFullAnswers[:], "\n")
	replyText = fmt.Sprintf("%s\n%s", question, correctFullAnswersString)
	isStarted = true
}

func end() {
	if !isStarted {
		return
	}

	answers := qna.Answers
	question := qna.Question

	for i, answer := range answers {
		num := i + 1
		answerText := answer.Text
		answerScore := answer.Score
		correctFullAnswers[i] = fmt.Sprintf("%d. %s (%d)", num, answerText, answerScore)
	}

	correctFullAnswersString := strings.Join(correctFullAnswers[:], "\n")
	replyText = fmt.Sprintf("%s\n%s\n\n%s", question, correctFullAnswersString, "Yah kok nyerah? Better luck next time yah, semangat :)")
	reset()
}

func guess(input string, name string) {
	if !isStarted {
		return
	}

	for _, correctAnswer := range correctAnswers {
		if strings.EqualFold(input, correctAnswer) {
			return
		}
	}

	answers := qna.Answers
	question := qna.Question

	for i, answer := range answers {
		answerText := answer.Text
		answerScore := answer.Score
		num := i + 1

		if strings.EqualFold(input, answerText) {
			correctAnswers[i] = answerText
			correctFullAnswers[i] = fmt.Sprintf("%d. %s (%d) - %s", num, answerText, answerScore, name)
			correctFullAnswersString := strings.Join(correctFullAnswers[:], "\n")
			replyText = fmt.Sprintf("%s\n%s", question, correctFullAnswersString)
			correct++

			scoreExists := false
			for j, scoreBoard := range scoreBoards {
				if scoreBoard.Name == name {
					scoreBoards[j].Score = scoreBoard.Score + answerScore
					scoreExists = true
				}
			}
			if !scoreExists {
				scoreBoards = append(scoreBoards, models.ScoreBoard{Name: name, Score: answerScore})
			}

			if correct == len(answers) {
				replyText = fmt.Sprintf("%s\n\n%s", replyText, "Selamat kamu udah berhasil jawab semua pertanyaan dengan benar!! Ketik /mulai kalo mau main lagi :)")
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

	for i, correctAnswer := range correctAnswers {
		if strings.EqualFold("", correctAnswer) {
			answerText := qna.Answers[i].Text

			var letter rune

			for j, c := range answerText {
				letter = '_'

				if j == 0 || j == len(answerText)-1 {
					letter = c
				}

				if len(answerText) > 4 && (j == 2) {
					letter = c
				}

				if len(answerText) > 6 && (j == 2 || j == 5) {
					letter = c
				}

				if len(answerText) > 8 && (j == 2 || j == 5 || j == 7) {
					letter = c
				}

				hint = fmt.Sprintf("%s %c", hint, letter)
			}

			replyText = fmt.Sprintf("Ngestuck yah? Ini aku kasih hint buat kamu\nHint:%s\nTetep semangat :)", hint)
			return
		}
	}
}

func score() {
	if len(scoreBoards) == 0 {
		replyText = "Saat ini belum ada yang dapet score nih, kamu jawab dulu dong :D"
		return
	}

	sort.Slice(scoreBoards, func(i, j int) bool {
		if scoreBoards[i].Score > scoreBoards[j].Score {
			return true
		}
		return false
	})

	var scoreFullBoards []string
	for i, scoreBoard := range scoreBoards {
		num := i + 1
		scoreFullBoards = append(scoreFullBoards, fmt.Sprintf("%d. %s - %d", num, scoreBoard.Name, scoreBoard.Score))
	}

	scoreFullBoardsString := strings.Join(scoreFullBoards[:], "\n")
	replyText = fmt.Sprintf("%s\n%s", "Hiyaaa ini score sementara ya, ganbatte!!", scoreFullBoardsString)
}

func reset() {
	correct = 0
	correctAnswers = nil
	correctFullAnswers = nil
	qna = models.QNA{}
	isStarted = false
}

func reply(text string) {
	if text == "" {
		return
	}

	bot := common.GetBot()
	if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(text)).Do(); err != nil {
		log.Println(err)
	}
}
