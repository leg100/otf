import { ClassNames } from '../interfaces/class-names';
import { EventTypes } from '../interfaces/event-type';
import { EventMap } from '../interfaces';
export default class WrappedElement<T extends HTMLInputElement | HTMLSelectElement> {
    element: T;
    classNames: ClassNames;
    isDisabled: boolean;
    constructor({ element, classNames }: {
        element: any;
        classNames: any;
    });
    get isActive(): boolean;
    get dir(): string;
    get value(): string;
    set value(value: string);
    conceal(): void;
    reveal(): void;
    enable(): void;
    disable(): void;
    triggerEvent<K extends EventTypes>(eventType: EventTypes, data?: EventMap[K]['detail']): void;
}
