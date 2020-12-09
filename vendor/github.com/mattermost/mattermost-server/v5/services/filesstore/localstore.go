// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package filesstore

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils"
)

const (
	TEST_FILE_PATH = "/testfile"
)

type LocalFileBackend struct {
	directory string
}

func (b *LocalFileBackend) TestConnection() *model.AppError {
	f := bytes.NewReader([]byte("testingwrite"))
	if _, err := writeFileLocally(f, filepath.Join(b.directory, TEST_FILE_PATH)); err != nil {
		return model.NewAppError("TestFileConnection", "api.file.test_connection.local.connection.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	os.Remove(filepath.Join(b.directory, TEST_FILE_PATH))
	mlog.Debug("Able to write files to local storage.")
	return nil
}

func (b *LocalFileBackend) Reader(path string) (ReadCloseSeeker, *model.AppError) {
	f, err := os.Open(filepath.Join(b.directory, path))
	if err != nil {
		return nil, model.NewAppError("Reader", "api.file.reader.reading_local.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return f, nil
}

func (b *LocalFileBackend) ReadFile(path string) ([]byte, *model.AppError) {
	f, err := ioutil.ReadFile(filepath.Join(b.directory, path))
	if err != nil {
		return nil, model.NewAppError("ReadFile", "api.file.read_file.reading_local.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return f, nil
}

func (b *LocalFileBackend) FileExists(path string) (bool, *model.AppError) {
	_, err := os.Stat(filepath.Join(b.directory, path))

	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, model.NewAppError("ReadFile", "api.file.file_exists.exists_local.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return true, nil
}

func (b *LocalFileBackend) FileSize(path string) (int64, *model.AppError) {
	info, err := os.Stat(filepath.Join(b.directory, path))
	if err != nil {
		return 0, model.NewAppError("FileSize", "api.file.file_size.local.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return info.Size(), nil
}

func (b *LocalFileBackend) CopyFile(oldPath, newPath string) *model.AppError {
	if err := utils.CopyFile(filepath.Join(b.directory, oldPath), filepath.Join(b.directory, newPath)); err != nil {
		return model.NewAppError("copyFile", "api.file.move_file.rename.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (b *LocalFileBackend) MoveFile(oldPath, newPath string) *model.AppError {
	if err := os.MkdirAll(filepath.Dir(filepath.Join(b.directory, newPath)), 0750); err != nil {
		return model.NewAppError("moveFile", "api.file.move_file.rename.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if err := os.Rename(filepath.Join(b.directory, oldPath), filepath.Join(b.directory, newPath)); err != nil {
		return model.NewAppError("moveFile", "api.file.move_file.rename.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (b *LocalFileBackend) WriteFile(fr io.Reader, path string) (int64, *model.AppError) {
	return writeFileLocally(fr, filepath.Join(b.directory, path))
}

func writeFileLocally(fr io.Reader, path string) (int64, *model.AppError) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		directory, _ := filepath.Abs(filepath.Dir(path))
		return 0, model.NewAppError("WriteFile", "api.file.write_file_locally.create_dir.app_error", nil, "directory="+directory+", err="+err.Error(), http.StatusInternalServerError)
	}
	fw, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return 0, model.NewAppError("WriteFile", "api.file.write_file_locally.writing.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	defer fw.Close()
	written, err := io.Copy(fw, fr)
	if err != nil {
		return written, model.NewAppError("WriteFile", "api.file.write_file_locally.writing.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return written, nil
}

func (b *LocalFileBackend) AppendFile(fr io.Reader, path string) (int64, *model.AppError) {
	fp := filepath.Join(b.directory, path)
	if _, err := os.Stat(fp); err != nil {
		return 0, model.NewAppError("AppendFile", "api.file.append_file.no_exist.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	fw, err := os.OpenFile(fp, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, model.NewAppError("AppendFile", "api.file.append_file.opening.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	defer fw.Close()
	written, err := io.Copy(fw, fr)
	if err != nil {
		return written, model.NewAppError("AppendFile", "api.file.append_file.writing.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return written, nil
}

func (b *LocalFileBackend) RemoveFile(path string) *model.AppError {
	if err := os.Remove(filepath.Join(b.directory, path)); err != nil {
		return model.NewAppError("RemoveFile", "utils.file.remove_file.local.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (b *LocalFileBackend) ListDirectory(path string) (*[]string, *model.AppError) {
	var paths []string
	fileInfos, err := ioutil.ReadDir(filepath.Join(b.directory, path))
	if err != nil {
		if os.IsNotExist(err) {
			return &paths, nil
		}
		return nil, model.NewAppError("ListDirectory", "utils.file.list_directory.local.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	for _, fileInfo := range fileInfos {
		paths = append(paths, filepath.Join(path, fileInfo.Name()))
	}
	return &paths, nil
}

func (b *LocalFileBackend) RemoveDirectory(path string) *model.AppError {
	if err := os.RemoveAll(filepath.Join(b.directory, path)); err != nil {
		return model.NewAppError("RemoveDirectory", "utils.file.remove_directory.local.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return nil
}
