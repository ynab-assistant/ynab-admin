package main

import (
	"github.com/ynab-assistant/ynab-admin/pkg/app"
	"log"
	"os"
)

const configPath = "configs/main"

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	logger := log.New(os.Stdout, "BOT : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	if err := app.Run(logger, configPath, build); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}
