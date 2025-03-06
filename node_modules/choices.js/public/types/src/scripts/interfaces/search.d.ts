export interface SearchResult<T extends object> {
    item: T;
    score: number;
    rank: number;
}
export interface Searcher<T extends object> {
    reset(): void;
    isEmptyIndex(): boolean;
    index(data: T[]): void;
    search(needle: string): SearchResult<T>[];
}
