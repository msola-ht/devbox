package devtools

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cavaliergopher/grab/v3"
)

const uvReleaseAPI = "https://api.github.com/repos/astral-sh/uv/releases/latest"
const uvReleaseDownload = "https://github.com/astral-sh/uv/releases/download/"

type uvReleaseAsset struct {
	Name   string `json:"name"`
	URL    string `json:"browser_download_url"`
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
}

func latestUVAsset(ctx context.Context, client *http.Client, name string) (uvReleaseAsset, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uvReleaseAPI, nil)
	if err != nil {
		return uvReleaseAsset{}, "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	response, err := client.Do(req)
	if err != nil {
		return uvReleaseAsset{}, "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return uvReleaseAsset{}, "", fmt.Errorf("GitHub 返回 HTTP %d", response.StatusCode)
	}
	var release struct {
		Tag    string           `json:"tag_name"`
		Assets []uvReleaseAsset `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&release); err != nil {
		return uvReleaseAsset{}, "", err
	}
	for _, asset := range release.Assets {
		if asset.Name == name {
			return asset, release.Tag, nil
		}
	}
	return uvReleaseAsset{}, "", fmt.Errorf("最新版本没有提供 %s", name)
}

// An empty result means the project already has the latest release.
func (i *toolInstallation) downloadUVRelease(ctx context.Context, client *http.Client, directory, currentVersion string, job *toolInstallJob) (string, error) {
	var arch string
	switch runtime.GOARCH {
	case "arm64":
		arch = "aarch64"
	case "amd64":
		arch = "x86_64"
	default:
		return "", fmt.Errorf("当前架构 %s 不支持 uv 安装", runtime.GOARCH)
	}
	target := "apple-darwin"
	if runtime.GOOS == "linux" {
		target = "unknown-linux-gnu"
	}
	name := "uv-" + arch + "-" + target + ".tar.gz"
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "https" || len(via) >= 10 {
			return errors.New("uv 下载的重定向无效")
		}
		return nil
	}
	i.updateMessage(job, "正在查询 uv 最新版本…")
	asset, tag, err := latestUVAsset(ctx, client, name)
	if err != nil {
		return "", fmt.Errorf("无法查询 uv 最新版本：%w", err)
	}
	if current, _, _ := strings.Cut(currentVersion, " "); current != "" && current == tag {
		return "", nil
	}
	digest, hasSHA256 := strings.CutPrefix(asset.Digest, "sha256:")
	sum, err := hex.DecodeString(digest)
	if !hasSHA256 || err != nil || len(sum) != sha256.Size {
		return "", errors.New("uv 发行包缺少有效的 SHA-256 校验值")
	}
	if !strings.HasPrefix(asset.URL, uvReleaseDownload) || asset.Size <= 0 {
		return "", errors.New("uv 发行包的下载信息无效")
	}
	archive := filepath.Join(directory, name)
	req, err := grab.NewRequest(archive, asset.URL)
	if err != nil {
		return "", err
	}
	req.Size = asset.Size
	req.SetChecksum(sha256.New(), sum, true)
	req.AfterCopy = func(*grab.Response) error {
		percent := 100
		i.updateProgress(job, "正在校验 uv 发行包…", &percent)
		return nil
	}
	downloader := grab.NewClient()
	downloader.HTTPClient = client
	downloadCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	const downloading = "正在下载 uv…"
	percent := 0
	i.updateProgress(job, downloading, &percent)
	response := downloader.Do(req.WithContext(downloadCtx))
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-response.Done:
			if err := response.Err(); err != nil {
				return "", fmt.Errorf("无法下载或校验 uv 发行包：%w", err)
			}
			return archive, nil
		case <-ticker.C:
			if received := response.BytesComplete(); received < asset.Size {
				percent := int(float64(received) / float64(asset.Size) * 100)
				i.updateProgress(job, downloading, &percent)
			}
		}
	}
}

func extractUVRelease(ctx context.Context, archive, bin string) error {
	file, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer compressed.Close()
	if err := os.Mkdir(bin, 0700); err != nil {
		return err
	}

	directory := strings.TrimSuffix(filepath.Base(archive), ".tar.gz")
	remaining := map[string]string{directory + "/uv": "uv", directory + "/uvx": "uvx"}
	reader := tar.NewReader(compressed)
	for len(remaining) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		entry, err := reader.Next()
		if err == io.EOF {
			return errors.New("uv 发行包缺少 uv 或 uvx")
		}
		if err != nil {
			return err
		}
		name, needed := remaining[entry.Name]
		if !needed {
			continue
		}
		if entry.Typeflag != tar.TypeReg || entry.Size <= 0 || entry.Size > maxUVBinarySize {
			return fmt.Errorf("发行包中的 %s 不是可用的工具文件", name)
		}
		output, err := os.OpenFile(filepath.Join(bin, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0700)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(output, reader)
		if err := errors.Join(copyErr, output.Close()); err != nil {
			return err
		}
		delete(remaining, entry.Name)
	}
	return nil
}
