package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	progressbar "github.com/schollz/progressbar/v3"

	"github.com/bookandmusic/dev-tools/internal/ui"
)

// ExtractTarGzWithProgress 解压 tar.gz 包，支持 strip 参数
func ExtractTarGzWithProgress(console ui.UI, tarGzPath, targetDir string, strip int) error {
	// 清理路径
	cleanTarGzPath := filepath.Clean(tarGzPath)
	cleanTargetDir := filepath.Clean(targetDir)

	// 打开文件
	f, err := os.Open(cleanTarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	compressedSize := stat.Size()

	// 扫描进度条
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

	// TeeReader 更新扫描进度
	scanReader := io.TeeReader(f, scanBar)

	// gzip 解压
	gzr, err := gzip.NewReader(scanReader)
	if err != nil {
		fmt.Println("")
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// 先扫描统计解压大小
	totalSize, err := scanArchive(tr, strip)
	if err != nil {
		fmt.Println("")
		return err
	}
	fmt.Println("") // 扫描结束换行

	// 重新打开文件，开始解压
	f2, err := os.Open(cleanTarGzPath)
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

	// 解压进度条
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
	return extractFiles(tr2, buf, cleanTargetDir, strip, extractBar)
}

// scanArchive 扫描归档文件并计算总大小
func scanArchive(tr *tar.Reader, strip int) (int64, error) {
	var totalSize int64
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		if hdr.Typeflag == tar.TypeReg {
			relPath := stripPathComponents(hdr.Name, strip)
			if relPath != "" {
				totalSize += hdr.Size
			}
		}
	}
	return totalSize, nil
}

// extractFiles 提取文件
func extractFiles(tr *tar.Reader, buf []byte, targetDir string, strip int, bar *progressbar.ProgressBar) error {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// strip 处理
		relPath := stripPathComponents(hdr.Name, strip)
		if relPath == "" {
			continue
		}

		destPath := filepath.Join(targetDir, relPath)
		cleanDestPath := filepath.Clean(destPath)

		// 判断是否目录
		isDir := hdr.Typeflag == tar.TypeDir || strings.HasSuffix(hdr.Name, "/")

		if isDir {
			if err := os.MkdirAll(cleanDestPath, 0o700); err != nil {
				return err
			}
			continue
		}

		// 普通文件
		if err := extractRegularFile(tr, buf, cleanDestPath, bar); err != nil {
			fmt.Println("")
			return err
		}
	}

	fmt.Println("") // 解压完成换行
	return nil
}

// extractRegularFile 提取普通文件
func extractRegularFile(tr *tar.Reader, buf []byte, destPath string, bar *progressbar.ProgressBar) error {
	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(destPath), 0o700); err != nil {
		return err
	}

	// 如果已经是目录，先删掉
	if fi, err := os.Stat(destPath); err == nil && fi.IsDir() {
		if rmErr := os.RemoveAll(destPath); rmErr != nil {
			return rmErr
		}
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 读取并写入文件内容
	for {
		n, err := tr.Read(buf)
		if n > 0 {
			if _, wErr := out.Write(buf[:n]); wErr != nil {
				return wErr
			}
			if barErr := bar.Add(n); barErr != nil {
				return barErr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// 设置文件权限
	if err := os.Chmod(destPath, 0o700); err != nil {
		return err
	}

	return nil
}

// stripPathComponents 去掉路径前 n 层
func stripPathComponents(path string, strip int) string {
	parts := strings.Split(path, "/")
	if len(parts) <= strip {
		return ""
	}
	return filepath.Join(parts[strip:]...)
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

		if err := bar.Add(1); err != nil {
			fmt.Println("")
			return err
		} // 每删除一个文件，进度条前进一格
	}

	fmt.Println("") // 换行让输出更干净

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
		if err := os.MkdirAll(filepath.Dir(dstFile), 0o700); err != nil {
			return err
		}

		// 复制文件
		if err := CopyFile(srcFile, dstFile); err != nil {
			return err
		}

		if err := bar.Add(1); err != nil {
			return err
		}
	}

	fmt.Println("") // 换行
	return nil
}

// copyFile 按块复制文件并保留权限
func CopyFile(srcFile, dstFile string) error {
	// 清理路径防止路径遍历攻击
	cleanSrcFile := filepath.Clean(srcFile)
	cleanDstFile := filepath.Clean(dstFile)
	src, err := os.Open(cleanSrcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.OpenFile(cleanDstFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	buf := make([]byte, 32*1024) // 32KB buffer
	_, err = io.CopyBuffer(dst, src, buf)
	return err
}
