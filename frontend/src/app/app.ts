import { Component, signal } from '@angular/core';
import { RouterModule } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatIconModule } from '@angular/material/icon';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatListModule } from '@angular/material/list';

import Handsontable from "handsontable/base";
import { registerAllModules } from "handsontable/registry";

registerAllModules();

@Component({
  selector: 'app-root',
  imports: [RouterModule, MatButtonModule, MatSidenavModule, MatIconModule, MatToolbarModule, MatListModule],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App {
  protected readonly title = signal('Broadcast Facility Controller');
  showFiller = false;
}
