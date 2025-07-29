package utils

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// RemoveBinaries attempts to remove the given list of binary names under the specified dir.
// It returns a combined error if any removal fails.
func RemoveBinaries(binDir string, names []string) error {
	var errList []string
	for _, name := range names {
		path := filepath.Join(binDir, name)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errList = append(errList, fmt.Sprintf("remove %s: %v", path, err))
		}
	}

	if len(errList) > 0 {
		return fmt.Errorf("errors removing binaries: %s", strings.Join(errList, "; "))
	}
	return nil
}

func ExtractTarGz(tarGzPath string, targetDir string) error {
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
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
			os.Chmod(destPath, 0o755)
		}
	}
	return nil
}

func CopyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcFile := filepath.Join(src, entry.Name())
		dstFile := filepath.Join(dst, entry.Name())

		data, err := os.ReadFile(srcFile)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dstFile, data, 0o755); err != nil {
			return err
		}
	}
	return nil
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
