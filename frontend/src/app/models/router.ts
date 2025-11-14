import { RouterCrosspoint } from "./routerCrosspoint";
import { RouterDestination } from "./routerDestination";
import { RouterLevel } from "./routerLevel";
import { RouterSource } from "./routerSource";

export interface Router {
    id:             number;
    display_name:   string;
    short_name:     string;
    levels:         RouterLevel[];
    sources:        RouterSource[];
    destinations:   RouterDestination[];
    crosspoints:    RouterCrosspoint[];
}