package harrislrc

import (
	"cmp"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"

	"github.com/cassaram/bfc/backend/router"
)

type HarrisLRCRouter struct {
	Hostname              string
	Port                  uint16
	conn                  net.Conn
	crosspointNotify      chan router.Crosspoint
	stop                  chan bool
	replyMessages         chan lrcMessage
	receiverReady         int
	Levels                map[int]router.Level
	LevelsMutex           sync.Mutex
	LevelsName            map[string]int // Stores Name -> ID mapping
	LevelsNameMutex       sync.Mutex
	Destinations          map[int]router.Destination
	DestinationsMutex     sync.Mutex
	DestinationsName      map[string]int // Stores Name -> ID mapping
	DestinationsNameMutex sync.Mutex
}

type HarrisLRCRouterConfig struct {
	Hostname string
	Port     int
}

func (r *HarrisLRCRouter) Init(conf HarrisLRCRouterConfig) {
	// Error handling
	if conf.Port < 0 || conf.Port > 0xFFFF {
		log.Error(fmt.Sprintf("Harris LRC Router: Port (%v) out of range", conf.Port))
		return
	}

	r.Hostname = conf.Hostname
	r.Port = uint16(conf.Port)
	r.conn = nil
	r.crosspointNotify = make(chan router.Crosspoint)
	r.stop = make(chan bool)
	r.replyMessages = make(chan lrcMessage)
	r.receiverReady = 0
	r.Levels = make(map[int]router.Level)
	r.LevelsName = make(map[string]int)
	r.Destinations = make(map[int]router.Destination)
	r.DestinationsName = make(map[string]int)
}

func (r *HarrisLRCRouter) Start() {
	portStr := strconv.FormatUint(uint64(r.Port), 10)
	conn, err := net.Dial("tcp", r.Hostname+":"+portStr)
	if err != nil {
		log.Error("Harris LRC Router:", err.Error())
		conn.Close()
		return
	}
	log.Info("Harris LRC Router: Connected to ", r.Hostname+":"+portStr)
	r.conn = conn

	go r.replyHandler()
	go r.replyListener()

	go func() {
		r.sendCommand("~CHANNELS?\\")
		time.Sleep(100 * time.Millisecond)
		r.sendCommand("~DEST?Q${NAME,CHANNELS}\\")
		time.Sleep(100 * time.Millisecond)
		//r.sendCommand("~DEST?Q${CHANNELS}\\")
	}()

}

func (r *HarrisLRCRouter) Stop() {
	r.stop <- true
	err := r.conn.Close()
	if err != nil {
		log.Error("Harris LRC Router: ", err.Error())
	}
}

func (r *HarrisLRCRouter) GetCrosspointNotifyChannel() chan router.Crosspoint {
	return r.crosspointNotify
}

func (r *HarrisLRCRouter) sendCommand(cmd string) error {
	for r.receiverReady < 2 {
		time.Sleep(1000)
	}
	cmdBytes := []byte(cmd)
	_, err := r.conn.Write(cmdBytes)
	if err != nil {
		return err
	}
	return nil
}

func (r *HarrisLRCRouter) replyListener() {
	shortBuffer := make([]byte, 1500)
	largeBuffer := ""

	r.receiverReady++
	for {
		select {
		case <-r.stop:
			return
		default:
			n, err := r.conn.Read(shortBuffer)
			if err != nil && err.Error() == "EOF" {
				log.Info("Harris LRC Router: Connection closed by remote")
				r.stop <- true
				return
			} else if err != nil {
				log.Error("Harris LRC Router:", err.Error())
				r.stop <- true
				err := r.conn.Close()
				if err != nil {
					log.Error("Harris LRC Router:", err.Error())
				}
				return
			}
			// Insert into large buffer
			largeBuffer += string(shortBuffer[:n])
			largeBuffer = strings.ReplaceAll(largeBuffer, "\r", "")
			largeBuffer = strings.ReplaceAll(largeBuffer, "\n", "")
			// Process large buffer
			for {
				if len(largeBuffer) == 0 {
					break
				}
				msgStart := strings.Index(largeBuffer, "~")
				msgEnd := strings.Index(largeBuffer, "\\")
				if msgEnd < 0 {
					log.Debug("Harris LRC Router: Buffer size: ", len(largeBuffer), " Contents: ", string(largeBuffer))
					break
				}
				msgStr := largeBuffer[msgStart : msgEnd+1]
				if msgEnd == len(largeBuffer)-1 {
					largeBuffer = largeBuffer[:msgStart]
				} else {
					largeBuffer = largeBuffer[:msgStart] + largeBuffer[msgEnd+1:]
				}
				log.Debug("Harris LRC Router: Received", msgStr)
				msg := lrcMessageFromString(msgStr)
				r.replyMessages <- msg
			}
		}
	}
}

