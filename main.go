package main

import (
	"github.com/AfazTech/b9m/cli"
	"github.com/AfazTech/logger/v2"
)

func main() {
	logger.SetLogFile("/var/log/b9m.log")
	logger.SetOutput(logger.CONSOLE_AND_FILE)
	cli.StartCLI()
}
