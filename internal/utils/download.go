package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	progressbar "github.com/schollz/progressbar/v3"

	"github.com/bookandmusic/dev-tools/internal/ui"
)

// DownloadFileWithProgress 下载指定 URL 的文件到 destPath，并通过传入的 UI 输出提示信息和进度。
func DownloadFileWithProgress(url, destPath string, console ui.UI) error {
	// 校验文件路径安全性
	cleanDestPath := filepath.Clean(destPath)

	// 确保目标文件目录存在
	if err := os.MkdirAll(filepath.Dir(cleanDestPath), os.ModePerm); err != nil {
		console.Error("Failed to create directory: %v", err)
		return err
	}

	// 发起 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	// 获取响应
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		console.Error("Request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		console.Error("Download failed, status code: %d", resp.StatusCode)
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 创建文件
	outFile, err := os.Create(cleanDestPath)
	if err != nil {
		console.Error("Failed to create file %s: %v", cleanDestPath, err)
		return err
	}
	defer outFile.Close()

	// 设置进度条
	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetDescription(fmt.Sprintf("Downloading %s", filepath.Base(cleanDestPath))),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// 下载文件并显示进度
	_, err = io.Copy(io.MultiWriter(outFile, bar), resp.Body)
	console.Println("")
	return err
}
