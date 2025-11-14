import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Hotdemo } from './hotdemo';

describe('Hotdemo', () => {
  let component: Hotdemo;
  let fixture: ComponentFixture<Hotdemo>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Hotdemo]
    })
    .compileComponents();

    fixture = TestBed.createComponent(Hotdemo);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
