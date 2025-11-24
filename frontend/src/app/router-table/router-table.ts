import { ChangeDetectorRef, Component, ViewChild } from '@angular/core';
import { FormControl, FormsModule, ReactiveFormsModule } from '@angular/forms';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { BackendService } from '../backendapi.service';
import { Router } from '../models/router';
import { HotTableComponent, HotTableModule } from '@handsontable/angular-wrapper';
import Handsontable from 'handsontable';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { RouterTableLine } from '../models/routertableline';
import { RouterTableValidSources } from '../models/routertablevalidsources';
import { autocompleteRenderer, baseRenderer, registerRenderer, textRenderer } from 'handsontable/renderers';
import { RouterDestination } from '../models/routerDestination';
import { RouterCrosspoint } from '../models/routerCrosspoint';
import { Subscription } from 'rxjs';


@Component({
  selector: 'app-router-table',
  imports: [MatToolbarModule, MatButtonModule, MatIconModule, MatMenuModule, HotTableModule, MatFormFieldModule, MatInputModule, FormsModule, ReactiveFormsModule],
  providers: [BackendService],
  templateUrl: './router-table.html',
  styleUrl: './router-table.scss',
})
export class RouterTable {
  @ViewChild(HotTableComponent, { static: false }) hotTable!: HotTableComponent;
  filterFieldControl: FormControl = new FormControl('');
  filterValue: string = "";

  private crosspointSubscription: Subscription;
  lastSocketReceivedTime: number = 0;

  routers: Router[] = [];
  selectedRouter: Router = {} as Router;
  routerTable: RouterTableLine[] = [];
  routerTableById: Map<number, RouterTableLine> = new Map<number, RouterTableLine>();  
  routerValidSources: RouterTableValidSources = {
    sources_as_string: [],
    sources: [],
  };
  queuedChanges: number[][] = [];

  hot_data: any[][] = [];

  gridSettings_colHeaders: string[] = [];
  
  readonly routerCrosspointRenderer = (instance: Handsontable, td: HTMLTableCellElement, row: number, col: number, prop: any, value: any, cellProperties: any) => {
    Handsontable.renderers.AutocompleteRenderer( instance, td, row, col, prop, value, cellProperties);
    let rowData = instance.getDataAtRow(row);
    let destination_id: number = rowData[0] as number;
    let destination_level_id: number = col-1
    // Handle queued changes
    for (let i = 0; i < this.queuedChanges.length; i++) {
      if (this.queuedChanges[i][0] === destination_id as number && this.queuedChanges[i][1] === destination_level_id) {
        td.style.backgroundColor = 'green';
      }
    }
    // Handle Locks
    if (this.routerTableById.get(destination_id)!.crosspoints[destination_level_id-1].locked) {
      if (td.style.backgroundColor == 'green') {
        // Cell is queued for take but locked
        td.style.borderColor = 'orangered';
      } else {
        td.style.backgroundColor = 'orangered';
      }
    }
  };
   

  gridSettings: Handsontable.GridSettings = {
    rowHeaders: false,
    colHeaders: true,
    autoWrapRow: false,
    autoWrapCol: false,
    fillHandle: {
      autoInsertRow: false
    },
    //selectionMode: 'range',
    outsideClickDeselects: false,
    columns: (index: number) => {
      switch (index) {
        case 0:
          return {
            readOnly: true,
            type: "numeric"
          }
        case 1:
          return {
            readOnly: true,
          }
        default:
          return {
            type: "autocomplete",
            source: this.routerValidSources.sources_as_string[index-2],
            strict: true,
          }
      }
    },
    filters: true,
    dropdownMenu: ['filter_by_condition', 'filter_by_value', 'filter_action_bar', 'filter_operators'],
    cells: (row, column, prop) => {
      if (column >= 2) {
        if (!this.selectedRouter.destinations[row].levels.includes(column-1)) {
          return {
            readOnly: true,
          }
        }
        return {
          renderer: this.routerCrosspointRenderer,
        }
      }
      return {};
    },
    afterChange: (changes, source) => {
      // Ignore page loads
      if (source === 'loadData' || changes === null) {
        return
      }
      for (let i: number = 0; i < changes!.length; i++) {
        let row = changes[i][0];
        let col: number = changes[i][1] as number;
        //let oldValue = changes[i][2];
        let newValue = changes[i][3];
        //let destID = this.routerTable[row].id
        let rowData = this.hotTable.hotInstance!.getDataAtRow(row); 
        let destID = rowData[0] as number;
        let destLvl = this.selectedRouter.levels[col-2].id
        let sourceID = this.routerValidSources.sources[col-2][this.routerValidSources.sources_as_string[col-2].indexOf(newValue)].source_id
        let sourceLevelID = this.routerValidSources.sources[col-2][this.routerValidSources.sources_as_string[col-2].indexOf(newValue)].source_level_id
        this.queueChange(destID, destLvl, sourceID, sourceLevelID);
      }
      this.hotTable.hotInstance!.render();
    },
  };

