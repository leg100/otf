import { default as FuseFull, IFuseOptions } from 'fuse.js';
import { default as FuseBasic } from 'fuse.js/basic';
import { Options } from '../interfaces/options';
import { Searcher, SearchResult } from '../interfaces/search';
export declare class SearchByFuse<T extends object> implements Searcher<T> {
    _fuseOptions: IFuseOptions<T>;
    _haystack: T[];
    _fuse: FuseFull<T> | FuseBasic<T> | undefined;
    constructor(config: Options);
    index(data: T[]): void;
    reset(): void;
    isEmptyIndex(): boolean;
    search(needle: string): SearchResult<T>[];
}
