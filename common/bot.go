package common

import "github.com/line/line-bot-sdk-go/linebot"

var bot *linebot.Client
var err error

func InitBot() {
	channelSecret := "022bbabe5172412a7a1b8c92cc293e7c"
	channelAccessToken := "0UX2kP8ikKnu90ALoDwxTLUabQtt7S/c+SC1OcKY3I6IVtwWX0hZ9tvm5tSr6zxUVgqlZH+MiVCj40V55bJEqpajRetQWnfI2xrm3fumk3rbFGqwUC/7Yd+mAhwf4btZ81K91h4gOzjeb6ep3/NmWQdB04t89/1O/w1cDnyilFU="
	bot, err = linebot.New(channelSecret, channelAccessToken)
	if err != nil {
		panic(err.Error())
	}
}

func GetBot() *linebot.Client {
	return bot
}
