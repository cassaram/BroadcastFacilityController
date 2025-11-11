package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cassaram/bfc/backend/router/harrislrc"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetOutput(os.Stdout)
	//log.SetLevel(log.DebugLevel)

	cerebrum := harrislrc.HarrisLRCRouter{}
	cerebrum.Init(harrislrc.HarrisLRCRouterConfig{
		Hostname: "10.10.60.10",
		Port:     52116,
	})
	cerebrum.Start()
	time.Sleep(5000 * time.Millisecond)
	fmt.Println(cerebrum.GetLevels())
	//fmt.Println(cerebrum.GetDestinations())
	srcsJson, _ := json.MarshalIndent(cerebrum.GetSources(), "", "    ")
	os.WriteFile("sources.json", srcsJson, 0744)
	fmt.Println(cerebrum.GetSources())
	//fmt.Println(cerebrum.GetCrosspoints())
	//time.Sleep(time.Second)
	//<-make(chan bool)
}
