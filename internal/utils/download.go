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
	console.Info("Downloading from %s", url)

	// 创建目标目录（如果不存在）
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		console.Error("Failed to create directory: %v", err)
		return err
	}

	// 发起请求
	resp, err := http.Get(url)
	if err != nil {
		console.Error("Request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		console.Error("Download failed, status code: %d", resp.StatusCode)
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 创建目标文件
	outFile, err := os.Create(destPath)
	if err != nil {
		console.Error("Failed to create file %s: %v", destPath, err)
		return err
	}
	defer outFile.Close()

	// 进度条
	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetDescription(fmt.Sprintf("Downloading %s", filepath.Base(destPath))),
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

	// 写入文件并更新进度
	_, err = io.Copy(io.MultiWriter(outFile, bar), resp.Body)
	if err != nil {
		console.Error("Failed to write file: %v", err)
		return err
	}

	console.Success("Download completed: %s", destPath)
	return nil
}
