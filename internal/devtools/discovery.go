package devtools

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type toolCandidate struct {
	path string
	err  error
}

func detectTools(
	ctx context.Context,
	name, notFoundMessage string,
	probe func(context.Context, string) ToolDetection,
	extraPaths ...string,
) ([]ToolDetection, error) {
	probeCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	candidates := findToolCandidates(probeCtx, name, extraPaths)
	if err := probeCtx.Err(); err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return []ToolDetection{{Status: statusNotFound, Message: notFoundMessage}}, nil
	}

	results := make([]ToolDetection, len(candidates))
	workers := min(len(candidates), 4)
	var probes sync.WaitGroup
	for worker := range workers {
		probes.Go(func() {
			for index := worker; index < len(candidates) && probeCtx.Err() == nil; index += workers {
				candidate := candidates[index]
				if candidate.err != nil {
					results[index] = pathFailure(candidate.path, candidate.err)
				} else {
					results[index] = probe(probeCtx, candidate.path)
				}
			}
		})
	}
	probes.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for index, result := range results {
		if result.Status == "" {
			results[index] = ToolDetection{
				Status: statusFailed, Path: candidates[index].path,
				Message: "检测已达时间上限，尚未探测此入口。",
			}
		}
	}
	return uniqueToolResults(results), nil
}

func findToolCandidates(ctx context.Context, name string, extraPaths []string) []toolCandidate {
	var candidates []toolCandidate
	seen := make(map[string]bool)
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if ctx.Err() != nil {
			return nil
		}
		if filepath.IsAbs(dir) {
			path, err := exec.LookPath(filepath.Join(dir, name))
			if err == nil && !seen[path] {
				candidates = append(candidates, toolCandidate{path: path})
				seen[path] = true
			}
		}
	}

	for _, path := range extraPaths {
		if ctx.Err() != nil {
			return nil
		}
		if seen[path] {
			continue
		}
		_, err := exec.LookPath(path)
		if !errors.Is(err, fs.ErrNotExist) {
			candidates = append(candidates, toolCandidate{path: path, err: err})
			seen[path] = true
		}
	}

	if len(candidates) == 0 && ctx.Err() == nil {
		if path, err := findInPath(name); err != nil {
			candidates = append(candidates, toolCandidate{path: path, err: err})
		}
	}
	return candidates
}

func findInPath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if errors.Is(err, exec.ErrNotFound) {
		return "", nil
	}
	if err == nil && !filepath.IsAbs(path) {
		return path, exec.ErrDot
	}
	return path, err
}

func uniqueToolResults(results []ToolDetection) []ToolDetection {
	type resultKey struct {
		path, version string
		status        ToolStatus
	}
	seen := make(map[resultKey]bool)
	unique := results[:0]
	for _, result := range results {
		path := result.resolvedPath
		if path == "" {
			path = result.Path
		}
		key := resultKey{path: path, version: result.Version, status: result.Status}
		if !seen[key] {
			unique = append(unique, result)
			seen[key] = true
		}
	}
	return unique
}
