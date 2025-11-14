import { Routes } from '@angular/router';
import { RouterTable } from './router-table/router-table';
import { Hotdemo } from './hotdemo/hotdemo';

export const routes: Routes = [
    { path: '', redirectTo: '/router', pathMatch: 'full' },
    { path: 'router', component: RouterTable },
    { path: 'hotdemo', component: Hotdemo },
];
