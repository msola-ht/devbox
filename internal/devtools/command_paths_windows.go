package devtools

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func platformCommandPaths(name string) []string {
	var paths []string
	add := func(dir string) {
		if filepath.IsAbs(dir) {
			paths = append(paths, filepath.Join(dir, name))
		}
	}
	for _, key := range []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`},
		{registry.CURRENT_USER, "Environment"},
	} {
		handle, err := registry.OpenKey(key.root, key.path, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		value, kind, err := handle.GetStringValue("Path")
		_ = handle.Close()
		if err != nil {
			continue
		}
		if kind == registry.EXPAND_SZ {
			value, err = registry.ExpandString(value)
			if err != nil {
				continue
			}
		}
		for _, dir := range filepath.SplitList(value) {
			add(dir)
		}
	}
	if dir := os.Getenv("LOCALAPPDATA"); filepath.IsAbs(dir) {
		add(filepath.Join(dir, "Microsoft", "WinGet", "Links"))
		add(filepath.Join(dir, "Microsoft", "WindowsApps"))
	}
	if dir := os.Getenv("ProgramFiles"); filepath.IsAbs(dir) {
		add(filepath.Join(dir, "WinGet", "Links"))
	}
	if name == "python.exe" {
		paths = append(paths, registeredWindowsPython()...)
	}
	return paths
}

func registeredWindowsPython() []string {
	var paths []string
	for _, entry := range []struct {
		root registry.Key
		view uint32
	}{
		{registry.CURRENT_USER, 0},
		{registry.LOCAL_MACHINE, registry.WOW64_64KEY},
		{registry.LOCAL_MACHINE, registry.WOW64_32KEY},
	} {
		const company = `Software\Python\PythonCore`
		key, err := registry.OpenKey(entry.root, company, registry.ENUMERATE_SUB_KEYS|entry.view)
		if err != nil {
			continue
		}
		tags, _ := key.ReadSubKeyNames(-1)
		_ = key.Close()
		for _, tag := range tags {
			key, err := registry.OpenKey(entry.root, company+`\`+tag+`\InstallPath`, registry.QUERY_VALUE|entry.view)
			if err != nil {
				continue
			}
			executable, _, _ := key.GetStringValue("ExecutablePath")
			_ = key.Close()
			if filepath.IsAbs(executable) {
				paths = append(paths, executable)
			}
		}
	}
	return paths
}
