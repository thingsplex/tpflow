package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func BackupDirectory(sourceDir, targetDir, prefix string) error {
	dir, err := os.Open(sourceDir)

	if err != nil {
		return err
	}

	defer dir.Close()

	destFile := fmt.Sprintf("%s/%s_%s.tar.gz", targetDir, prefix, time.Now().Format("2006_01_02T15_04_05"))
	tarfile, err := os.Create(destFile)

	if err != nil {
		return err
	}

	defer tarfile.Close()
	var fileWriter io.WriteCloser = tarfile

	fileWriter = gzip.NewWriter(tarfile) // add a gzip filter
	defer fileWriter.Close()             // if user add .gz in the destination filename

	tarfileWriter := tar.NewWriter(fileWriter)
	defer tarfileWriter.Close()

	filepath.Walk(sourceDir, func(file string, fi os.FileInfo, err error) error {
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(strings.Replace(file, sourceDir, "", -1), string(filepath.Separator))
		// write header
		if err := tarfileWriter.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tarfileWriter, data); err != nil {
				data.Close()
				return err
			}
			data.Close()
		}
		return nil
	})

	return nil

}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}