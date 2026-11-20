package devtools

import "context"

// DetectUV checks the uv executable found in the app's PATH.
func (Service) DetectUV(ctx context.Context) (ToolDetection, error) {
	if err := ctx.Err(); err != nil {
		return ToolDetection{}, err
	}
	path, err := findInPath("uv")
	if err != nil {
		return pathFailure(path, err), nil
	}
	if path == "" {
		return ToolDetection{
			Status:  statusNotFound,
			Message: "应用的 PATH 中没有找到 uv。",
		}, nil
	}
	return detectVersion(ctx, path, "uv ")
}
