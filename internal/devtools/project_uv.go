package devtools

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"devbox/internal/projectdir"
)

const uvRecordName = "install.json"
const maxUVBinarySize = 128 << 20

type projectUVRecord struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Version string `json:"version"`
	UV      string `json:"uv"`
	UVX     string `json:"uvx"`
}

type projectUVSnapshot struct {
	record       projectUVRecord
	recordDigest [sha256.Size]byte
}

func (s projectUVSnapshot) compatible() bool {
	return s.record.OS == runtime.GOOS && s.record.Arch == runtime.GOARCH
}

func executableDigest(ctx context.Context, root *os.Root, name string) (string, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 || info.Size() > maxUVBinarySize {
		return "", fmt.Errorf("%s 不是可用的工具文件", name)
	}
	file, err := root.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	reader := io.LimitReader(file, maxUVBinarySize+1)
	buffer := make([]byte, 64<<10)
	var size int64
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		n, err := reader.Read(buffer)
		hash.Write(buffer[:n])
		size += int64(n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	if size != info.Size() {
		return "", fmt.Errorf("%s 在读取时发生变化", name)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// DetectProjectUV checks the selected project's recorded uv installation.
func (s *Service) DetectProjectUV(ctx context.Context, project string, _ bool) ([]ToolDetection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, err := projectdir.Open(project)
	if err != nil {
		message := fmt.Sprintf("无法访问工作目录（%v）。", err)
		if errors.Is(err, fs.ErrNotExist) {
			message = "工作目录已不存在。请在概览打开或恢复默认目录，或选择其他目录。"
		}
		return []ToolDetection{{Status: statusProjectUnavailable, Message: message}}, ctx.Err()
	}
	defer root.Close()
	result := readProjectUV(ctx, root)
	return []ToolDetection{result}, ctx.Err()
}

func readProjectUV(ctx context.Context, root *os.Root) ToolDetection {
	tools, err := projectdir.Tools(root, false)
	if err == nil {
		defer tools.Close()
		var target *os.Root
		target, err = projectdir.OpenChild(tools, "uv", false)
		if err == nil {
			defer target.Close()
			path := filepath.Join(target.Name(), "uv")
			snapshot, err := inspectProjectUV(ctx, target)
			if err != nil {
				return ToolDetection{Status: statusFailed, Path: path, Message: err.Error()}
			}
			if !snapshot.compatible() {
				return ToolDetection{
					Status: statusIncompatible, Path: path, Version: snapshot.record.Version,
					Message: fmt.Sprintf("已安装的 uv %s 适用于 %s/%s，当前为 %s/%s。",
						snapshot.record.Version, snapshot.record.OS, snapshot.record.Arch, runtime.GOOS, runtime.GOARCH),
				}
			}
			return ToolDetection{Status: statusAvailable, Path: path, Version: snapshot.record.Version}
		}
	}
	if errors.Is(err, fs.ErrNotExist) {
		return ToolDetection{Status: statusNotFound, Message: "当前项目未安装 uv。"}
	}
	return ToolDetection{Status: statusFailed, Message: fmt.Sprintf("无法读取项目工具目录（%v）。", err)}
}

func inspectProjectUV(ctx context.Context, root *os.Root) (projectUVSnapshot, error) {
	const invalid = "项目 uv 的安装记录缺失或文件校验失败。"
	info, err := root.Lstat(uvRecordName)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 4096 {
		return projectUVSnapshot{}, errors.New(invalid)
	}
	data, err := root.ReadFile(uvRecordName)
	var record projectUVRecord
	if err != nil || json.Unmarshal(data, &record) != nil || record.Version == "" || record.OS == "" || record.Arch == "" {
		return projectUVSnapshot{}, errors.New(invalid)
	}
	uv, err := executableDigest(ctx, root, "uv")
	if err != nil || uv != record.UV {
		return projectUVSnapshot{}, errors.New(invalid)
	}
	uvx, err := executableDigest(ctx, root, "uvx")
	if err != nil || uvx != record.UVX {
		return projectUVSnapshot{}, errors.New(invalid)
	}
	return projectUVSnapshot{record: record, recordDigest: sha256.Sum256(data)}, nil
}
