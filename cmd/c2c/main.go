package main

import (
	"flag"

	app "c2c/internal/api"
	"c2c/internal/api/utils"

	log "github.com/sirupsen/ logrus"
)

func main() {

	var confPath = flag.String("conf", "development.json", "conf file path")

	flag.Parse()

	// Formatter := new(log.TextFormatter)
	// Formatter.TimestampFormat = "02-01-2006 15:04:05"
	// Formatter.FullTimestamp = true
	// log.SetFormatter(Formatter)

	config, err := utils.LoadConfiguration(confPath)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s, error:%s", *confPath, err.Error())
		return
	}

	if err = utils.LoadLogger(&config); err != nil {
		log.Fatalf("Can't configure logger %s", err.Error())
		return
	}

	a := app.App{}

	a.Env.ConfigFileName = *confPath

	err = a.Initialize(&config)
	if err != nil {
		log.Fatal(err)
		return
	}

	a.Run()
}
