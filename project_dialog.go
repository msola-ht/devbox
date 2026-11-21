package main

import (
	"path/filepath"

	"devbox/internal/projectdir"
	"github.com/egoist/mygo"
)

type projectDialog struct{}

func (projectDialog) DefaultDirectory() (string, error) {
	return projectdir.Default()
}

func (projectDialog) CreateDefaultDirectory() (string, error) {
	return projectdir.EnsureDefault()
}

func (projectDialog) ValidateDirectory(path string) error {
	root, err := projectdir.Open(path)
	if err != nil {
		return err
	}
	return root.Close()
}

func (projectDialog) OpenDefaultDirectory() error {
	path, err := projectdir.EnsureDefault()
	if err != nil {
		return err
	}
	return mygo.Shell.OpenPath(path)
}

func (projectDialog) OpenDirectory(path string) error {
	root, err := projectdir.Open(path)
	if err != nil {
		return err
	}
	defer root.Close()
	return mygo.Shell.OpenPath(root.Name())
}

func (projectDialog) SelectDirectory() (string, error) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Title: "选择项目工作目录", Directory: true, CreateDirectories: true})
	if err != nil || len(paths) == 0 {
		return "", err
	}
	return filepath.EvalSymlinks(paths[0])
}
