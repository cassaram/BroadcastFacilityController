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
	Sources               map[int]router.Source
	SourcesMutex          sync.Mutex
	SourcesName           map[string]int // Stores Name -> ID mapping
	SourcesNameMutex      sync.Mutex
	Crosspoints           map[int]map[int]router.Crosspoint // Destination -> Level -> Crosspoint
	CrosspointMutex       sync.Mutex
	crosspointNotify      chan router.Crosspoint
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
	r.stop = make(chan bool)
	r.replyMessages = make(chan lrcMessage, 100) // Buffered to add some level of async capabilitiy between listener and handler
	r.receiverReady = 0
	r.Levels = make(map[int]router.Level)
	r.LevelsName = make(map[string]int)
	r.Destinations = make(map[int]router.Destination)
	r.DestinationsName = make(map[string]int)
	r.Sources = make(map[int]router.Source)
	r.SourcesName = make(map[string]int)
	r.Crosspoints = make(map[int]map[int]router.Crosspoint)
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

	// Get initial configs
	r.getConfig()
}

func (r *HarrisLRCRouter) getConfig() {
	log.Infoln("Harris LRC Router: Fetching full configuration")
	go func() {
		r.sendCommand("~CHANNELS?\\")
		time.Sleep(10 * time.Millisecond)
		r.sendCommand("~DEST?Q${NAME,CHANNELS}\\")
		time.Sleep(10 * time.Millisecond)
		r.sendCommand("~SRC?Q${NAME,CHANNELS}\\")
		time.Sleep(10 * time.Millisecond)
		r.sendCommand("~XPOINT?\\")
		time.Sleep(10 * time.Millisecond)
		r.sendCommand("~LOCK?\\")
	}()
}

func (r *HarrisLRCRouter) Stop() {
	r.stop <- true
	err := r.conn.Close()
	if err != nil {
		log.Error("Harris LRC Router: ", err.Error())
	}
}