  constructor(
    private backendService: BackendService,
    private changeDetectorRef: ChangeDetectorRef,
  ) {
    this.crosspointSubscription = this.backendService.getWebsocketCrosspoints().subscribe(
      (crosspoint) => {
        this.updateCrosspoint(crosspoint);
      }
    );
  }

  ngOnInit(): void {
    this.backendService.getRouters().subscribe(routers => this.updateRouters(routers));

    registerRenderer('routerCrosspointRenderer', this.routerCrosspointRenderer);
    
    this.filterFieldControl.valueChanges.subscribe((filterValue: string) => {
      const filters = this.hotTable.hotInstance!.getPlugin('filters');
      filters.removeConditions(1);
      filters.addCondition(1, 'contains', [filterValue]);
      filters.filter();
      this.hotTable.hotInstance!.render();
    });
  }

  ngOnDestroy(): void {
    this.crosspointSubscription.unsubscribe();
    this.backendService.closeWebsocketConnection();
  }

  updateRouters(routers: Router[]): void {
    this.routers = routers;
    this.changeDetectorRef.detectChanges();
    if (this.routers.length > 0) {
      this.selectRouter(routers[0])
    }
  }
  
  selectRouter(rtr: Router): void {
    this.selectedRouter = rtr;
    this.changeDetectorRef.detectChanges();
    this.fetchRouterTable();
  }

  fetchRouterTable(): void {
    this.backendService.getRouterLevels(this.selectedRouter.id).subscribe(lvls => {this.selectedRouter.levels = lvls; this.setHeaders()})
    this.backendService.getRouterSources(this.selectedRouter.id).subscribe(srcs => {this.selectedRouter.sources = srcs});
    this.backendService.getRouterDestinations(this.selectedRouter.id).subscribe(dsts => {this.selectedRouter.destinations = dsts; });
    this.backendService.getRouterCrosspoints(this.selectedRouter.id).subscribe(xpts => {this.selectedRouter.crosspoints = xpts});
    this.backendService.getRouterTable(this.selectedRouter.id).subscribe(table => {this.routerTable = table; this.setDestinations()});
    this.backendService.getRouterTableValidSources(this.selectedRouter.id).subscribe(validSources => {this.routerValidSources = validSources; this.setHeaders()});
  }

  setHeaders(): void {
    let headers: string[] = ["ID", "Destination"];
    for (let i = 0; i < this.selectedRouter.levels.length; i++) {
      headers.push(this.selectedRouter.levels[i].name);     
    }
    this.gridSettings.colHeaders = headers
    this.hotTable.hotInstance!.updateSettings(this.gridSettings);
  }
  
  setDestinations(): void {
    let routerTableMapData = this.routerTable.map(line => [line.id, line] as [number, RouterTableLine]);
    this.routerTableById = new Map<number, RouterTableLine>(routerTableMapData);
    for (let i = 0; i < this.routerTable.length; i++) {
      this.hot_data[i] = [
        this.routerTable[i].id,
        this.routerTable[i].name
      ]
      for (let j = 0; j < this.routerTable[i].crosspoints_as_string.length; j++) {
        this.hot_data[i].push(this.routerTable[i].crosspoints_as_string[j])
      }
    }
    this.hotTable.hotInstance!.updateData(this.hot_data)
  }

