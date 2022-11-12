package otf

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Unpack a .tar.gz byte stream to a directory
func Unpack(r io.Reader, dst string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to decompress archive: %w", err)
	}
	tr := tar.NewReader(gr)

	// Unpackage contents to the destination directory.
	for {
		header, err := tr.Next()
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

		// Ensure path's parent directories exist
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
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
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Open a handle to the destination.
		fh, err := os.Create(path)
		if err != nil {
			// This mimics tar's behavior wrt the tar file containing duplicate files
			// and it allowing later ones to clobber earlier ones even if the file
			// has perms that don't allow overwriting.
			if os.IsPermission(err) {
				os.Chmod(path, 0o600)
				fh, err = os.Create(path)
			}

			if err != nil {
				return fmt.Errorf("failed creating file %q: %w", path, err)
			}
		}

		// Copy the contents.
		_, err = io.Copy(fh, tr)
		// TODO: why are we closing it here?
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

// Pack a directory into tarball (.tar.gz) and return its contents
func Pack(src string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		target := info.Name()

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			target, err = os.Readlink(path)
			if err != nil {
				return nil
			}
		}

		hdr, err := tar.FileInfoHeader(info, target)
		if err != nil {
			return nil
		}
		name, err := filepath.Rel(src, path)
		if err != nil {
			return nil
		}
		hdr.Name = name

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		// Unless it's a normal file there's nothing more to be done
		if hdr.Typeflag != tar.TypeReg {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Write tar and gzip headers
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
