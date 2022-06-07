package main

import (
	"golangServer/app/controllers"
	server "golangServer/app/tserver"
	config "golangServer/config"
	manager "golangServer/manager"
)

func main() {
	InitControllers()
	go server.StartServer(config.TCPport)
	manager.Plot()
}

// All Controllers Will Init On Here
func InitControllers() {
	controllers.InitUserController()
	controllers.InitGameController()
}
