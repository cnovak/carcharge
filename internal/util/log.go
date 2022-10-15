/*
Copyright © 2022 Chris Novak <canovak@gmail.com>
*/
package util

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/apex/log/handlers/logfmt"
	"github.com/apex/log/handlers/multi"
)

func InitializeLogs(config Configuration) {
	var handlers []log.Handler

	if config.Logfile != "" {
		file, err := os.OpenFile(config.Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.WithField("logfile", config.Logfile).Fatal("could not open logfile for reading")
		}
		handlers = append(handlers, logfmt.New(file))
	}

	handlers = append(handlers, cli.Default)

	log.SetHandler(multi.New(
		handlers...,
	))
	log.SetLevel(log.DebugLevel)

}
