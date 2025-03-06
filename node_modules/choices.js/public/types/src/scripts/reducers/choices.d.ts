import { Options, State } from '../interfaces';
import { StateUpdate } from '../interfaces/store';
import { ChoiceActions } from '../actions/choices';
import { ItemActions } from '../actions/items';
type ActionTypes = ChoiceActions | ItemActions;
type StateType = State['choices'];
export default function choices(s: StateType, action: ActionTypes, context?: Options): StateUpdate<StateType>;
export {};
