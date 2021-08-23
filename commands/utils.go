// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/isacikgoz/prompt"
	"github.com/pkg/errors"
)

func checkInteractiveTerminal() error {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return err
	}

	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return errors.New("this is not an interactive shell")
	}

	return nil
}

func zipDir(zipPath, dir string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("cannot create file %q: %w", zipPath, err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	if err := addToZip(zipWriter, dir, "."); err != nil {
		return fmt.Errorf("could not add %q to zip: %w", dir, err)
	}

	return nil
}

func getConfirmation(question string, dbConfirmation bool) error {
	if err := checkInteractiveTerminal(); err != nil {
		return fmt.Errorf("could not proceed, either enable --confirm flag or use an interactive shell to complete operation: %w", err)
	}

	if dbConfirmation {
		s, err := prompt.NewSelection("Have you performed a database backup?", []string{"no", "yes"}, "", 2)
		if err != nil {
			return fmt.Errorf("could not initiate prompt: %w", err)
		}
		ans, err := s.Run()
		if err != nil {
			return fmt.Errorf("error running prompt: %w", err)
		}
		if ans != "yes" {
			return errors.New("aborted")
		}
	}

	s, err := prompt.NewSelection(question, []string{"no", "yes"}, "WARNING: This operation is not reversible.", 2)
	if err != nil {
		return fmt.Errorf("could not initiate prompt: %w", err)
	}
	ans, err := s.Run()
	if err != nil {
		return fmt.Errorf("error running prompt: %w", err)
	}
	if ans != "yes" {
		return errors.New("aborted")
	}

	return nil
}

func addToZip(zipWriter *zip.Writer, basedir, path string) error {
	dirPath := filepath.Join(basedir, path)
	fileInfos, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("cannot read directory %q: %w", dirPath, err)
	}

	for _, fileInfo := range fileInfos {
		filePath := filepath.Join(path, fileInfo.Name())
		if fileInfo.IsDir() {
			filePath += "/"
		}
		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return fmt.Errorf("cannot create zip file info header for %q path: %w", filePath, err)
		}
		header.Name = filePath
		header.Method = zip.Deflate

		w, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("cannot create header for path %q: %w", filePath, err)
		}

		if fileInfo.IsDir() {
			if err = addToZip(zipWriter, basedir, filePath); err != nil {
				return err
			}
			continue
		}

		file, err := os.Open(filepath.Join(dirPath, fileInfo.Name()))
		if err != nil {
			return fmt.Errorf("cannot open file %q: %w", filePath, err)
		}

		_, err = io.Copy(w, file)
		file.Close()
		if err != nil {
			return fmt.Errorf("cannot zip file contents for file %q: %w", filePath, err)
		}
	}

	return nil
}
