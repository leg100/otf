import { StringUntrusted } from './string-untrusted';
import { Types } from './types';
export interface InputChoice {
    id?: number;
    highlighted?: boolean;
    labelClass?: string | Array<string>;
    labelDescription?: string;
    customProperties?: Types.CustomProperties;
    disabled?: boolean;
    active?: boolean;
    label: StringUntrusted | string;
    placeholder?: boolean;
    selected?: boolean;
    value: any;
}
