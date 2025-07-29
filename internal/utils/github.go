package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GetLatestReleaseTag 查询 GitHub 仓库的最新 release tag
// repoFullName 格式示例："docker/buildx"
// proxy 为可选代理前缀，例如 "https://ghproxy.com/"
// 如果不需要代理，传空字符串 "" 即可。
func GetLatestReleaseTag(repoFullName, proxy string) (string, error) {
	if repoFullName == "" || !strings.Contains(repoFullName, "/") {
		return "", fmt.Errorf("invalid repoFullName: %s, expected format 'owner/repo'", repoFullName)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repoFullName)
	if proxy != "" {
		// 注意：拼接代理 URL 需要确保格式正确
		if strings.HasSuffix(proxy, "/") {
			apiURL = proxy + apiURL
		} else {
			apiURL = proxy + "/" + apiURL
		}
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get latest release: HTTP %d", resp.StatusCode)
	}

	var body struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}

	if body.TagName == "" {
		return "", fmt.Errorf("tag_name not found in release info")
	}

	return body.TagName, nil
}
