import { inject, Injectable } from "@angular/core";
import { HttpClient, HttpHeaders, HttpParams } from "@angular/common/http";
import { Observable } from "rxjs";
import { Router } from "./models/router";
import { RouterSource } from "./models/routerSource";
import { RouterDestination } from "./models/routerDestination";
import { RouterCrosspoint } from "./models/routerCrosspoint";
import { RouterLevel } from "./models/routerLevel";
import { RouterTableLine } from "./models/routertableline";
import { RouterTableValidSources } from "./models/routertablevalidsources";
import { FetchBackend } from "@angular/common/http";
import { webSocket, WebSocketSubject } from "rxjs/webSocket";

const httpOptions = {
    headers: new HttpHeaders({
        'Content-Type': 'application/json',
    })
};

@Injectable({
    providedIn: 'root'
  })
export class BackendService {
    private http = inject(HttpClient);
    private socket_crosspoints$: WebSocketSubject<RouterCrosspoint>;
    
    constructor(
        //private http: HttpClient
    ) {
        this.socket_crosspoints$ = webSocket(import.meta.env.NG_APP_BACKEND_API_URL + '/api/v1/ws')
    }

    // Websocket API
    getWebsocketCrosspoints(): Observable<any> {
        return this.socket_crosspoints$.asObservable();
    }

    closeWebsocketConnection(): void {
        this.socket_crosspoints$.complete();
    }

    // HTTP API

    getRouters(): Observable<Router[]> {
        return this.http.get<Router[]>(import.meta.env.NG_APP_BACKEND_API_URL + '/api/v1/routers', {responseType: 'json'});
    }

    getRouterSources(rtrid: number): Observable<RouterSource[]> {
        return this.http.get<RouterSource[]>(import.meta.env.NG_APP_BACKEND_API_URL + '/api/v1/routers/'+rtrid+'/sources', {responseType: 'json'});
    }

    getRouterDestinations(rtrid: number): Observable<RouterDestination[]> {
        return this.http.get<RouterDestination[]>(import.meta.env.NG_APP_BACKEND_API_URL + '/api/v1/routers/'+rtrid+'/destinations', {responseType: 'json'});
    }

    getRouterCrosspoints(rtrid: number): Observable<RouterCrosspoint[]> {
        return this.http.get<RouterCrosspoint[]>(import.meta.env.NG_APP_BACKEND_API_URL + '/api/v1/routers/'+rtrid+'/crosspoints', {responseType: 'json'});
    }

    getRouterLevels(rtrid: number): Observable<RouterLevel[]> {
        return this.http.get<RouterLevel[]>(import.meta.env.NG_APP_BACKEND_API_URL + '/api/v1/routers/'+rtrid+'/levels', {responseType: 'json'});
    }

    getRouterTable(rtrid: number): Observable<RouterTableLine[]> {
        return this.http.get<RouterTableLine[]>(import.meta.env.NG_APP_BACKEND_API_URL + '/api/v1/routers/'+rtrid+'/table', {responseType: 'json'});
    }

    getRouterTableValidSources(rtrid: number): Observable<RouterTableValidSources> {
        return this.http.get<RouterTableValidSources>(import.meta.env.NG_APP_BACKEND_API_URL + '/api/v1/routers/'+rtrid+'/validsources', {responseType: 'json'});
    }

    putRouterCrosspoint(rtr_id: number, destination_id: number, destination_level_id: number, source_id: number, source_level_id: number): Observable<any> {
        let url = import.meta.env.NG_APP_BACKEND_API_URL + '/api/v1/routers/'+rtr_id+'/crosspoints';
        let body = {
            "destination_id": destination_id,
            "destination_level_id": destination_level_id,
            "source_id": source_id,
            "source_level_id": source_level_id
        };
        let httpOptions = {
            headers: new HttpHeaders({
                'Content-Type': 'application/json',
                'Access-Control-Allow-Origin': '*',
            }),
            keepalive: true,
            mode: 'cors' as RequestMode
        };
        console.log(url, body, httpOptions)
        return this.http.put<any>(url, body, httpOptions);
    }
}