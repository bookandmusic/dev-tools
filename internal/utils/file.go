package utils

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bookandmusic/dev-tools/internal/ui"
	progressbar "github.com/schollz/progressbar/v3"
)

func ExtractTarGzWithProgress(tarGzPath, targetDir string, console ui.UI) error {
	// 打开文件和获取压缩包大小
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	compressedSize := stat.Size()

	// 扫描进度条（读取压缩包的字节进度）
	scanBar := progressbar.NewOptions64(
		compressedSize,
		progressbar.OptionSetDescription("Scanning archive"),
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

	// 用 io.TeeReader 包装，扫描时实时更新扫描进度条
	scanReader := io.TeeReader(f, scanBar)

	// gzip 解压
	gzr, err := gzip.NewReader(scanReader)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// 扫描阶段，统计解压文件总大小
	var totalSize int64
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeReg {
			totalSize += hdr.Size
		}
	}
	console.Println("") // 扫描结束换行

	// --- 重新打开文件开始解压 ---
	f2, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f2.Close()

	gzr2, err := gzip.NewReader(f2)
	if err != nil {
		return err
	}
	defer gzr2.Close()

	tr2 := tar.NewReader(gzr2)

	// 解压进度条（基于实际写入字节）
	extractBar := progressbar.NewOptions64(
		totalSize,
		progressbar.OptionSetDescription("Extracting files"),
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

	buf := make([]byte, 32*1024)
	for {
		hdr, err := tr2.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		destPath := filepath.Join(targetDir, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
				return err
			}
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}

			for {
				n, err := tr2.Read(buf)
				if n > 0 {
					if _, wErr := out.Write(buf[:n]); wErr != nil {
						out.Close()
						return wErr
					}
					extractBar.Add(n)
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					out.Close()
					return err
				}
			}

			out.Close()
			os.Chmod(destPath, 0o755)
		}
	}

	console.Println("") // 解压完成换行
	return nil
}

// ExtractTarGzWithFileLogs 单次遍历解压，解压每个文件时输出日志，结束时打印总数
func ExtractTarGzWithFileLogs(tarGzPath, targetDir string, console ui.UI) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	fileCount := 0

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		destPath := filepath.Join(targetDir, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			// 创建目录
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
				return err
			}
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}

			// 复制文件内容
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()

			// 设置权限
			if err := os.Chmod(destPath, 0o755); err != nil {
				return err
			}

			fileCount++
			console.Println(fmt.Sprintf("Extracted: %s", hdr.Name))
		}
	}

	console.Println(fmt.Sprintf("Done extracting %d files.", fileCount))
	return nil
}

// RemoveBinaries attempts to remove the given list of binary names under the specified dir.
// It returns a combined error if any removal fails.
func RemoveFilesWithProgress(binDir string, names []string, console ui.UI) error {
	total := len(names)
	if total == 0 {
		return nil
	}

	bar := progressbar.NewOptions(
		total,
		progressbar.OptionSetDescription("Removing binaries"),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	var errList []string

	for _, name := range names {
		path := filepath.Join(binDir, name)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errList = append(errList, fmt.Sprintf("remove %s: %v", path, err))
		}

		bar.Add(1) // 每删除一个文件，进度条前进一格
	}

	console.Println("") // 换行让输出更干净

	if len(errList) > 0 {
		return fmt.Errorf("errors removing binaries: %s", errList)
	}

	return nil
}

// CopyDirWithProgress 递归复制目录并显示文件数量进度条
func CopyDirWithProgress(src, dst string, console ui.UI) error {
	// 先统计文件数量
	var files []string
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	bar := progressbar.NewOptions(
		len(files),
		progressbar.OptionSetDescription(fmt.Sprintf("Copying %s", filepath.Base(src))),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	for _, srcFile := range files {
		relPath, _ := filepath.Rel(src, srcFile)
		dstFile := filepath.Join(dst, relPath)

		// 创建目标目录
		if err := os.MkdirAll(filepath.Dir(dstFile), 0o755); err != nil {
			return err
		}

		// 复制文件
		if err := copyFile(srcFile, dstFile); err != nil {
			return err
		}

		bar.Add(1)
	}

	console.Println("") // 换行
	return nil
}

// copyFile 按块复制文件并保留权限
func copyFile(srcFile, dstFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.OpenFile(dstFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	buf := make([]byte, 32*1024) // 32KB buffer
	_, err = io.CopyBuffer(dst, src, buf)
	return err
}

func CreateDirIfNotExists(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return nil // already exists
		}
		return &os.PathError{Op: "mkdir", Path: path, Err: os.ErrExist}
	}

	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0o755)
	}

	return err
}

func AddExecPermission(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	mode := info.Mode()
	// 添加执行权限：对于用户、组、其他（u+x, g+x, o+x）
	newMode := mode | 0o111

	if err := os.Chmod(path, newMode); err != nil {
		return fmt.Errorf("chmod file: %w", err)
	}
	return nil
}

func UpdateJSONConfigFile(path string, updater func(map[string]interface{}) error) error {
	config := make(map[string]interface{})

	// 读取文件
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read config file %s: %w", path, err)
		}
		// 文件不存在，使用空配置
	} else if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parse config file %s: %w", path, err)
	}

	// 调用传入的 updater 函数修改配置
	if err := updater(config); err != nil {
		return fmt.Errorf("update config file %s: %w", path, err)
	}

	// 创建目录（防止写文件失败）
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir config dir %s: %w", filepath.Dir(path), err)
	}

	// 写回文件，格式化 JSON
	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config file %s: %w", path, err)
	}
	if err := os.WriteFile(path, newData, 0o644); err != nil {
		return fmt.Errorf("write config file %s: %w", path, err)
	}

	return nil
}
