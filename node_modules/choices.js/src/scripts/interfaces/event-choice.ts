import { InputChoice } from './input-choice';

export type EventChoiceValueType<B extends boolean> = B extends true ? string : EventChoice;

export interface EventChoice extends InputChoice {
  element?: HTMLOptionElement | HTMLOptGroupElement;
  groupValue?: string;
  keyCode?: number;
}
