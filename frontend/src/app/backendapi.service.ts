import { Injectable } from "@angular/core";
import { HttpClient, HttpHeaders } from "@angular/common/http";
import { Observable } from "rxjs";
import { Router } from "./models/router";

const httpOptions = {
    headers: new HttpHeaders({
        'Content-Type': 'application/json',
    })
};

@Injectable({
    providedIn: 'root'
  })
export class BackendService {
    constructor(
        private http: HttpClient
    ) {}

    getRouters(): Observable<Router[]> {
        console.log(import.meta.env.NG_APP_BACKEND_API_URL)
        return this.http.get<Router[]>(import.meta.env.NG_APP_BACKEND_API_URL + '/api/v1/routers', {responseType: 'json'});
    }
}