import { ChoiceFull } from '../interfaces/choice-full';
import { ActionType } from '../interfaces';
import { SearchResult } from '../interfaces/search';
import { AnyAction } from '../interfaces/store';
export type ChoiceActions = AddChoiceAction | RemoveChoiceAction | FilterChoicesAction | ActivateChoicesAction | ClearChoicesAction;
export interface AddChoiceAction extends AnyAction<typeof ActionType.ADD_CHOICE> {
    choice: ChoiceFull;
}
export interface RemoveChoiceAction extends AnyAction<typeof ActionType.REMOVE_CHOICE> {
    choice: ChoiceFull;
}
export interface FilterChoicesAction extends AnyAction<typeof ActionType.FILTER_CHOICES> {
    results: SearchResult<ChoiceFull>[];
}
export interface ActivateChoicesAction extends AnyAction<typeof ActionType.ACTIVATE_CHOICES> {
    active: boolean;
}
/**
 * @deprecated use clearStore() or clearChoices() instead.
 */
export interface ClearChoicesAction extends AnyAction<typeof ActionType.CLEAR_CHOICES> {
}
export declare const addChoice: (choice: ChoiceFull) => AddChoiceAction;
export declare const removeChoice: (choice: ChoiceFull) => RemoveChoiceAction;
export declare const filterChoices: (results: SearchResult<ChoiceFull>[]) => FilterChoicesAction;
export declare const activateChoices: (active?: boolean) => ActivateChoicesAction;
/**
 * @deprecated use clearStore() or clearChoices() instead.
 */
export declare const clearChoices: () => ClearChoicesAction;
