package main

import (
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
	time.Sleep(500 * time.Millisecond)
	fmt.Println(cerebrum.GetLevels())
	//fmt.Println(cerebrum.GetDestinations())
	fmt.Println(cerebrum.GetSources())
	time.Sleep(time.Second)
}
