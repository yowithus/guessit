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

	// map userID -> user
	var user = map[string]string{
		"Ud822cc4d292ebff2d209750748424cf9": "Yonatan",
		"U031ffb3a10863fd17494562bf24a9902": "Nicholas",
		"Uc615753e7866a219df34a79cbc9fac4f": "Edw",
		"U9d262243d1ab45795b73cfed1dc21462": "Nathan",
		"U72b299757d1d1e14c4a58b59ff0ef3ef": "Ricky Cibai",
		"U37129fe3076d20a39cd14785897c2cb4": "Septian J",
	}

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

		var name string
		if profile != nil {
			name = profile.DisplayName
		} else {
			name = user[userID]
			if name == "" {
				name = "Guest"
			}
		}

		user := models.User{
			UserID: userID,
			Name:   name,
		}

		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				input := message.Text
				replyText = ""
				game(input, user)
				reply(replyText)
			}
		}
	}

	c.JSON(http.StatusOK, replyText)
}

func PlayTest(c *gin.Context) {
	input := c.Query("input")
	userID := c.Query("userId")
	name := c.Query("name")

	user := models.User{
		UserID: userID,
		Name:   name,
	}

	replyText = ""
	game(input, user)
	c.String(http.StatusOK, replyText)
}

func game(input string, user models.User) {
	switch input {
	case "/mulai":
		start()
	case "/ganti":
		restart()
	case "/nyerah":
		end()
	case "/bantu":
		help()
	case "/nilai":
		score()
	case "/perintah":
		command()
	default:
		guess(input, user)
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
		if correctAnswers[i] == "" {
			correctFullAnswers[i] = fmt.Sprintf("%d. %s (%d)", num, answerText, answerScore)
		}
	}

	correctFullAnswersString := strings.Join(correctFullAnswers[:], "\n")
	replyText = fmt.Sprintf("%s\n%s\n\n%s", question, correctFullAnswersString, "Yah kok nyerah? Better luck next time yah, semangat :)")
	reset()
}

func guess(input string, user models.User) {
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
	userID := user.UserID
	name := user.Name

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
				if scoreBoard.UserID == userID {
					scoreBoards[j].Score = scoreBoard.Score + answerScore
					scoreExists = true
				}
			}
			if !scoreExists {
				scoreBoards = append(scoreBoards, models.ScoreBoard{UserID: userID, Name: name, Score: answerScore})
			}

			if correct == len(answers) {
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
				highestScoreBoard := scoreBoards[0]
				scoreFullBoardsString := strings.Join(scoreFullBoards[:], "\n")
				replyText = fmt.Sprintf("%s\n\nDan score sementara saat ini adalah *jrengjreng*\n%s\n\nWoohoo selamat %s saat ini kamu yang paling unggul lho!", replyText, scoreFullBoardsString, highestScoreBoard.Name)
				reset()
			}
			return
		}
	}
}

func help() {
	if !isStarted {
		return
	}

	var hint = ""
	var incorrectAnswers []string

	for i, correctAnswer := range correctAnswers {
		if strings.EqualFold("", correctAnswer) {
			incorrectAnswers = append(incorrectAnswers, qna.Answers[i].Text)
		}
	}

	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(len(incorrectAnswers))
	answerText := incorrectAnswers[random]

	var letter rune

	for j, c := range answerText {
		if c == ' ' || c == '-' || c == '&' {
			letter = c
		} else {
			letter = '_'
		}

		if j == 0 || j == len(answerText)-1 {
			letter = c
		}
		if len(answerText) > 3 && j == 2 {
			letter = c
		}
		if len(answerText) > 5 && j == 3 {
			letter = c
		}
		if len(answerText) > 7 && j == 6 {
			letter = c
		}
		if len(answerText) > 9 && j == 7 {
			letter = c
		}
		if len(answerText) > 11 && j == 9 {
			letter = c
		}
		if len(answerText) > 13 && j == 12 {
			letter = c
		}
		if len(answerText) > 15 && j == 13 {
			letter = c
		}

		hint = fmt.Sprintf("%s %c", hint, letter)
	}

	replyText = fmt.Sprintf("Ngestuck yah? Aku bantu dikit deh\nHint:%s", hint)
	return
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

	highestScoreBoard := scoreBoards[0]

	scoreFullBoardsString := strings.Join(scoreFullBoards[:], "\n")
	replyText = fmt.Sprintf("Hiyaaa dan score sementara saat ini adalah *jrengjreng*\n%s\n\nWoohoo selamat %s saat ini kamu yang paling unggul lho!", scoreFullBoardsString, highestScoreBoard.Name)
}

func command() {
	commands := []string{
		"1. /mulai - untuk memulai permainan",
		"2. /ganti - kalo kamu bingung sama pertanyaannya dan mau diganti",
		"3. /nyerah - coba berusaha dulu ya, kalo udah mentok baru deh boleh nyerah",
		"4. /bantu - tenang, aku bakal kasih kamu hint kok",
		"5. /nilai - lihat deh siapa yang paling unggul score nya",
	}

	commandsString := strings.Join(commands[:], "\n")
	replyText = fmt.Sprintf("Oya gengs, ini daftar perintah yang tersedia\n%s\n\nSelamat bermain dan enjoy ya :)", commandsString)
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
