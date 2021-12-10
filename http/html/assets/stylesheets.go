package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"net/url"
)

// CacheBustingPaths walks fs for files matching pattern, such as CSS files, and
// returns a list of cache-busting relative URLs for each file. A query string
// is appended to each path, v=hash, where hash is the SHA256 of the contents of
// the file.
func CacheBustingPaths(root fs.FS, pattern string) ([]string, error) {
	var paths []string

	matches, err := fs.Glob(root, pattern)
	if err != nil {
		return nil, err
	}

	for _, m := range matches {
		f, err := root.Open(m)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		h := sha256.New()
		_, err = io.Copy(h, f)
		if err != nil {
			return nil, err
		}

		u := url.URL{Path: m}

		q := u.Query()
		q.Add("v", hex.EncodeToString(h.Sum(nil)))
		u.RawQuery = q.Encode()

		paths = append(paths, u.String())
	}

	return paths, nil
}
