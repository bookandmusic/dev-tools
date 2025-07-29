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
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		console.Error("Failed to create directory: %v", err)
		return err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		console.Error("Request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		console.Error("Download failed, status code: %d", resp.StatusCode)
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		console.Error("Failed to create file %s: %v", destPath, err)
		return err
	}
	defer outFile.Close()

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

	_, err = io.Copy(io.MultiWriter(outFile, bar), resp.Body)
	console.Println("")
	return err
}
