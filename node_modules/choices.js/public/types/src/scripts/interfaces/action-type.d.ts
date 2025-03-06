import { Types } from './types';
export declare const ActionType: {
    readonly ADD_CHOICE: "ADD_CHOICE";
    readonly REMOVE_CHOICE: "REMOVE_CHOICE";
    readonly FILTER_CHOICES: "FILTER_CHOICES";
    readonly ACTIVATE_CHOICES: "ACTIVATE_CHOICES";
    readonly CLEAR_CHOICES: "CLEAR_CHOICES";
    readonly ADD_GROUP: "ADD_GROUP";
    readonly ADD_ITEM: "ADD_ITEM";
    readonly REMOVE_ITEM: "REMOVE_ITEM";
    readonly HIGHLIGHT_ITEM: "HIGHLIGHT_ITEM";
};
export type ActionTypes = Types.ValueOf<typeof ActionType>;
