package ioutil

import (
	"os"
	"path/filepath"
)

// ReplaceFile atomically installs srcPath at dstPath by renaming.
// On systems where os.Rename cannot overwrite an existing file (e.g. Windows),
// the destination is removed first and the rename is retried.
// srcPath is consumed by this call; it is the caller's responsibility to clean
// it up if an error is returned.
func ReplaceFile(srcPath, dstPath string) error {
	if err := os.Rename(srcPath, dstPath); err == nil {
		return nil
	}

	// Rename failed — check whether the destination exists.
	if _, statErr := os.Stat(dstPath); statErr != nil {
		if os.IsNotExist(statErr) {
			// Destination does not exist; return the original rename error.
			return os.Rename(srcPath, dstPath)
		}
		return statErr
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
