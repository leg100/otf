package html

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	_ http.FileSystem = (*CacheBuster)(nil)

	// regexp for a hex-formatted sha256 hash sum
	sha256re = regexp.MustCompile(`[0-9a-f]{32}`)
)

// CacheBuster provides a cache-busting filesystem wrapper, mapping paths with a
// specific format containing a sha256 hash to paths without the hash in the
// wrapped filesystem. i.e. mapping
//
// /css/main.1fc822f99a2cfb6b5f316f00107a7d2770d547b64f3e0ea69baec12001a95f9f.css
// ->
// /css/main.css
type CacheBuster struct {
	fs.FS
}

// Open strips the hash from the name before opening it in the wrapped
// filesystem.
func (cb *CacheBuster) Open(fname string) (http.File, error) {
	var partsSansHash []string

	// Reconstruct filename without hash
	parts := strings.SplitSeq(fname, ".")
	for p := range parts {
		if !sha256re.MatchString(p) {
			partsSansHash = append(partsSansHash, p)
		}
	}

	// new name without hash
	fname = strings.Join(partsSansHash, ".")

	return http.FS(cb.FS).Open(fname)
}

// Path inserts a hash of the named file into its filename, before the filename
// extension: <path>.<ext> -> <path>.<hash>.<ext>, where <hash> is the hex
// format of the SHA256 hash of the contents of the file.
func (cb *CacheBuster) Path(fname string) (string, error) {
	var leadingSlash bool

	// fs.FS expects paths without a leading slash
	if strings.HasPrefix(fname, "/") {
		leadingSlash = true
		fname = strings.TrimPrefix(fname, "/")
	}

	f, err := cb.FS.Open(fname)
	if err != nil {
		return "", err
	}

	// TODO: this is an expensive operation to perform if this method is to be
	// called everytime a template is rendered; consider caching result.
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}

	nameWithoutExt, ext := splitFilenameOnExt(fname)
	nameWithHash := fmt.Sprintf("%s.%x%s", nameWithoutExt, h.Sum(nil), ext)

	if leadingSlash {
		nameWithHash = "/" + nameWithHash
	}

	return nameWithHash, nil
}

func splitFilenameOnExt(fname string) (string, string) {
	ext := filepath.Ext(fname)
	nameWithoutExt := strings.TrimSuffix(fname, ext)

	return nameWithoutExt, ext
}
