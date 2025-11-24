package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/cassaram/bfc/backend/config"
	"github.com/cassaram/bfc/backend/router"
	"github.com/cassaram/bfc/backend/router/harrislrc"
	"github.com/coder/websocket"
	log "github.com/sirupsen/logrus"
)

var Routers map[int]router.Router
var ConfigFile config.ConfigFile
var WebsocketConnections []*websocket.Conn
var API APIHandler

func main() {
	log.SetOutput(os.Stdout)
	API = *NewAPIHandler()

	Routers = make(map[int]router.Router)
	WebsocketConnections = make([]*websocket.Conn, 0)

	// Load config file
	configFileBytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(configFileBytes, &ConfigFile)

	// Handle logging
	switch strings.ToLower(ConfigFile.LogLevel) {
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
	for _, rtrCfg := range ConfigFile.Routers {
		switch strings.ToLower(rtrCfg.Type) {
		case "harrislrc":
			rtr := harrislrc.HarrisLRCRouter{}
			rtr.Init(rtrCfg.Config)
			Routers[rtrCfg.ID] = router.Router(&rtr)
		default:
			log.Fatal("Invalid router type: ", rtrCfg.Type)
		}
		Routers[rtrCfg.ID].SetCrosspointNotifyFunc(API.APIV1SendCrosspoint)
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
	apiMux := API.GetServeMux()
	rootMux.Handle("/api/", http.StripPrefix("/api", apiMux))

	go func() {
		log.Fatal(http.ListenAndServe(":80", httpMiddlewareCors(rootMux)))
	}()
	go func() {
		log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", httpMiddlewareCors(rootMux)))
	}()
}

func httpMiddlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			// Handle CORS Preflight
			return
		}
		next.ServeHTTP(w, r)
	})
}
