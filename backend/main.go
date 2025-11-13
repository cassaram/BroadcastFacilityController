package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/cassaram/bfc/backend/config"
	"github.com/cassaram/bfc/backend/router"
	"github.com/cassaram/bfc/backend/router/harrislrc"
	log "github.com/sirupsen/logrus"
)

var Routers map[int]router.Router

func main() {
	log.SetOutput(os.Stdout)
	//log.SetLevel(log.DebugLevel)

	Routers = make(map[int]router.Router)

	// Load config file
	var configFile config.ConfigFile
	configFileBytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(configFileBytes, &configFile)

	// Handle logging
	switch strings.ToLower(configFile.LogLevel) {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	}

	// Handle HTTP Server
	go HandleHTTP()

	// Handle Routers
	for _, rtrCfg := range configFile.Routers {
		switch strings.ToLower(rtrCfg.Type) {
		case "harrislrc":
			rtr := harrislrc.HarrisLRCRouter{}
			rtr.Init(rtrCfg.Config)
			Routers[rtrCfg.ID] = router.Router(&rtr)
		default:
			log.Fatal("Invalid router type: ", rtrCfg.Type)
		}
	}

	// Start Routers
	for _, rtr := range Routers {
		rtr.Start()
	}

	// Run forever
	<-make(chan bool)
}

func HandleHTTP() {
	rootMux := http.NewServeMux()
	api := NewAPIHandler()
	apiMux := api.GetServeMux()
	rootMux.Handle("/api/", http.StripPrefix("/api", apiMux))

	go func() { log.Fatal(http.ListenAndServe(":80", rootMux)) }()
	go func() { log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", rootMux)) }()
}
