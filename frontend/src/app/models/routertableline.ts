import { RouterTableCrosspoint } from "./routertablecrosspoint";

export interface RouterTableLine {
    id:             number;
    name:           string;
    crosspoints:    RouterTableCrosspoint[];
    crosspoints_as_string: string[];
}