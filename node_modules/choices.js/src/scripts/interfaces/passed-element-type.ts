import { Types } from './types';

export const PassedElementTypes = {
  Text: 'text',
  SelectOne: 'select-one',
  SelectMultiple: 'select-multiple',
} as const;

export type PassedElementType = Types.ValueOf<typeof PassedElementTypes>;
