package html

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
)

var _ http.FileSystem = (*cacheBuster)(nil)

// cacheBuster provides a cache-busting filesystem wrapper, mapping paths with a
// specific format containing a sha256 hash to paths without the hash in the
// wrapped filesystem. i.e. mapping
//
// /css/main.1fc822f99a2cfb6b5f316f00107a7d2770d547b64f3e0ea69baec12001a95f9f.css
// ->
// /css/main.css
type cacheBuster struct {
	fs.FS
}

// Open strips the hash from the name before opening it in the wrapped
// filesystem.
func (cb *cacheBuster) Open(name string) (http.File, error) {
	parts := strings.Split(name, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("expected two dots in path: %s", name)
	}

	// new name without hash
	name = fmt.Sprintf("%s.%s", parts[0], parts[2])

	return http.FS(cb.FS).Open(name)
}

// Path inserts a hash of the named file into its filename, before the filename
// extension: <path>.<ext> -> <path>.<hash>.<ext>, where <hash> is the hex
// format of the SHA256 hash of the contents of the file.
func (cb *cacheBuster) Path(name string) (string, error) {
	f, err := cb.FS.Open(name)
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

	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("expected one dot in path: %s", name)
	}

	nameWithHash := fmt.Sprintf("%s.%x.%s", parts[0], h.Sum(nil), parts[1])

	return nameWithHash, nil
}
