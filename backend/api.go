package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type APIV1Router struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	ShortName   string `json:"short_name"`
}

type APIHandler struct {
}

func NewAPIHandler() *APIHandler {
	api := APIHandler{}
	return &api
}

func (a *APIHandler) GetServeMux() *http.ServeMux {
	// API V1
	muxV1 := http.NewServeMux()
	muxV1.HandleFunc("GET /routers", APIV1HandleRouters)
	muxV1.HandleFunc("GET /routers/{router_id}/crosspoints", APIV1HandleCrosspoints)
	muxV1.HandleFunc("GET /routers/{router_id}/destinations", APIV1HandleDestinations)
	muxV1.HandleFunc("GET /routers/{router_id}/levels", APIV1HandleLevels)
	muxV1.HandleFunc("GET /routers/{router_id}/sources", APIV1HandleSources)

	// Full API handler
	muxAPI := http.NewServeMux()
	muxAPI.Handle("/v1/", http.StripPrefix("/v1", muxV1))

	return muxAPI
}

func APIV1HandleRouters(w http.ResponseWriter, r *http.Request) {
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

func APIV1HandleDestinations(w http.ResponseWriter, r *http.Request) {
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

func APIV1HandleSources(w http.ResponseWriter, r *http.Request) {
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

func APIV1HandleLevels(w http.ResponseWriter, r *http.Request) {
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

func APIV1HandleCrosspoints(w http.ResponseWriter, r *http.Request) {
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
