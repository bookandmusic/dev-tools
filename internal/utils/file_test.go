package utils_test

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bookandmusic/dev-tools/internal/ui"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

func PrepareTestTarGz(t *testing.T, baseDir, tarGzName string, files map[string]string) string {
	t.Helper()

	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatalf("failed to create base dir: %v", err)
	}

	tarGzPath := filepath.Join(baseDir, tarGzName)
	f, err := os.Create(tarGzPath)
	if err != nil {
		t.Fatalf("failed to create tar.gz file: %v", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if strings.HasSuffix(name, "/") {
			hdr.Typeflag = tar.TypeDir
		} else {
			hdr.Typeflag = tar.TypeReg
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write header: %v", err)
		}
		if hdr.Typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte(content)); err != nil {
				t.Fatalf("failed to write file content: %v", err)
			}
		}
	}

	return tarGzPath
}

func createTestFiles(t *testing.T, baseDir string, files map[string]string) {
	t.Helper()
	for path, content := range files {
		fullPath := filepath.Join(baseDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("failed to create dir for %s: %v", fullPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write file %s: %v", fullPath, err)
		}
	}
}

func TestExtractTarGzWithFileLogs(t *testing.T) {
	baseDir := t.TempDir()
	extractDir := filepath.Join(baseDir, "extracted")
	tarGzName := "test1.tar.gz"
	files := map[string]string{
		"file1.txt":    "Hello World",
		"folder/file2": "Test content",
	}

	tarGzPath := PrepareTestTarGz(t, baseDir, tarGzName, files)
	console := ui.ConsoleUI{}

	err := utils.ExtractTarGzWithFileLogs(tarGzPath, extractDir, console)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	checkFiles := []string{
		"file1.txt",
		"folder/file2",
	}

	for _, f := range checkFiles {
		path := filepath.Join(extractDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s not found", path)
		}
	}
}

func TestExtractTarGzWithProgress(t *testing.T) {
	baseDir := t.TempDir()
	extractDir := filepath.Join(baseDir, "extracted")
	tarGzName := "test1.tar.gz"
	files := map[string]string{
		"file1.txt":    "Hello World",
		"folder/file2": "Test content",
	}

	tarGzPath := PrepareTestTarGz(t, baseDir, tarGzName, files)
	console := ui.ConsoleUI{}

	err := utils.ExtractTarGzWithProgress(tarGzPath, extractDir, console)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	checkFiles := []string{
		"file1.txt",
		"folder/file2",
	}

	for _, f := range checkFiles {
		path := filepath.Join(extractDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s not found", path)
		}
	}
}

func TestRemoveFilesWithProgress(t *testing.T) {
	baseDir := t.TempDir()
	binDir := filepath.Join(baseDir, "bin")

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}

	fileNames := []string{"file1.exe", "file2.exe", "file3.exe"}

	for _, f := range fileNames {
		path := filepath.Join(binDir, f)
		if err := os.WriteFile(path, []byte("test content"), 0o644); err != nil {
			t.Fatalf("failed to create test file %s: %v", f, err)
		}
	}

	console := ui.ConsoleUI{}
	err := utils.RemoveFilesWithProgress(binDir, fileNames, console)
	if err != nil {
		t.Errorf("RemoveFilesWithProgress returned error: %v", err)
	}

	for _, f := range fileNames {
		path := filepath.Join(binDir, f)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("file %s should have been removed but still exists", f)
		}
	}
}

func TestCopyDirWithProgress(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	testFiles := map[string]string{
		"file1.txt":         "hello world",
		"folder1/file2.md":  "# markdown",
		"folder2/file3.log": "log content",
	}
	createTestFiles(t, srcDir, testFiles)

	console := ui.ConsoleUI{}
	err := utils.CopyDirWithProgress(srcDir, dstDir, console)
	if err != nil {
		t.Fatalf("CopyDirWithProgress failed: %v", err)
	}

	for path, content := range testFiles {
		dstPath := filepath.Join(dstDir, path)
		data, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("expected file %s missing: %v", dstPath, err)
			continue
		}
		if string(data) != content {
			t.Errorf("file %s content mismatch: got %q want %q", dstPath, string(data), content)
		}
	}

	err = filepath.Walk(srcDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(srcDir, path)
		dstPath := filepath.Join(dstDir, rel)
		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			t.Errorf("expected %s exists, but got error: %v", dstPath, err)
			return nil
		}
		if info.IsDir() != dstInfo.IsDir() {
			t.Errorf("type mismatch for %s and %s", path, dstPath)
		}
		return nil
	})
	if err != nil {
		t.Errorf("filepath.Walk error: %v", err)
	}
}

func TestCreateDirIfNotExists(t *testing.T) {
	baseDir := t.TempDir()
	dirPath := filepath.Join(baseDir, "subdir")

	if err := utils.CreateDirIfNotExists(dirPath); err != nil {
		t.Fatalf("CreateDirIfNotExists failed: %v", err)
	}
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("expected dir exists, but got error: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected a directory but found not a dir")
	}
	t.Logf("Directory created successfully: %s", dirPath)

	if err := utils.CreateDirIfNotExists(dirPath); err != nil {
		t.Fatalf("CreateDirIfNotExists returned error on existing dir: %v", err)
	}
	t.Logf("Directory exists check passed: %s", dirPath)

	filePath := filepath.Join(baseDir, "not_a_dir")
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	err = utils.CreateDirIfNotExists(filePath)
	if err == nil {
		t.Fatalf("expected error when path is a file, but got nil")
	}
	t.Logf("Got expected error when path exists as a file: %v", err)
}

