export interface SearchResult<T extends object> {
  item: T;
  score: number;
  rank: number; // values of 0 means this item is not in the search-result set, and should be discarded
}

export interface Searcher<T extends object> {
  reset(): void;
  isEmptyIndex(): boolean;
  index(data: T[]): void;
  search(needle: string): SearchResult<T>[];
}
