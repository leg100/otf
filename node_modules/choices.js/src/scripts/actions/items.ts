import { ChoiceFull } from '../interfaces/choice-full';
import { ActionType } from '../interfaces';
import { AnyAction } from '../interfaces/store';

export type ItemActions = AddItemAction | RemoveItemAction | HighlightItemAction;

export interface AddItemAction extends AnyAction<typeof ActionType.ADD_ITEM> {
  item: ChoiceFull;
}

export interface RemoveItemAction extends AnyAction<typeof ActionType.REMOVE_ITEM> {
  item: ChoiceFull;
}

export interface HighlightItemAction extends AnyAction<typeof ActionType.HIGHLIGHT_ITEM> {
  item: ChoiceFull;
  highlighted: boolean;
}

export const addItem = (item: ChoiceFull): AddItemAction => ({
  type: ActionType.ADD_ITEM,
  item,
});

export const removeItem = (item: ChoiceFull): RemoveItemAction => ({
  type: ActionType.REMOVE_ITEM,
  item,
});

export const highlightItem = (item: ChoiceFull, highlighted: boolean): HighlightItemAction => ({
  type: ActionType.HIGHLIGHT_ITEM,
  item,
  highlighted,
});