func (r *HarrisLRCRouter) replyHandler() {
	r.receiverReady++
	for {
		select {
		case <-r.stop:
			return
		case msg := <-r.replyMessages:
			log.Debug("Harris LRC Router: Parsed ", msg)
			switch msg.msgType {
			case "CHANNELS":
				switch msg.op {
				case _QUERYRESP:
					// List of levels
					for i := range msg.args["I"].values {
						id, err := strconv.Atoi(msg.args["I"].values[i])
						if err != nil {
							log.Error("Harris LRC Router: Error parsing argument", err.Error())
							continue
						}

						foundLvl := false
						r.LevelsMutex.Lock()
						for key, val := range r.Levels {
							if val.ID == id {
								foundLvl = true
								r.LevelsNameMutex.Lock()
								delete(r.LevelsName, val.Name)
								r.LevelsNameMutex.Unlock()
								val.Name = msg.args["NAME"].values[i]
								r.Levels[key] = val
								r.LevelsNameMutex.Lock()
								r.LevelsName[val.Name] = val.ID
								r.LevelsNameMutex.Unlock()
								break
							}
						}
						r.LevelsMutex.Unlock()
						if !foundLvl {
							lvl := router.Level{
								ID:   id,
								Name: msg.args["NAME"].values[i],
							}
							r.LevelsMutex.Lock()
							r.Levels[lvl.ID] = lvl
							r.LevelsMutex.Unlock()
							r.LevelsNameMutex.Lock()
							r.LevelsName[lvl.Name] = lvl.ID
							r.LevelsNameMutex.Unlock()
						}

					}
				}
			case "DEST":
				switch msg.op {
				case _QUERYRESP:
					// A destination
					_, arg_name := msg.args["NAME"]
					_, arg_I := msg.args["I"]
					_, arg_channels := msg.args["CHANNELS"]
					_, arg_count := msg.args["COUNT"]
					if arg_count {

					}
					if arg_I && arg_name {
						// Name reports
						id, err := strconv.Atoi(msg.args["I"].values[0])
						if err != nil {
							log.Error("Harris LRC Router: Error parsing argument", err.Error())
							continue
						}
						r.DestinationsMutex.Lock()
						dest, dest_exists := r.Destinations[id]
						r.DestinationsMutex.Unlock()
						if dest_exists {
							r.DestinationsNameMutex.Lock()
							delete(r.DestinationsName, dest.Name)
							r.DestinationsNameMutex.Unlock()
							dest.Name = msg.args["NAME"].values[0]
							r.DestinationsMutex.Lock()
							r.Destinations[dest.ID] = dest
							r.DestinationsMutex.Unlock()
							r.DestinationsNameMutex.Lock()
							r.DestinationsName[dest.Name] = dest.ID
							r.DestinationsNameMutex.Unlock()
						} else {
							dest.ID = id
							dest.Name = msg.args["NAME"].values[0]
							r.DestinationsMutex.Lock()
							r.Destinations[dest.ID] = dest
							r.DestinationsMutex.Unlock()
							r.DestinationsNameMutex.Lock()
							r.DestinationsName[dest.Name] = dest.ID
							r.DestinationsNameMutex.Unlock()
						}
					}
					if arg_I && arg_channels {
						// Supported level reports
						destID := -1
						switch msg.args["I"].argType {
						case _NUMERIC:
							destID, _ = strconv.Atoi(msg.args["I"].values[0])
						case _STRING:
							r.DestinationsNameMutex.Lock()
							destID = r.DestinationsName[msg.args["I"].values[0]]
							r.DestinationsNameMutex.Unlock()
						case _UTF:
							r.DestinationsNameMutex.Lock()
							destID = r.DestinationsName[msg.args["I"].values[0]]
							r.DestinationsNameMutex.Unlock()
						}
						if destID < 0 {
							log.Error("Harris LRC Router: Error parsing destination ID ", msg.args["I"])

						}
						r.DestinationsMutex.Lock()
						dest, dest_exists := r.Destinations[destID]
						r.DestinationsMutex.Unlock()
						if !dest_exists {
							dest = router.Destination{
								ID:     destID,
								Name:   "",
								Levels: make([]router.Level, 0),
							}
						}

						for _, lvlstr := range msg.args["CHANNELS"].values {
							lvlID := -1
							switch msg.args["CHANNELS"].argType {
							case _NUMERIC:
								lvlID, _ = strconv.Atoi(lvlstr)
							case _STRING:
								r.LevelsNameMutex.Lock()
								lvlID = r.LevelsName[lvlstr]
								r.LevelsNameMutex.Unlock()
							case _UTF:
								r.LevelsNameMutex.Lock()
								lvlID = r.LevelsName[lvlstr]
								r.LevelsNameMutex.Unlock()
							}
							r.LevelsMutex.Lock()
							lvl := r.Levels[lvlID]
							r.LevelsMutex.Unlock()
							foundlvl := false
							for _, testlvl := range dest.Levels {
								if testlvl.ID == lvl.ID {
									// do nothing
									foundlvl = true
									break
								}
							}
							if !foundlvl {
								dest.Levels = append(dest.Levels, lvl)
							}
						}
						slices.SortFunc(dest.Levels, func(a router.Level, b router.Level) int {
							return cmp.Compare(a.ID, b.ID)
						})

						r.DestinationsMutex.Lock()
						r.Destinations[dest.ID] = dest
						r.DestinationsMutex.Unlock()
					}
				}
			}

		}
	}
}

func (r *HarrisLRCRouter) GetLevels() []router.Level {
	r.LevelsMutex.Lock()
	levels := maps.Values(r.Levels)
	r.LevelsMutex.Unlock()
	slices.SortFunc(levels, func(a router.Level, b router.Level) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return levels
}

func (r *HarrisLRCRouter) GetDestinations() []router.Destination {
	r.DestinationsMutex.Lock()
	dests := maps.Values(r.Destinations)
	fmt.Println(len(r.Destinations))
	r.DestinationsMutex.Unlock()
	slices.SortFunc(dests, func(a router.Destination, b router.Destination) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return dests
}
