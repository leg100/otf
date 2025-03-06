import { Options } from '../interfaces';
import { Searcher } from '../interfaces/search';
export declare function getSearcher<T extends object>(config: Options): Searcher<T>;
