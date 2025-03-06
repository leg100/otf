import { Types } from './types';

export const ActionType = {
  ADD_CHOICE: 'ADD_CHOICE',
  REMOVE_CHOICE: 'REMOVE_CHOICE',
  FILTER_CHOICES: 'FILTER_CHOICES',
  ACTIVATE_CHOICES: 'ACTIVATE_CHOICES',
  CLEAR_CHOICES: 'CLEAR_CHOICES',
  ADD_GROUP: 'ADD_GROUP',
  ADD_ITEM: 'ADD_ITEM',
  REMOVE_ITEM: 'REMOVE_ITEM',
  HIGHLIGHT_ITEM: 'HIGHLIGHT_ITEM',
} as const;

export type ActionTypes = Types.ValueOf<typeof ActionType>;
