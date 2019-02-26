package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/yowithus/guessit/common"
	"github.com/yowithus/guessit/controllers"
)

func main() {
	common.InitBot()
	common.InitQNA()

	port := fmt.Sprintf(":%s", getPort())

	router := gin.Default()
	router.POST("/callback", controllers.Play)
	router.POST("/playtest", controllers.PlayTest)
	router.Run(port)
}

func getPort() string {
	var port string
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	} else {
		port = "2205"
	}

	return port
}
