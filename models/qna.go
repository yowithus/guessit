package models

type QNA struct {
	Question string   `json: "question"`
	Answers  []Answer `json: "answers"`
}

type Answer struct {
	Text  string `json: "text"`
	Score int    `json: "score"`
}

type ScoreBoard struct {
	Name  string `json: "name"`
	Score int    `json: "score"`
}
