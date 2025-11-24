package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/cassaram/bfc/backend/config"
	"github.com/cassaram/bfc/backend/router"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	log "github.com/sirupsen/logrus"
)

type APIV1Router struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	ShortName   string `json:"short_name"`
}

type APIV1RouterTableCrosspoint struct {
	DestinationLevelID int  `json:"destination_level_id"`
	SourceID           int  `json:"source_id"`
	SourceLevelID      int  `json:"source_level_id"`
	Locked             bool `json:"locked"`
}

type APIV1RouterTableLine struct {
	ID                  int                          `json:"id"`
	Name                string                       `json:"name"`
	Crosspoints         []APIV1RouterTableCrosspoint `json:"crosspoints"`
	CrosspointsAsString []string                     `json:"crosspoints_as_string"`
}

type APIV1RouterTableValidSource struct {
	SourceID      int `json:"source_id"`
	SourceLevelID int `json:"source_level_id"`
}

type APIV1RouterTableValidSources struct {
	Sources         [][]APIV1RouterTableValidSource `json:"sources"`
	SourcesAsString [][]string                      `json:"sources_as_string"`
}

type APIHandler struct {
	websocketClients []*websocket.Conn
}

func NewAPIHandler() *APIHandler {
	api := APIHandler{}
	return &api
}

func (a *APIHandler) GetServeMux() *http.ServeMux {
	// API V1
	muxV1 := http.NewServeMux()
	muxV1.HandleFunc("/ws", a.APIV1HandleWS)
	muxV1.HandleFunc("GET /routers", a.APIV1HandleRouters)
	muxV1.HandleFunc("GET /routers/{router_id}/table", a.APIV1HandleRouterTable)
	muxV1.HandleFunc("GET /routers/{router_id}/validsources", a.APIV1HandleRouterTableValidSources)
	muxV1.HandleFunc("GET /routers/{router_id}/crosspoints", a.APIV1HandleCrosspoints)
	muxV1.HandleFunc("PUT /routers/{router_id}/crosspoints", a.APIV1HandleCrosspointsPut)
	muxV1.HandleFunc("GET /routers/{router_id}/destinations", a.APIV1HandleDestinations)
	muxV1.HandleFunc("GET /routers/{router_id}/levels", a.APIV1HandleLevels)
	muxV1.HandleFunc("GET /routers/{router_id}/sources", a.APIV1HandleSources)

	// Full API handler
	muxAPI := http.NewServeMux()
	muxAPI.Handle("/v1/", http.StripPrefix("/v1", muxV1))

	return muxAPI
}

func (a *APIHandler) APIV1HandleWS(w http.ResponseWriter, r *http.Request) {
	options := websocket.AcceptOptions{
		InsecureSkipVerify: true,
	}
	wsConn, err := websocket.Accept(w, r, &options)
	if err != nil {
		log.Error("API V1 Websocket Handler: ", err.Error())
		return
	}
	a.websocketClients = append(a.websocketClients, wsConn)
}

func (a *APIHandler) APIV1SendCrosspoint(crosspoint router.Crosspoint) {
	for i, c := range a.websocketClients {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		err := wsjson.Write(ctx, c, crosspoint)
		if err != nil {
			// Websocket connection probably closed
			// Close in case its not already. We can ignore the error since if the other side closed it it doesn't matter
			c.Close(websocket.StatusProtocolError, err.Error())
			a.websocketClients = slices.Delete(a.websocketClients, i, i)
		}
	}
}

