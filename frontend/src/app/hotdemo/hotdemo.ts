import { Component, ViewChild } from '@angular/core';
import { HotTableComponent, HotTableModule, GridSettings } from '@handsontable/angular-wrapper';

@Component({
  selector: 'app-hotdemo',
  imports: [HotTableModule],
  templateUrl: './hotdemo.html',
  styleUrl: './hotdemo.scss',
})
export class Hotdemo {
  @ViewChild(HotTableComponent, { static: false })
  readonly hotTable!: HotTableComponent;

  readonly data = [
    ["", "Tesla", "Volvo", "Toyota", "Ford"],
    ["2019", 10, 11, 12, 13],
    ["2020", 20, 11, 14, 13],
    ["2021", 30, 15, 12, 13],
  ];
  readonly gridSettings = <GridSettings>{
    rowHeaders: true,
    colHeaders: true,
    height: "auto",
    autoWrapRow: true,
    autoWrapCol: true,
  };
}
