package common

import (
	"encoding/json"
	"io/ioutil"

	"github.com/yowithus/guessit/models"
)

var qnas []models.QNA

func InitQNA() {
	qnaJson, err := ioutil.ReadFile("./models/qna.json")
	if err != nil {
		panic(err.Error())
	}

	err = json.Unmarshal(qnaJson, &qnas)
	if err != nil {
		panic(err.Error())
	}
}

func GetQNAs() []models.QNA {
	return qnas
}
