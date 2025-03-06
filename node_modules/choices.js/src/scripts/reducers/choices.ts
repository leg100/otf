/* eslint-disable */
import { ActionType, Options, State } from '../interfaces';
import { StateUpdate } from '../interfaces/store';
import { ChoiceActions } from '../actions/choices';
import { ItemActions } from '../actions/items';
import { SearchResult } from '../interfaces/search';
import { ChoiceFull } from '../interfaces/choice-full';

type ActionTypes = ChoiceActions | ItemActions;
type StateType = State['choices'];

export default function choices(s: StateType, action: ActionTypes, context?: Options): StateUpdate<StateType> {
  let state = s;
  let update = true;

  switch (action.type) {
    case ActionType.ADD_CHOICE: {
      state.push(action.choice);
      break;
    }

    case ActionType.REMOVE_CHOICE: {
      action.choice.choiceEl = undefined;

      if (action.choice.group) {
        action.choice.group.choices = action.choice.group.choices.filter((obj) => obj.id !== action.choice.id);
      }
      state = state.filter((obj) => obj.id !== action.choice.id);
      break;
    }

    case ActionType.ADD_ITEM:
    case ActionType.REMOVE_ITEM: {
      action.item.choiceEl = undefined;
      break;
    }

    case ActionType.FILTER_CHOICES: {
      // avoid O(n^2) algorithm complexity when searching/filtering choices
      const scoreLookup: SearchResult<ChoiceFull>[] = [];
      action.results.forEach((result) => {
        scoreLookup[result.item.id] = result;
      });

      state.forEach((choice) => {
        const result = scoreLookup[choice.id];
        if (result !== undefined) {
          choice.score = result.score;
          choice.rank = result.rank;
          choice.active = true;
        } else {
          choice.score = 0;
          choice.rank = 0;
          choice.active = false;
        }
        if (context && context.appendGroupInSearch) {
          choice.choiceEl = undefined;
        }
      });

      break;
    }

    case ActionType.ACTIVATE_CHOICES: {
      state.forEach((choice) => {
        choice.active = action.active;
        if (context && context.appendGroupInSearch) {
          choice.choiceEl = undefined;
        }
      });
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
