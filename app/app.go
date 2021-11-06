/*
Package app implements services, co-ordinating between the layers of the project.
*/
package app

import (
	"context"

	"github.com/leg100/otf"
)

// cacheLogChunk not only caches the log chunk but ensures the cache is not
// missing any previous chunks. If the chunk is the first chunk then nothing
// more is done. Otherwise it'll retrieve all chunks from the cache and if
// they're missing, it'll repopulate the cache from the backend store using fn.
func cacheLogChunk(ctx context.Context, cache otf.Cache, store otf.ChunkStore, id string, chunk []byte, start bool) error {
	// if first chunk then just set in cache and exit
	if start {
		return cache.Set(otf.LogCacheKey(id), chunk)
	}

	// retrieve logs from cache. if there is a hit then append chunk. else get
	// logs from db and cache that.

	if _, err := cache.Get(otf.LogCacheKey(id)); err == nil {
		return cache.AppendChunk(otf.LogCacheKey(id), chunk)
	}

	// cache is empty, repopulate it with all chunks
	logs, err := store.GetChunk(ctx, id, otf.GetChunkOptions{})
	if err != nil {
		return err
	}
	if err := cache.Set(otf.LogCacheKey(id), logs); err != nil {
		return err
	}

	return nil
}
