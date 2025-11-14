import { Component, ViewChild } from '@angular/core';
import { BackendService } from '../backendapi.service';
import { Router } from '../models/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { GridSettings, HotTableComponent, HotTableModule } from '@handsontable/angular-wrapper';

@Component({
  selector: 'app-router-table',
  imports: [MatToolbarModule, MatButtonModule, MatIconModule, MatMenuModule, HotTableModule],
  providers: [BackendService],
  templateUrl: './router-table.html',
  styleUrl: './router-table.scss',
})
export class RouterTable {
  @ViewChild(HotTableComponent, { static: false })

  readonly hotTable!: HotTableComponent;

  routers: Router[] = [];
  selectedRouter: Router = {} as Router;
  readonly hot_data = [["", "Tesla", "Volvo", "Toyota", "Ford"],
   ["2019", 10, 11, 12, 13],
   ["2020", 20, 11, 14, 13],
   ["2021", 30, 15, 12, 13],
 ];
   

  readonly gridSettings = <GridSettings>{
    rowHeaders: true,
    colHeaders: true,
    width: "auto",
    height: "auto",
    autoWrapRow: false,
    autoWrapCol: false,
  };

  constructor(
    private backendService: BackendService
  ) {}

  ngOnInit(): void {
    this.backendService.getRouters().subscribe(routers => this.routers = routers);
    console.log(this.routers);
    if (this.routers.length > 0) {
      this.selectedRouter = this.routers[0]
    }
  }
  
  selectRouter(rtr: Router): void {
    this.selectedRouter = rtr;
  }
}