func TestAddExecPermission(t *testing.T) {
	baseDir := t.TempDir()

	t.Logf("Using temp test directory: %s", baseDir)

	// 1. 普通文件初始无执行权限，加执行权限成功
	filePath := filepath.Join(baseDir, "testfile.sh")
	if err := os.WriteFile(filePath, []byte("#!/bin/bash\necho hello"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	t.Logf("Created test file with 0644 permission: %s", filePath)

	if err := utils.AddExecPermission(filePath); err != nil {
		t.Fatalf("AddExecPermission failed: %v", err)
	}
	t.Logf("AddExecPermission succeeded on file: %s", filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat file failed: %v", err)
	}
	mode := info.Mode()
	t.Logf("File mode after AddExecPermission: %v", mode)
	if mode&0o111 == 0 {
		t.Errorf("expected execute permission bits to be set, but got mode: %v", mode)
	} else {
		t.Logf("Execute permission bits correctly set")
	}

	// 2. 已经有执行权限的文件，再次加执行权限应无错误且权限保持
	if err := utils.AddExecPermission(filePath); err != nil {
		t.Errorf("AddExecPermission on executable file returned error: %v", err)
	} else {
		t.Logf("AddExecPermission on already executable file succeeded")
	}
	info2, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat file failed: %v", err)
	}
	if info2.Mode() != mode {
		t.Errorf("expected mode unchanged, got %v vs %v", info2.Mode(), mode)
	} else {
		t.Logf("File mode unchanged on repeated AddExecPermission")
	}

	// 3. 对目录调用，预期可能成功或失败，均输出日志
	dirPath := filepath.Join(baseDir, "subdir")
	if err := os.Mkdir(dirPath, 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	t.Logf("Created directory: %s", dirPath)

	err = utils.AddExecPermission(dirPath)
	if err != nil {
		t.Logf("AddExecPermission on directory returned error (acceptable): %v", err)
	} else {
		t.Log("AddExecPermission on directory succeeded (acceptable)")
	}

	// 4. 不存在路径应返回错误
	nonExistPath := filepath.Join(baseDir, "nonexist.sh")
	err = utils.AddExecPermission(nonExistPath)
	if err == nil {
		t.Errorf("expected error for non-existent file, got nil")
	} else {
		t.Logf("Got expected error for non-existent file: %v", err)
	}

	// 5. 只读文件尝试加执行权限，输出结果
	readonlyPath := filepath.Join(baseDir, "readonly.sh")
	if err := os.WriteFile(readonlyPath, []byte("readonly content"), 0o444); err != nil {
		t.Fatalf("failed to create readonly file: %v", err)
	}
	t.Logf("Created readonly file with 0444 permission: %s", readonlyPath)

	err = utils.AddExecPermission(readonlyPath)
	if err != nil {
		t.Logf("AddExecPermission on readonly file returned error (acceptable): %v", err)
	} else {
		t.Log("AddExecPermission on readonly file succeeded (acceptable)")
		infoR, err := os.Stat(readonlyPath)
		if err != nil {
			t.Fatalf("stat readonly file failed: %v", err)
		}
		t.Logf("Readonly file mode after AddExecPermission: %v", infoR.Mode())
		if infoR.Mode()&0o111 == 0 {
			t.Errorf("expected execute permission bits set on readonly file, but got mode: %v", infoR.Mode())
		} else {
			t.Logf("Execute permission bits correctly set on readonly file")
		}
	}
}

func TestUpdateJSONConfigFile(t *testing.T) {
	baseDir := t.TempDir()
	configPath := filepath.Join(baseDir, "config.json")

	updater1 := func(cfg map[string]interface{}) error {
		cfg["key1"] = "value1"
		return nil
	}
	if err := utils.UpdateJSONConfigFile(configPath, updater1); err != nil {
		t.Fatalf("UpdateJSONConfigFile failed on non-existent file: %v", err)
	}
	t.Log("Test1 passed: Successfully created new config file with key1=value1")

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config after write: %v", err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse written config: %v", err)
	}
	if val, ok := cfg["key1"]; !ok || val != "value1" {
		t.Errorf("Test1 verification failed: expected key1=value1, got %v", cfg)
	} else {
		t.Log("Test1 verification passed: key1=value1 present in config")
	}

	updater2 := func(cfg map[string]interface{}) error {
		cfg["key2"] = 123
		return nil
	}
	if err := utils.UpdateJSONConfigFile(configPath, updater2); err != nil {
		t.Fatalf("UpdateJSONConfigFile failed on existing file: %v", err)
	}
	t.Log("Test2 passed: Successfully updated existing config file with key2=123")

	data, _ = os.ReadFile(configPath)
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse updated config: %v", err)
	}
	if val, ok := cfg["key2"]; !ok || val.(float64) != 123 {
		t.Errorf("Test2 verification failed: expected key2=123, got %v", cfg)
	} else {
		t.Log("Test2 verification passed: key2=123 present in config")
	}

	updaterErr := func(cfg map[string]interface{}) error {
		return os.ErrInvalid
	}
	err = utils.UpdateJSONConfigFile(configPath, updaterErr)
	if err == nil {
		t.Errorf("Test3 failed: expected error from updater, got nil")
	} else {
		t.Logf("Test3 passed: got expected updater error: %v", err)
	}
}