func (r *HarrisLRCRouter) SetCrosspointNotifyChannel(c chan router.Crosspoint) {
	r.crosspointNotify = c
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
			case "DBCHANGE":
				// Router reconfigured itself. Re-request all sources, destinations, channels, etc.
				r.getConfig()
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
							// Also setup crosspoints for destination
							r.CrosspointMutex.Lock()
							r.Crosspoints[dest.ID] = make(map[int]router.Crosspoint)
							r.CrosspointMutex.Unlock()
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
							continue
						}
						r.DestinationsMutex.Lock()
						dest, dest_exists := r.Destinations[destID]
						r.DestinationsMutex.Unlock()
						if !dest_exists {
							dest = router.Destination{
								ID:     destID,
								Name:   "",
								Levels: make([]int, 0),
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
							foundlvl := false
							for _, testlvl := range dest.Levels {
								if testlvl == lvlID {
									// do nothing
									foundlvl = true
									break
								}
							}
							if !foundlvl {
								dest.Levels = append(dest.Levels, lvlID)
							}
						}
						slices.SortFunc(dest.Levels, func(a int, b int) int {
							return cmp.Compare(a, b)
						})

						r.DestinationsMutex.Lock()
						r.Destinations[dest.ID] = dest
						r.DestinationsMutex.Unlock()
					}
				}
			case "SRC":
				switch msg.op {
				case _QUERYRESP:
					_, arg_name := msg.args["NAME"]
					_, arg_I := msg.args["I"]
					_, arg_channels := msg.args["CHANNELS"]
					if arg_I && arg_name {
						// Source name report
						id, err := strconv.Atoi(msg.args["I"].values[0])
						if err != nil {
							log.Errorln("Harris LRC Router: Error parsing argument", err.Error())
							continue
						}
						r.SourcesMutex.Lock()
						src, src_exists := r.Sources[id]
						r.SourcesMutex.Unlock()

						if !src_exists {
							src = router.Source{
								ID:     id,
								Name:   msg.args["NAME"].values[0],
								Levels: make([]int, 0),
							}
							r.SourcesMutex.Lock()
							r.Sources[id] = src
							r.SourcesMutex.Unlock()
							r.SourcesNameMutex.Lock()
							r.SourcesName[src.Name] = id
							r.SourcesNameMutex.Unlock()
						} else {
							oldName := src.Name
							src.Name = msg.args["NAME"].values[0]
							r.SourcesMutex.Lock()
							r.Sources[id] = src
							r.SourcesMutex.Unlock()
							r.SourcesNameMutex.Lock()
							delete(r.SourcesName, oldName)
							r.SourcesName[src.Name] = id
							r.SourcesNameMutex.Unlock()
						}
					}
					if arg_I && arg_channels {
						// Channel report
						srcId := -1
						switch msg.args["I"].argType {
						case _NUMERIC:
							srcIdTemp, err := strconv.Atoi(msg.args["I"].values[0])
							if err != nil {
								log.Errorln("Harris LRC Router: Error ", err)
								continue
							}
							srcId = srcIdTemp
						case _STRING:
							r.SourcesNameMutex.Lock()
							srcIdTemp, ok := r.SourcesName[msg.args["I"].values[0]]
							r.SourcesNameMutex.Unlock()
							if !ok {
								log.Errorln("Harris LRC Router: Error parsing message ", msg)
								continue
							}
							srcId = srcIdTemp
						}
						if srcId < 0 {
							log.Errorln("Harris LRC Router: Error parsing message ", msg)
							continue
						}

						r.SourcesMutex.Lock()
						src, ok := r.Sources[srcId]
						r.SourcesMutex.Unlock()
						if !ok {
							log.Errorln("Harris LRC Router: Source does not exist ", srcId)
						}

						for _, lvlStr := range msg.args["CHANNELS"].values {
							lvlId := -1
							switch msg.args["CHANNELS"].argType {
							case _NUMERIC:
								lvlIdTemp, err := strconv.Atoi(lvlStr)
								if err != nil {
									log.Errorln("Harris LRC Router: Error ", err)
									continue
								}
								lvlId = lvlIdTemp
							case _STRING:
								r.LevelsNameMutex.Lock()
								lvlIdTemp, ok := r.LevelsName[lvlStr]
								r.LevelsNameMutex.Unlock()
								if !ok {
									log.Errorln("Harris LRC Router: Error parsing message ", msg)
									continue
								}
								lvlId = lvlIdTemp
							}
							if lvlId < 0 {
								log.Errorln("Harris LRC Router: Error parsing message ", msg)
								continue
							}
							src.Levels = append(src.Levels, lvlId)
							slices.SortFunc(src.Levels, func(a int, b int) int {
								return cmp.Compare(a, b)
							})
						}
						r.SourcesMutex.Lock()
						r.Sources[srcId] = src
						r.SourcesMutex.Unlock()
					}
				}
			case "XPOINT":
				if msg.op == _QUERYRESP || msg.op == _CHANGENOTIFY {
					// Crosspoint update
					arg_d, arg_d_ok := msg.args["D"]
					arg_s, arg_s_ok := msg.args["S"]
					if arg_d_ok && arg_d.argType != _NUMERIC {
						// Ignore reports that aren't numeric
						// Reports are sent twice as numeric and string based
						// We can ignore the strings to save error-handling and processing time
						continue
					}
					if arg_d_ok && arg_s_ok {
						// Crosspoint report
						destID := -1
						destLvlID := -1
						srcID := -1
						srcLvlID := -1
						followMode := false

						destStrs := strings.Split(arg_d.values[0], ".")
						if len(destStrs) == 1 {
							// Follow mode
							followMode = true
						} else if len(destStrs) < 2 {
							log.Errorln("Harris LRC Router: Error parsing message ", msg)
							continue
						} else {
							// Normal breakaway mode
							var err error
							destID, err = strconv.Atoi(destStrs[0])
							if err != nil {
								log.Errorln("Harris LRC Router: Error parsing message", msg)
								continue
							}
							destLvlID, err = strconv.Atoi(destStrs[1])
							if err != nil {
								log.Errorln("Harris LRC Router: Error parsing message", msg)
								continue
							}
						}

						if len(arg_s.values[0]) == 0 {
							// Channel is in breakaway
							// There should be other messages that report the individual level changes
							// We can just ignore this then
							continue
						}
						srcStrs := strings.Split(arg_s.values[0], ".")
						if len(srcStrs) == 1 && len(destStrs) == 1 {
							// Follow mode. All destination levels pull from source levels.
							followMode = true
						} else if len(srcStrs) < 2 {
							// Error occured
							log.Errorln("Harris LRC Router: Error parsing message ", msg)
							continue
						} else {
							// Breakaway mode
							var err error
							srcID, err = strconv.Atoi(srcStrs[0])
							if err != nil {
								log.Errorln("Harris LRC Router: Error parsing message", msg)
								continue
							}
							srcLvlID, err = strconv.Atoi(srcStrs[1])
							if err != nil {
								log.Errorln("Harris LRC Router: Error parsing message", msg)
								continue
							}
						}

						r.CrosspointMutex.Lock()
						destCrosspoints, ok := r.Crosspoints[destID]
						r.CrosspointMutex.Unlock()
						if !ok {
							destCrosspoints = make(map[int]router.Crosspoint)
						}

						if followMode {
							// Route all destination crosspoints to be from single source
							r.DestinationsMutex.Lock()
							dest := r.Destinations[destID]
							r.DestinationsMutex.Unlock()
							for _, followDestLevelID := range dest.Levels {
								lvlCrosspoint, lvlCrosspointok := destCrosspoints[followDestLevelID]
								if !lvlCrosspointok {
									lvlCrosspoint = router.Crosspoint{
										Destination:      destID,
										DestinationLevel: followDestLevelID,
										Source:           srcID,
										SourceLevel:      followDestLevelID,
										Locked:           false,
									}
								} else {
									lvlCrosspoint.Source = srcID
									lvlCrosspoint.SourceLevel = followDestLevelID
								}
								destCrosspoints[followDestLevelID] = lvlCrosspoint
								if r.crosspointNotify != nil {
									r.crosspointNotify <- lvlCrosspoint
								}
							}
						} else {
							// Breakaway mode
							crosspoint, ok := destCrosspoints[destLvlID]
							if !ok {
								crosspoint = router.Crosspoint{
									Destination:      destID,
									DestinationLevel: destLvlID,
									Source:           srcID,
									SourceLevel:      srcLvlID,
									Locked:           false,
								}
							} else {
								crosspoint.Source = srcID
								crosspoint.SourceLevel = srcLvlID
							}
							destCrosspoints[destLvlID] = crosspoint
							if r.crosspointNotify != nil {
								r.crosspointNotify <- crosspoint
							}
						}
						r.CrosspointMutex.Lock()
						r.Crosspoints[destID] = destCrosspoints
						r.CrosspointMutex.Unlock()
					}
				}
			case "LOCK":
				if msg.op == _CHANGENOTIFY || msg.op == _QUERYRESP {
					arg_d, arg_d_ok := msg.args["D"]
					arg_v, arg_v_ok := msg.args["V"]

					if arg_d_ok && arg_v_ok {
						destID := -1
						switch arg_d.argType {
						case _NUMERIC:
							var err error
							destID, err = strconv.Atoi(arg_d.values[0])
							if err != nil {
								log.Errorln("Harris LRC Router: Error parsing message ", msg, err)
								continue
							}
						case _STRING:
							r.DestinationsNameMutex.Lock()
							var destIDOkay bool
							destID, destIDOkay = r.DestinationsName[arg_d.values[0]]
							r.DestinationsNameMutex.Unlock()
							if !destIDOkay {
								log.Errorln("Harris LRC Router: Error parsing message ", msg)
								continue
							}
						}
						locked := arg_v.values[0] == "OFF"

						// Update crosspoints for destination
						r.CrosspointMutex.Lock()
						destCrosspoints := r.Crosspoints[destID]
						r.CrosspointMutex.Unlock()
						for lvlID, crosspoint := range destCrosspoints {
							crosspoint.Locked = locked
							destCrosspoints[lvlID] = crosspoint
						}
						r.CrosspointMutex.Lock()
						r.Crosspoints[destID] = destCrosspoints
						r.CrosspointMutex.Unlock()
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
	r.DestinationsMutex.Unlock()
	slices.SortFunc(dests, func(a router.Destination, b router.Destination) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return dests
}

func (r *HarrisLRCRouter) GetSources() []router.Source {
	r.SourcesMutex.Lock()
	srcs := maps.Values(r.Sources)
	r.SourcesMutex.Unlock()
	slices.SortFunc(srcs, func(a router.Source, b router.Source) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return srcs
}

func (r *HarrisLRCRouter) GetCrosspoints() []router.Crosspoint {
	crosspoints := make([]router.Crosspoint, 0)
	r.CrosspointMutex.Lock()
	for _, destCrosspoint := range r.Crosspoints {
		destXpntsSlice := maps.Values(destCrosspoint)
		crosspoints = append(crosspoints, destXpntsSlice...)
	}
	r.CrosspointMutex.Unlock()
	slices.SortFunc(crosspoints, func(a router.Crosspoint, b router.Crosspoint) int {
		destCmp := cmp.Compare(a.Destination, b.Destination)
		if destCmp != 0 {
			return destCmp
		}
		return cmp.Compare(a.DestinationLevel, b.DestinationLevel)
	})
	return crosspoints
}
