package agent

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Unpack a .tar.gz byte stream to a directory
func Unpack(r io.Reader, dst string) error {
	// Decompress as we read.
	decompressed, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to decompress archive: %w", err)
	}

	// Untar as we read.
	untar := tar.NewReader(decompressed)

	// Unpackage all the contents into the directory.
	for {
		header, err := untar.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to untar archive: %w", err)
		}

		// Get rid of absolute paths.
		path := header.Name
		if path[0] == '/' {
			path = path[1:]
		}
		path = filepath.Join(dst, path)

		// Make the directories to the path.
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// If we have a symlink, just link it.
		if header.Typeflag == tar.TypeSymlink {
			if err := os.Symlink(header.Linkname, path); err != nil {
				return fmt.Errorf("failed creating symlink %q => %q: %w",
					path, header.Linkname, err)
			}
			continue
		}

		// Only unpack regular files from this point on.
		if header.Typeflag == tar.TypeDir {
			continue
		} else if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			return fmt.Errorf("failed creating %q: unsupported type %c", path,
				header.Typeflag)
		}

		// Open a handle to the destination.
		fh, err := os.Create(path)
		if err != nil {
			// This mimics tar's behavior wrt the tar file containing duplicate files
			// and it allowing later ones to clobber earlier ones even if the file
			// has perms that don't allow overwriting.
			if os.IsPermission(err) {
				os.Chmod(path, 0600)
				fh, err = os.Create(path)
			}

			if err != nil {
				return fmt.Errorf("failed creating file %q: %w", path, err)
			}
		}

		// Copy the contents.
		_, err = io.Copy(fh, untar)
		fh.Close()
		if err != nil {
			return fmt.Errorf("failed to copy file %q: %w", path, err)
		}

		// Restore the file mode. We have to do this after writing the file,
		// since it is possible we have a read-only mode.
		mode := header.FileInfo().Mode()
		if err := os.Chmod(path, mode); err != nil {
			return fmt.Errorf("failed setting permissions on %q: %w", path, err)
		}
	}
	return nil
}
