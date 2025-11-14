import { bootstrapApplication } from '@angular/platform-browser';
import { appConfig } from './app/app.config';
import { App } from './app/app';

import { HotTableModule } from '@handsontable/angular';
import Handsontable from 'handsontable';

console.log(`Handsontable: v${Handsontable.version} (${Handsontable.buildDate}) Wrapper: v${HotTableModule.version}`);

bootstrapApplication(App, appConfig)
  .catch((err) => console.error(err));
