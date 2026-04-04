package ioutil

import (
	"errors"
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

// WriteFileAtomic writes content to dstPath atomically by first writing to a
// temporary file in the same directory, then replacing dstPath via ReplaceFile.
// If dstPath already exists its permissions are preserved; otherwise defaultPerm
// is used for the new file.
func WriteFileAtomic(dstPath string, content []byte, defaultPerm os.FileMode) error {
	perm := defaultPerm
	if info, err := os.Stat(dstPath); err == nil {
		perm = info.Mode().Perm()
	}

	tmp, err := os.CreateTemp(filepath.Dir(dstPath), ".atomic-write-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return ReplaceFile(tmpPath, dstPath)
}
