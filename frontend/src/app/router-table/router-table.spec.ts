import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RouterTable } from './router-table';

describe('RouterTable', () => {
  let component: RouterTable;
  let fixture: ComponentFixture<RouterTable>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [RouterTable]
    })
    .compileComponents();

    fixture = TestBed.createComponent(RouterTable);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
