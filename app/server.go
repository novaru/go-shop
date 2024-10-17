package app

import (
	"flag"
	"github.com/novaru/go-shop/app/config"
	"github.com/novaru/go-shop/app/controllers"
)

func Run() {
	server := controllers.Server{}
	env := &config.Env{}
	env.Init()

	flag.Parse()
	arg := flag.Arg(0)

	if arg != "" {
		server.InitCommands(*env)
	} else {
		server.Initialize(*env)
		server.Run(":" + env.AppPort)
	}
}
