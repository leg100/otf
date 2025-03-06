import { Types } from './types';

export const EventType = {
  showDropdown: 'showDropdown',
  hideDropdown: 'hideDropdown',
  change: 'change',
  choice: 'choice',
  search: 'search',
  addItem: 'addItem',
  removeItem: 'removeItem',
  highlightItem: 'highlightItem',
  highlightChoice: 'highlightChoice',
  unhighlightItem: 'unhighlightItem',
} as const;

export type EventTypes = Types.ValueOf<typeof EventType>;
