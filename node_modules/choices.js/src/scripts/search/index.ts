import { Options } from '../interfaces';
import { Searcher } from '../interfaces/search';
import { SearchByPrefixFilter } from './prefix-filter';
import { SearchByFuse } from './fuse';
import { searchFuse } from '../interfaces/build-flags';

export function getSearcher<T extends object>(config: Options): Searcher<T> {
  if (searchFuse) {
    return new SearchByFuse<T>(config);
  }

  return new SearchByPrefixFilter<T>(config);
}
