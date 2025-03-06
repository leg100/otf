import { GroupActions } from '../actions/groups';
import { State } from '../interfaces/state';
import { ActionType } from '../interfaces';
import { StateUpdate } from '../interfaces/store';
import { ChoiceActions } from '../actions/choices';

type ActionTypes = ChoiceActions | GroupActions;
type StateType = State['groups'];

export default function groups(s: StateType, action: ActionTypes): StateUpdate<StateType> {
  let state = s;
  let update = true;

  switch (action.type) {
    case ActionType.ADD_GROUP: {
      state.push(action.group);
      break;
    }

    case ActionType.CLEAR_CHOICES: {
      state = [];
      break;
    }

    default: {
      update = false;
      break;
    }
  }

  return { state, update };
}
