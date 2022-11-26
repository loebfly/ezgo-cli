package tools

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func File(filePath string) fileT {
	return fileT{path: filePath}
}

type fileT struct {
	path string
}

func (receiver fileT) UnzipTo(targetPath string) error {
	reader, err := zip.OpenReader(receiver.path)
	if err != nil {
		return err
	}
	defer func(reader *zip.ReadCloser) {
		_ = reader.Close()
	}(reader)

	for _, file := range reader.File {
		err = receiver.unZipFile(file, targetPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (receiver fileT) unZipFile(file *zip.File, targetPath string) error {
	filePath := path.Join(targetPath, file.Name)
	if file.FileInfo().IsDir() {
		err := os.MkdirAll(filePath, file.Mode())
		if err != nil {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(path.Dir(filePath), file.Mode()); err != nil {
		return err
	}

	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer func(rc io.ReadCloser) {
		_ = rc.Close()
	}(rc)

	w, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func(w io.WriteCloser) {
		_ = w.Close()
	}(w)

	_, err = io.Copy(w, rc)
	if err != nil {
		return err
	}
	return nil
}

// MoveDirSubShowFilesTo 移动目录下的所有文件到目标目录
func (receiver fileT) MoveDirSubShowFilesTo(targetDir string) error {
	fis, err := ioutil.ReadDir(receiver.path)
	if err != nil {
		return err
	}
	for _, file := range fis {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		fromPath := path.Join(receiver.path, file.Name())
		filePath := path.Join(targetDir, file.Name())
		err = os.Rename(fromPath, filePath)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteDirSubFiles 删除目录下的所有文件
func (receiver fileT) DeleteDirSubFiles() error {
	fis, err := ioutil.ReadDir(receiver.path)
	if err != nil {
		return err
	}
	for _, file := range fis {
		filePath := path.Join(receiver.path, file.Name())
		err = os.RemoveAll(filePath)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadString 读取文件内容到字符串
func (receiver fileT) ReadString() (string, error) {
	b, err := ioutil.ReadFile(receiver.path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// WriteString 写入字符串到文件
func (receiver fileT) WriteString(content string) error {
	return ioutil.WriteFile(receiver.path, []byte(content), os.ModePerm)
}

// Exists 判断文件是否存在
func (receiver fileT) Exists() bool {
	_, err := os.Stat(receiver.path)
	return err == nil || os.IsExist(err)
}
