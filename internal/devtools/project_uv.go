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

type projectUVRecord struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Version string `json:"version"`
	UV      string `json:"uv"`
	UVX     string `json:"uvx"`
}

func executableDigest(ctx context.Context, root *os.Root, name string) (string, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return "", err
	}
	const limit = 128 << 20
	if !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 || info.Size() > limit {
		return "", fmt.Errorf("%s 不是可用的工具文件", name)
	}
	file, err := root.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	reader := io.LimitReader(file, limit+1)
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
		return nil, err
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
			return readProjectUVRecord(ctx, target)
		}
	}
	if errors.Is(err, fs.ErrNotExist) {
		return ToolDetection{Status: statusNotFound, Message: "当前项目未安装 uv。"}
	}
	return ToolDetection{Status: statusFailed, Message: fmt.Sprintf("无法读取项目工具目录（%v）。", err)}
}

func readProjectUVRecord(ctx context.Context, root *os.Root) ToolDetection {
	path := filepath.Join(root.Name(), "uv")
	failure := ToolDetection{Status: statusFailed, Path: path, Message: "项目 uv 的安装记录缺失或文件校验失败。"}
	info, err := root.Lstat(uvRecordName)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 4096 {
		return failure
	}
	data, err := root.ReadFile(uvRecordName)
	var record projectUVRecord
	if err != nil || json.Unmarshal(data, &record) != nil || record.Version == "" {
		return failure
	}
	if record.OS != runtime.GOOS || record.Arch != runtime.GOARCH {
		failure.Message = "项目 uv 的安装平台与当前系统或架构不匹配。"
		return failure
	}
	uv, err := executableDigest(ctx, root, "uv")
	if err != nil || uv != record.UV {
		return failure
	}
	uvx, err := executableDigest(ctx, root, "uvx")
	if err != nil || uvx != record.UVX {
		return failure
	}
	return ToolDetection{Status: statusAvailable, Path: path, Version: record.Version}
}
