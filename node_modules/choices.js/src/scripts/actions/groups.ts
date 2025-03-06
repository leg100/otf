import { GroupFull } from '../interfaces/group-full';
import { ActionType } from '../interfaces';
import { AnyAction } from '../interfaces/store';

export type GroupActions = AddGroupAction;

export interface AddGroupAction extends AnyAction<typeof ActionType.ADD_GROUP> {
  group: GroupFull;
}

export const addGroup = (group: GroupFull): AddGroupAction => ({
  type: ActionType.ADD_GROUP,
  group,
});
