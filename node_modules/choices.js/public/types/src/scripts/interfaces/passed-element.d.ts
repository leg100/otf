import { InputChoice } from './input-choice';
import { EventChoice } from './event-choice';
/**
 * Events fired by Choices behave the same as standard events. Each event is triggered on the element passed to Choices (accessible via `this.passedElement`. Arguments are accessible within the `event.detail` object.
 */
export interface EventMap {
    /**
     * Triggered each time an item is added (programmatically or by the user).
     *
     * **Input types affected:** text, select-one, select-multiple
     *
     * Arguments: id, value, label, groupValue
     */
    addItem: CustomEvent<EventChoice>;
    /**
     * Triggered each time an item is removed (programmatically or by the user).
     *
     * **Input types affected:** text, select-one, select-multiple
     *
     * Arguments: id, value, label, groupValue
     */
    removeItem: CustomEvent<EventChoice | undefined>;
    /**
     * Triggered each time an item is highlighted.
     *
     * **Input types affected:** text, select-multiple
     *
     * Arguments: id, value, label, groupValue
     */
    highlightItem: CustomEvent<EventChoice | undefined>;
    /**
     * Triggered each time an item is unhighlighted.
     *
     * **Input types affected:** text, select-multiple
     *
     * Arguments: id, value, label, groupValue
     */
    unhighlightItem: CustomEvent<EventChoice | undefined>;
    /**
     * Triggered each time a choice is selected **by a user**, regardless if it changes the value of the input.
     *
     * **Input types affected:** select-one, select-multiple
     *
     * Arguments: choice: Choice
     */
    choice: CustomEvent<{
        choice: InputChoice;
    }>;
    /**
     * Triggered each time an item is added/removed **by a user**.
     *
     * **Input types affected:** text, select-one, select-multiple
     *
     * Arguments: value
     */
    change: CustomEvent<{
        value: string;
    }>;
    /**
     * Triggered when a user types into an input to search choices. When a search is ended, a search event with an empty value with no resultCount is triggered.
     *
     * **Input types affected:** select-one, select-multiple
     *
     * Arguments: value, resultCount
     */
    search: CustomEvent<{
        value: string;
        resultCount: number;
    }>;
    /**
     * Triggered when the dropdown is shown.
     *
     * **Input types affected:** select-one, select-multiple
     *
     * Arguments: -
     */
    showDropdown: CustomEvent<undefined>;
    /**
     * Triggered when the dropdown is hidden.
     *
     * **Input types affected:** select-one, select-multiple
     *
     * Arguments: -
     */
    hideDropdown: CustomEvent<undefined>;
    /**
     * Triggered when a choice from the dropdown is highlighted.
     *
     * Input types affected: select-one, select-multiple
     * Arguments: el is the choice.passedElement that was affected.
     */
    highlightChoice: CustomEvent<{
        el: HTMLElement;
    }>;
}