func (a *APIHandler) APIV1HandleRouters(w http.ResponseWriter, r *http.Request) {
	rtrs := make([]APIV1Router, 0)
	for _, rtrCfg := range ConfigFile.Routers {
		//rtr := Routers[rtrCfg.ID]
		rtrs = append(rtrs, APIV1Router{
			ID:          rtrCfg.ID,
			DisplayName: rtrCfg.DisplayName,
			ShortName:   rtrCfg.ShortName,
		})
	}
	rtrsBody, err := json.Marshal(rtrs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(rtrsBody)
}

func (a *APIHandler) APIV1HandleDestinations(w http.ResponseWriter, r *http.Request) {
	routerIDStr := r.PathValue("router_id")
	routerID, err := strconv.Atoi(routerIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	router, router_ok := Routers[routerID]
	if !router_ok {
		http.Error(w, fmt.Sprintf("Router ID (%d) not found", routerID), http.StatusNotFound)
		return
	}
	dests := router.GetDestinations()
	destsBody, err := json.Marshal(dests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(destsBody)
}

func (a *APIHandler) APIV1HandleSources(w http.ResponseWriter, r *http.Request) {
	routerIDStr := r.PathValue("router_id")
	routerID, err := strconv.Atoi(routerIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	router, router_ok := Routers[routerID]
	if !router_ok {
		http.Error(w, fmt.Sprintf("Router ID (%d) not found", routerID), http.StatusNotFound)
		return
	}
	dests := router.GetSources()
	destsBody, err := json.Marshal(dests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(destsBody)
}

func (a *APIHandler) APIV1HandleLevels(w http.ResponseWriter, r *http.Request) {
	routerIDStr := r.PathValue("router_id")
	routerID, err := strconv.Atoi(routerIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	router, router_ok := Routers[routerID]
	if !router_ok {
		http.Error(w, fmt.Sprintf("Router ID (%d) not found", routerID), http.StatusNotFound)
		return
	}
	dests := router.GetLevels()
	destsBody, err := json.Marshal(dests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(destsBody)
}

func (a *APIHandler) APIV1HandleCrosspoints(w http.ResponseWriter, r *http.Request) {
	routerIDStr := r.PathValue("router_id")
	routerID, err := strconv.Atoi(routerIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	router, router_ok := Routers[routerID]
	if !router_ok {
		http.Error(w, fmt.Sprintf("Router ID (%d) not found", routerID), http.StatusNotFound)
		return
	}
	dests := router.GetCrosspoints()
	destsBody, err := json.Marshal(dests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(destsBody)
}

func (a *APIHandler) APIV1HandleRouterTable(w http.ResponseWriter, r *http.Request) {
	routerIDStr := r.PathValue("router_id")
	routerID, err := strconv.Atoi(routerIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	router, router_ok := Routers[routerID]
	if !router_ok {
		http.Error(w, fmt.Sprintf("Router ID (%d) not found", routerID), http.StatusNotFound)
		return
	}
	dests := router.GetDestinations()
	crosspoints := router.GetCrosspoints()
	response := make([]APIV1RouterTableLine, len(dests))
	respDestMap := make(map[int]int)
	for i, dest := range dests {
		respDestMap[dest.ID] = i
		response[i] = APIV1RouterTableLine{
			ID:                  dest.ID,
			Name:                dest.Name,
			Crosspoints:         make([]APIV1RouterTableCrosspoint, len(router.GetLevels())),
			CrosspointsAsString: make([]string, len(router.GetLevels())),
		}
	}
	for _, xpt := range crosspoints {
		response[respDestMap[xpt.Destination]].Crosspoints[xpt.DestinationLevel-1] = APIV1RouterTableCrosspoint{
			DestinationLevelID: xpt.DestinationLevel,
			SourceID:           xpt.Source,
			SourceLevelID:      xpt.SourceLevel,
			Locked:             xpt.Locked,
		}
		srcStr := router.GetSource(xpt.Source).Name + "." + router.GetLevel(xpt.SourceLevel).Name
		response[respDestMap[xpt.Destination]].CrosspointsAsString[xpt.DestinationLevel-1] = srcStr
	}

	respBody, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

func (a *APIHandler) APIV1HandleRouterTableValidSources(w http.ResponseWriter, r *http.Request) {
	routerIDStr := r.PathValue("router_id")
	routerID, err := strconv.Atoi(routerIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	router, router_ok := Routers[routerID]
	if !router_ok {
		http.Error(w, fmt.Sprintf("Router ID (%d) not found", routerID), http.StatusNotFound)
		return
	}
	sources := router.GetSources()
	levels := router.GetLevels()
	levelStrings := make([][]string, len(levels))
	levelSources := make([][]APIV1RouterTableValidSource, len(levels))
	for i := 0; i < len(levels); i++ {
		levelStrings[i] = make([]string, 0)
		levelSources[i] = make([]APIV1RouterTableValidSource, 0)
	}

	for _, src := range sources {
		for _, lvl := range src.Levels {
			srcNameStr := src.Name + "." + router.GetLevel(lvl).Name
			levelStrings[lvl-1] = append(levelStrings[lvl-1], srcNameStr)
			srcAPI := APIV1RouterTableValidSource{
				SourceID:      src.ID,
				SourceLevelID: lvl,
			}
			levelSources[lvl-1] = append(levelSources[lvl-1], srcAPI)
		}
	}

	var routerConfig config.RouterConfig
	for _, rtrCfg := range ConfigFile.Routers {
		if rtrCfg.ID == routerID {
			routerConfig = rtrCfg
			break
		}
	}

	response := APIV1RouterTableValidSources{
		Sources:         make([][]APIV1RouterTableValidSource, len(levels)),
		SourcesAsString: make([][]string, len(levels)),
	}

	for i := 0; i < len(levels); i++ {
		response.SourcesAsString[i] = levelStrings[i]
		response.Sources[i] = levelSources[i]
		altLevels := routerConfig.AlternateLevels[strconv.Itoa(i+1)]
		for _, lvl := range altLevels {
			if lvl-1 == i {
				continue
			}
			response.SourcesAsString[i] = append(response.SourcesAsString[i], levelStrings[lvl-1]...)
			response.Sources[i] = append(response.Sources[i], levelSources[lvl-1]...)
		}
	}

	respBody, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

func (a *APIHandler) APIV1HandleCrosspointsPut(w http.ResponseWriter, r *http.Request) {
	routerIDStr := r.PathValue("router_id")
	routerID, err := strconv.Atoi(routerIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	router, router_ok := Routers[routerID]
	if !router_ok {
		http.Error(w, fmt.Sprintf("Router ID (%d) not found", routerID), http.StatusNotFound)
		return
	}
	body := make(map[string]int)
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Error parsing body "+err.Error(), http.StatusBadRequest)
		return
	}
	destID, destID_ok := body["destination_id"]
	destLevelID, destLevelID_ok := body["destination_level_id"]
	srcID, srcID_ok := body["source_id"]
	srcLevelID, srcLevelID_ok := body["source_level_id"]
	if !destID_ok || !destLevelID_ok || !srcID_ok || !srcLevelID_ok {
		http.Error(w, "Error parsing body ", http.StatusBadRequest)
		return
	}
	err = router.SetCrosspoint(destID, destLevelID, srcID, srcLevelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
