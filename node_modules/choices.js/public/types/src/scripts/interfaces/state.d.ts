import { ChoiceFull } from './choice-full';
import { GroupFull } from './group-full';
export interface State {
    choices: ChoiceFull[];
    groups: GroupFull[];
    items: ChoiceFull[];
}
export type StateChangeSet = {
    [K in keyof State]: boolean;
};
