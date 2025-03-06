import { ClassNames } from '../interfaces/class-names';
import { EventTypes } from '../interfaces/event-type';
import { addClassesToElement, dispatchEvent, removeClassesFromElement } from '../lib/utils';
import { EventMap } from '../interfaces';

export default class WrappedElement<T extends HTMLInputElement | HTMLSelectElement> {
  element: T;

  classNames: ClassNames;

  isDisabled: boolean;

  constructor({ element, classNames }) {
    this.element = element;
    this.classNames = classNames;
    this.isDisabled = false;
  }

  get isActive(): boolean {
    return this.element.dataset.choice === 'active';
  }

  get dir(): string {
    return this.element.dir;
  }

  get value(): string {
    return this.element.value;
  }

  set value(value: string) {
    this.element.setAttribute('value', value);
    this.element.value = value;
  }

  conceal(): void {
    const el = this.element;
    // Hide passed input
    addClassesToElement(el, this.classNames.input);
    el.hidden = true;

    // Remove element from tab index
    el.tabIndex = -1;

    // Backup original styles if any
    const origStyle = el.getAttribute('style');

    if (origStyle) {
      el.setAttribute('data-choice-orig-style', origStyle);
    }

    el.setAttribute('data-choice', 'active');
  }

  reveal(): void {
    const el = this.element;
    // Reinstate passed element
    removeClassesFromElement(el, this.classNames.input);
    el.hidden = false;
    el.removeAttribute('tabindex');

    // Recover original styles if any
    const origStyle = el.getAttribute('data-choice-orig-style');

    if (origStyle) {
      el.removeAttribute('data-choice-orig-style');
      el.setAttribute('style', origStyle);
    } else {
      el.removeAttribute('style');
    }
    el.removeAttribute('data-choice');
  }

  enable(): void {
    this.element.removeAttribute('disabled');
    this.element.disabled = false;
    this.isDisabled = false;
  }

  disable(): void {
    this.element.setAttribute('disabled', '');
    this.element.disabled = true;
    this.isDisabled = true;
  }

  triggerEvent<K extends EventTypes>(eventType: EventTypes, data?: EventMap[K]['detail']): void {
    dispatchEvent(this.element, eventType, data || {});
  }
}
