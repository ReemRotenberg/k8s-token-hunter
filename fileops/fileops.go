package fileops

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"k8s-token-hunter/model"
)

// mapFiles reads the directory specified by path and returns a slice of fs.DirEntry.
// It returns an error if reading the directory fails.
func MapFiles(path string) ([]fs.DirEntry, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err // Return an empty slice and the error
	}
	return files, nil // Return the files and a nil error
}

func GetFileContent(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

// copyFile is a function that copies the file from src to dst.
// It should return an error if the copying fails.
func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}
	return nil
}

// performBackUp creates a backup of the original file by copying it to a new path with "_backup" appended before the extension.
// It returns the backup file path and an error if the backup fails.
func PerformBackUp(originalPath string) (string, string, error) {
	ext := filepath.Ext(originalPath)
	base := strings.TrimSuffix(originalPath, ext)
	backupPath := base + "_backup" + ext

	err := CopyFile(originalPath, backupPath)
	if err != nil {
		return "", "", err // Return an empty string and the error
	}

	return originalPath, backupPath, nil // Return the backup path and a nil error
}

func PerformSymLink(src string, dst string) string {
	tmpSymlibk := dst + ".tmp"
	err := os.Symlink(src, tmpSymlibk)
	if err != nil {
		log.Fatal(err)
	}
	return tmpSymlibk
}

// restoreFile moves a file from src to dst.
func RestoreFile(src string, dst string) error {
	err := os.Rename(src, dst)
	if err != nil {
		return err
	}
	return nil
}

func SaveResultToFile(result []*model.Result, outputPath *string) {
	// Open the file for writing
	// If the file doesn't exist, create it, or append to the file
	file, err := os.OpenFile(*outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range result {
		_, err = fmt.Fprintf(file, "Pod name: %s\nNamespace: %s\nPath: %s\nToken: %s\n\n", r.Pod, r.Namespace, r.Path, r.Token[54:])
		if err != nil {
			log.Fatal(err)
		}
	}
	defer file.Close()
}
