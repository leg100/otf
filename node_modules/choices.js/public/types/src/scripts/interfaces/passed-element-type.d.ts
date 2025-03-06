import { Types } from './types';
export declare const PassedElementTypes: {
    readonly Text: "text";
    readonly SelectOne: "select-one";
    readonly SelectMultiple: "select-multiple";
};
export type PassedElementType = Types.ValueOf<typeof PassedElementTypes>;
