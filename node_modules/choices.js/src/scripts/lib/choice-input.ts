import { InputChoice } from '../interfaces/input-choice';
import { InputGroup } from '../interfaces/input-group';
import { GroupFull } from '../interfaces/group-full';
import { ChoiceFull } from '../interfaces/choice-full';
import { sanitise, unwrapStringForRaw } from './utils';

type MappedInputTypeToChoiceType<T extends string | InputChoice | InputGroup> = T extends InputGroup
  ? GroupFull
  : ChoiceFull;

export const coerceBool = (arg: unknown, defaultValue: boolean = true): boolean =>
  typeof arg === 'undefined' ? defaultValue : !!arg;

export const stringToHtmlClass = (input: string | string[] | undefined): string[] | undefined => {
  if (typeof input === 'string') {
    // eslint-disable-next-line no-param-reassign
    input = input.split(' ').filter((s) => s.length);
  }

  if (Array.isArray(input) && input.length) {
    return input;
  }

  return undefined;
};

export const mapInputToChoice = <T extends string | InputChoice | InputGroup>(
  value: T,
  allowGroup: boolean,
  allowRawString: boolean = true,
): MappedInputTypeToChoiceType<T> => {
  if (typeof value === 'string') {
    const sanitisedValue = sanitise(value);
    const userValue = allowRawString || sanitisedValue === value ? value : { escaped: sanitisedValue, raw: value };

    const result: ChoiceFull = mapInputToChoice<InputChoice>(
      {
        value,
        label: userValue,
        selected: true,
      },
      false,
    );

    return result as MappedInputTypeToChoiceType<T>;
  }

  const groupOrChoice = value as InputChoice | InputGroup;
  if ('choices' in groupOrChoice) {
    if (!allowGroup) {
      // https://developer.mozilla.org/en-US/docs/Web/HTML/Element/optgroup
      throw new TypeError(`optGroup is not allowed`);
    }
    const group = groupOrChoice;
    const choices = group.choices.map((e) => mapInputToChoice<InputChoice>(e, false));

    const result: GroupFull = {
      id: 0, // actual ID will be assigned during _addGroup
      label: unwrapStringForRaw(group.label) || group.value,
      active: !!choices.length,
      disabled: !!group.disabled,
      choices,
    };

    return result as MappedInputTypeToChoiceType<T>;
  }

  const choice = groupOrChoice;

  const result: ChoiceFull = {
    id: 0, // actual ID will be assigned during _addChoice
    group: null, // actual group will be assigned during _addGroup but before _addChoice
    score: 0, // used in search
    rank: 0, // used in search, stable sort order
    value: choice.value,
    label: choice.label || choice.value,
    active: coerceBool(choice.active),
    selected: coerceBool(choice.selected, false),
    disabled: coerceBool(choice.disabled, false),
    placeholder: coerceBool(choice.placeholder, false),
    highlighted: false,
    labelClass: stringToHtmlClass(choice.labelClass),
    labelDescription: choice.labelDescription,
    customProperties: choice.customProperties,
  };

  return result as MappedInputTypeToChoiceType<T>;
};
