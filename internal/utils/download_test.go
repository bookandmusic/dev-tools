package utils_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/bookandmusic/dev-tools/internal/ui"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

func TestDownloadFileWithProgress(t *testing.T) {
	// 准备模拟 HTTP 服务器，返回简单文本数据
	const content = "Hello, this is test file content."
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		io.WriteString(w, content)
	}))
	defer server.Close()

	// 临时目录做下载目标
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "downloaded_test_file.txt")

	console := ui.ConsoleUI{}

	err := utils.DownloadFileWithProgress(server.URL, destPath, console)
	if err != nil {
		t.Fatalf("DownloadFileWithProgress failed: %v", err)
	}

	// 校验文件内容正确
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}

	if string(data) != content {
		t.Errorf("downloaded file content mismatch:\nexpected: %q\ngot: %q", content, string(data))
	}
}