  updateCrosspoint(crosspoint: RouterCrosspoint): void {
    this.lastSocketReceivedTime = Date.now();
    // Rotuer Table by ID
    let tableLine = this.routerTableById.get(crosspoint.destination) as RouterTableLine;
    tableLine.crosspoints[crosspoint.destination_level-1].source_id = crosspoint.source;
    tableLine.crosspoints[crosspoint.destination_level-1].source_level_id = crosspoint.source_level;
    tableLine.crosspoints[crosspoint.destination_level-1].locked = crosspoint.locked;
    tableLine.crosspoints_as_string[crosspoint.destination_level-1] = this.getCrosspointString(crosspoint.source, crosspoint.source_level);
    this.routerTableById.set(crosspoint.destination, tableLine);
    // Router table
    for (let i: number = 0; i < this.routerTable.length; i++) {
      if (this.routerTable[i].id == crosspoint.destination) {
        this.routerTable[i] = tableLine;
        break;
      }
    }
    // Update queued changes to remove any taken routes
    this.filterQueuedChanges();
    // Update draw
    this.hot_data = this.hot_data.slice(0, this.routerTable.length);
    for (let i = 0; i < this.routerTable.length; i++) {
      this.hot_data[i] = [
        this.routerTable[i].id,
        this.routerTable[i].name
      ]
      for (let j = 0; j < this.routerTable[i].crosspoints_as_string.length; j++) {
        this.hot_data[i].push(this.routerTable[i].crosspoints_as_string[j])
      }
    }
    setTimeout(() => {
      if (Date.now() - this.lastSocketReceivedTime > 100) {
        this.hotTable.hotInstance!.updateData(this.hot_data);
        this.hotTable.hotInstance!.render();
      }
    }, 150);
  }

  getCrosspointString(source_id: number, source_level_id: number): string {
    let result: string = "";
    for (let i: number = 0; i < this.selectedRouter.sources.length; i++) {
      if (this.selectedRouter.sources[i].id == source_id) {
        result = result + this.selectedRouter.sources[i].name;
        break;
      }
    }
    for (let i: number = 0; i < this.selectedRouter.levels.length; i++) {
      if (this.selectedRouter.levels[i].id == source_level_id) {
        result = result + "." + this.selectedRouter.levels[i].name
      }
    }
    return result;
  }

  queueChange(destID: number, destLevelID: number, sourceID: number, sourceLevelID: number): void {
    // Make sure destination does not already exist in array
    let destinationIndex: number = -1;
    for (let i: number = 0; i < this.queuedChanges.length; i++) {
      if (this.queuedChanges[i][0] === destID && this.queuedChanges[i][1] === destLevelID) {
        destinationIndex = i;
        break;
      }
    }
    if (destinationIndex === -1) {
      this.queuedChanges.push([destID, destLevelID, sourceID, sourceLevelID]);
    } else {
      this.queuedChanges[destinationIndex] = [destID, destLevelID, sourceID, sourceLevelID];
    }
    // Filter routes which don't change anything
    this.filterQueuedChanges();
  }

  filterQueuedChanges(): void {
    let deleteIdxs: number[] = [];
    for (let i: number = 0; i < this.queuedChanges.length; i++) {
      if (this.queuedChanges[i][2] === this.routerTableById.get(this.queuedChanges[i][0])!.crosspoints[this.queuedChanges[i][3]-1].source_id) {
        if (this.queuedChanges[i][3] === this.routerTableById.get(this.queuedChanges[i][0])!.crosspoints[this.queuedChanges[i][3]-1].source_level_id) {
          deleteIdxs.push(i);
        }
      }
    }
    for (let i: number = deleteIdxs.length-1; i >= 0; i--) {
      this.queuedChanges.splice(i, 1);
    }
  }

  take(): void {
    for (let i: number = 0; i < this.queuedChanges.length; i++) {
      let destination_id = this.queuedChanges[i][0];
      let destination_lvl_id = this.queuedChanges[i][1];
      let source_id = this.queuedChanges[i][2];
      let source_level_id = this.queuedChanges[i][3];
      this.backendService.putRouterCrosspoint(this.selectedRouter.id, destination_id, destination_lvl_id, source_id, source_level_id).subscribe();
    }
    this.filterQueuedChanges();
  }

  toggleLock(): void {
    const hot = this.hotTable?.hotInstance;
    const selected = hot?.getSelectedRange() || [];

    // Range selected
    for (let i: number = 0; i < selected.length; i++) {
      const row1 = selected[0].from.row;
      const row2 = selected[0].to.row;
      const col1 = selected[0].from.col;
      const col2 = selected[0].to.col;
      const rowStart = Math.max(row1, 0)
      const colStart = Math.max(col1, 2);
      for (let row_id: number = rowStart; row_id <= row2; row_id++) {
        const rowData = this.hotTable.hotInstance?.getDataAtRow(row_id) || [];
        const destination_id = rowData[0] as number || 0;
        for (let col_id: number = colStart; col_id <= col2; col_id++) {
          const destination_level_id = col_id-1;
          const locked = (!this.routerTableById.get(destination_id)?.crosspoints[destination_level_id-1].locked) || false;
          if (destination_id != 0) {
            this.backendService.putRouterCrosspointLock(this.selectedRouter.id, destination_id, destination_level_id, locked).subscribe();
          }
        }  
      }
    }
  }
}
