import { GroupFull } from '../interfaces/group-full';
import { ActionType } from '../interfaces';
import { AnyAction } from '../interfaces/store';
export type GroupActions = AddGroupAction;
export interface AddGroupAction extends AnyAction<typeof ActionType.ADD_GROUP> {
    group: GroupFull;
}
export declare const addGroup: (group: GroupFull) => AddGroupAction;
