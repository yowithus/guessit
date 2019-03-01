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
	UserID string `json: "user_id"`
	Name   string `json: "name"`
	Score  int    `json: "score"`
}

type User struct {
	UserID string `json: "user_id"`
	Name   string `json: "name"`
}
