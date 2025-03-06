import { Types } from './types';
export declare const EventType: {
    readonly showDropdown: "showDropdown";
    readonly hideDropdown: "hideDropdown";
    readonly change: "change";
    readonly choice: "choice";
    readonly search: "search";
    readonly addItem: "addItem";
    readonly removeItem: "removeItem";
    readonly highlightItem: "highlightItem";
    readonly highlightChoice: "highlightChoice";
    readonly unhighlightItem: "unhighlightItem";
};
export type EventTypes = Types.ValueOf<typeof EventType>;
