import { GroupActions } from '../actions/groups';
import { State } from '../interfaces/state';
import { StateUpdate } from '../interfaces/store';
import { ChoiceActions } from '../actions/choices';
type ActionTypes = ChoiceActions | GroupActions;
type StateType = State['groups'];
export default function groups(s: StateType, action: ActionTypes): StateUpdate<StateType>;
export {};
