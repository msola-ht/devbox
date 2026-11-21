package projectdir

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Default returns the default workspace path.
func Default() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(home) {
		return "", errors.New("无法确定用户主目录的绝对路径")
	}
	path := filepath.Join(home, ".heforge-devbox")
	resolved, err := filepath.EvalSymlinks(path)
	if errors.Is(err, os.ErrNotExist) {
		return path, nil
	}
	return resolved, err
}

// EnsureDefault creates the workspace root when needed.
func EnsureDefault() (string, error) {
	path, err := Default()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}

// Open opens a confirmed, canonical project directory.
func Open(project string) (*os.Root, error) {
	if !filepath.IsAbs(project) {
		return nil, errors.New("请选择项目工作目录。")
	}
	resolved, err := filepath.EvalSymlinks(project)
	if err != nil {
		return nil, err
	}
	if resolved != project {
		return nil, errors.New("项目路径已变化，请重新选择目录。")
	}
	return os.OpenRoot(project)
}

// OpenChild opens a normal child directory and optionally creates it.
func OpenChild(root *os.Root, name string, create bool) (*os.Root, error) {
	if create {
		if err := root.Mkdir(name, 0755); err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, err
		}
	}
	info, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s 必须是普通目录", name)
	}
	return root.OpenRoot(name)
}

// Tools opens or creates the project's managed tool directory.
func Tools(root *os.Root, create bool) (*os.Root, error) {
	config, err := OpenChild(root, ".devbox", create)
	if err != nil {
		return nil, err
	}
	defer config.Close()
	return OpenChild(config, "tools", create)
}
