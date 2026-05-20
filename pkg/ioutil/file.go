package ioutil

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// ReplaceFile installs srcPath at dstPath by renaming.
// On platforms where os.Rename atomically replaces an existing file (e.g. Unix),
// the replacement is atomic. On platforms where os.Rename cannot overwrite an
// existing file (e.g. Windows), the destination is removed first and the rename
// is retried; this is best-effort and not atomic — dstPath may be briefly absent
// or left missing if the rename fails after the remove.
// srcPath is consumed by this call; it is the caller's responsibility to clean
// it up if an error is returned.
func ReplaceFile(srcPath, dstPath string) error {
	renameErr := os.Rename(srcPath, dstPath)
	if renameErr == nil {
		return nil
	}

	// Only attempt remove+retry when the rename failed specifically because
	// the destination already exists (e.g. Windows does not allow os.Rename
	// to overwrite an existing file). For any other failure (missing source,
	// permission denied, cross-device move, etc.) return the original error
	// without touching dstPath to avoid data loss.
	if !errors.Is(renameErr, os.ErrExist) {
		return renameErr
	}

	if removeErr := os.Remove(dstPath); removeErr != nil && !os.IsNotExist(removeErr) {
		return removeErr
	}

	return os.Rename(srcPath, dstPath)
}

// atomicWriteFile is the shared core for WriteFileAtomic and WriteFileAtomicFrom.
// It creates a sibling temp file in the same directory as dstPath, resolves
// permissions (preserving existing ones when dstPath exists), calls fn to
// write data into the open file, then replaces dstPath via ReplaceFile.
// The temp file is always removed on failure.
func atomicWriteFile(dstPath string, defaultPerm os.FileMode, fn func(*os.File) (int64, error)) (int64, error) {
	perm := defaultPerm
	if info, err := os.Stat(dstPath); err == nil {
		perm = info.Mode().Perm()
	}

	tmp, err := os.CreateTemp(filepath.Dir(dstPath), ".atomic-write-*")
	if err != nil {
		return 0, err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return 0, err
	}

	n, fnErr := fn(tmp)
	closeErr := tmp.Close()

	if fnErr != nil {
		return 0, fnErr
	}
	if closeErr != nil {
		return 0, closeErr
	}

	if err := ReplaceFile(tmpPath, dstPath); err != nil {
		return 0, err
	}

	return n, nil
}

// WriteFileAtomic writes content to dstPath atomically by first writing to a
// temporary file in the same directory, then replacing dstPath via ReplaceFile.
// If dstPath already exists its permissions are preserved; otherwise defaultPerm
// is used for the new file.
func WriteFileAtomic(dstPath string, content []byte, defaultPerm os.FileMode) error {
	_, err := atomicWriteFile(dstPath, defaultPerm, func(f *os.File) (int64, error) {
		n, err := f.Write(content)
		return int64(n), err
	})
	return err
}

// WriteFileAtomicFrom writes the contents of r to dstPath atomically by first
// writing to a sibling temp file, then replacing dstPath via ReplaceFile.
// The temp file is always removed on failure.
// If dstPath already exists its permissions are preserved; otherwise defaultPerm
// is used. On Unix, Chmod is applied to the still-open temp file descriptor
// before writing: because permission checks occur at open time the write
// succeeds even when the resolved perm has no write bit set.
// It returns the number of bytes written.
func WriteFileAtomicFrom(dstPath string, r io.Reader, defaultPerm os.FileMode) (int64, error) {
	return atomicWriteFile(dstPath, defaultPerm, func(f *os.File) (int64, error) {
		return io.Copy(f, r)
	})
}
