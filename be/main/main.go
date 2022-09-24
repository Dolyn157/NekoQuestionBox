package main

import (
	"neko-question-box-be/internal/config"
	"neko-question-box-be/internal/database"
	"neko-question-box-be/internal/logger"
	"neko-question-box-be/internal/server"
	"neko-question-box-be/internal/telegram"
)

func main() {
	logger.InitLogger()
	config.InitConfig(false)
	database.InitDB()
	telegram.InitTG()

	r := server.InitServer()
	r.Run()
}
