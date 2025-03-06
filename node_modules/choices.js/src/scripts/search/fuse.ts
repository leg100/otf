// eslint-disable-next-line import/no-named-default
import { default as FuseFull, IFuseOptions } from 'fuse.js';
// eslint-disable-next-line import/no-named-default
import { default as FuseBasic } from 'fuse.js/basic';
import { Options } from '../interfaces/options';
import { Searcher, SearchResult } from '../interfaces/search';
import { searchFuse } from '../interfaces/build-flags';

export class SearchByFuse<T extends object> implements Searcher<T> {
  _fuseOptions: IFuseOptions<T>;

  _haystack: T[] = [];

  _fuse: FuseFull<T> | FuseBasic<T> | undefined;

  constructor(config: Options) {
    this._fuseOptions = {
      ...config.fuseOptions,
      keys: [...config.searchFields],
      includeMatches: true,
    };
  }

  index(data: T[]): void {
    this._haystack = data;
    if (this._fuse) {
      this._fuse.setCollection(data);
    }
  }

  reset(): void {
    this._haystack = [];
    this._fuse = undefined;
  }

  isEmptyIndex(): boolean {
    return !this._haystack.length;
  }

  search(needle: string): SearchResult<T>[] {
    if (!this._fuse) {
      if (searchFuse === 'full') {
        this._fuse = new FuseFull<T>(this._haystack, this._fuseOptions);
      } else {
        this._fuse = new FuseBasic<T>(this._haystack, this._fuseOptions);
      }
    }

    const results = this._fuse.search(needle);

    return results.map((value, i): SearchResult<T> => {
      return {
        item: value.item,
        score: value.score || 0,
        rank: i + 1, // If value.score is used for sorting, this can create non-stable sorts!
      };
    });
  }
}
