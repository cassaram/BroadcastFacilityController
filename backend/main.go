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
	time.Sleep(20000 * time.Millisecond)
	fmt.Println(cerebrum.GetLevels())
	destsJson, _ := json.MarshalIndent(cerebrum.GetDestinations(), "", "    ")
	os.WriteFile("destinations.json", destsJson, 0744)
	//fmt.Println(cerebrum.GetDestinations())
	srcsJson, _ := json.MarshalIndent(cerebrum.GetSources(), "", "    ")
	os.WriteFile("sources.json", srcsJson, 0744)
	//fmt.Println(cerebrum.GetSources())
	xpntsJson, _ := json.MarshalIndent(cerebrum.GetCrosspoints(), "", "    ")
	os.WriteFile("crosspoints.json", xpntsJson, 0744)
	//cerebrum.SetCrosspoint(88, -1, 13, -1)
	cerebrum.LockDestination(88, 1)
	//fmt.Println(cerebrum.GetCrosspoints())
	time.Sleep(time.Second)
	//<-make(chan bool)
}
