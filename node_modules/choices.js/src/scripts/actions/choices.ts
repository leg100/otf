import { ChoiceFull } from '../interfaces/choice-full';
import { ActionType } from '../interfaces';
import { SearchResult } from '../interfaces/search';
import { AnyAction } from '../interfaces/store';

export type ChoiceActions =
  | AddChoiceAction
  | RemoveChoiceAction
  | FilterChoicesAction
  | ActivateChoicesAction
  | ClearChoicesAction;

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
export interface ClearChoicesAction extends AnyAction<typeof ActionType.CLEAR_CHOICES> {}

export const addChoice = (choice: ChoiceFull): AddChoiceAction => ({
  type: ActionType.ADD_CHOICE,
  choice,
});

export const removeChoice = (choice: ChoiceFull): RemoveChoiceAction => ({
  type: ActionType.REMOVE_CHOICE,
  choice,
});

export const filterChoices = (results: SearchResult<ChoiceFull>[]): FilterChoicesAction => ({
  type: ActionType.FILTER_CHOICES,
  results,
});

export const activateChoices = (active = true): ActivateChoicesAction => ({
  type: ActionType.ACTIVATE_CHOICES,
  active,
});

/**
 * @deprecated use clearStore() or clearChoices() instead.
 */
export const clearChoices = (): ClearChoicesAction => ({
  type: ActionType.CLEAR_CHOICES,
});
