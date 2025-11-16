import { Routes } from '@angular/router';
import { RouterTable } from './router-table/router-table';

export const routes: Routes = [
    { path: '', redirectTo: '/router', pathMatch: 'full' },
    { path: 'router', component: RouterTable }
];
