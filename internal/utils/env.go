package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

// UpdateEnvFile 更新 env 文件
// mode = "add" 表示添加/更新变量
// mode = "remove" 表示删除变量
func UpdateEnvFile(envFile string, vars map[string]string, mode string) error {
	// 读取已有 env 文件
	envMap, err := godotenv.Read(envFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read env file: %w", err)
	}
	if envMap == nil {
		envMap = make(map[string]string)
	}

	splitFlag := string(os.PathListSeparator)

	switch mode {
	case "add":
		handleAddMode(envMap, vars, splitFlag)
	case "remove":
		handleRemoveMode(envMap, vars, splitFlag)
	default:
		return fmt.Errorf("invalid mode: %s (expected 'add' or 'remove')", mode)
	}

	// 末尾追加 $PATH，避免重复
	appendPathIfNeeded(envMap, splitFlag)

	// 写回文件
	if err := godotenv.Write(envMap, envFile); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	// 修正 godotenv 对 $PATH 的转义
	content, err := os.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("failed to read env file for fix: %w", err)
	}
	fixed := strings.ReplaceAll(string(content), `\$PATH`, `$PATH`)
	if err := os.WriteFile(envFile, []byte(fixed), 0o600); err != nil {
		return fmt.Errorf("failed to write fixed env file: %w", err)
	}

	return nil
}

// handleAddMode 处理添加模式
func handleAddMode(envMap, vars map[string]string, splitFlag string) {
	for k, v := range vars {
		if k == envPath {
			addPathVariable(envMap, v, splitFlag)
		} else {
			envMap[k] = v
		}
	}
}

// addPathVariable 添加PATH变量
func addPathVariable(envMap map[string]string, newPath, splitFlag string) {
	existing := envMap[envPath]
	if existing == "" {
		envMap[envPath] = newPath
		return
	}

	// 如果自定义路径已经在 PATH 中，就跳过
	found := false
	paths := strings.Split(existing, splitFlag)
	for _, p := range paths {
		if p == newPath {
			found = true
			break
		}
	}
	if !found {
		envMap[envPath] = newPath + splitFlag + existing
	}
}

// handleRemoveMode 处理删除模式
func handleRemoveMode(envMap, vars map[string]string, splitFlag string) {
	for k, v := range vars {
		if k == envPath {
			removePathVariable(envMap, v, splitFlag)
		} else {
			delete(envMap, k)
		}
	}
}

// removePathVariable 删除PATH变量中的指定路径
func removePathVariable(envMap map[string]string, pathToRemove, splitFlag string) {
	existing := envMap[envPath]
	if existing == "" {
		return
	}
	paths := strings.Split(existing, splitFlag)
	newPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		if p != pathToRemove && p != "$PATH" {
			newPaths = append(newPaths, p)
		}
	}
	// 如果删除后为空，不写 PATH
	if len(newPaths) > 0 {
		envMap[envPath] = strings.Join(newPaths, splitFlag)
	} else {
		delete(envMap, envPath)
	}
}

// appendPathIfNeeded 在PATH末尾追加$PATH
func appendPathIfNeeded(envMap map[string]string, splitFlag string) {
	if p, ok := envMap[envPath]; ok && !strings.HasSuffix(p, "$PATH") {
		if strings.HasSuffix(p, splitFlag) {
			envMap[envPath] = p + "$PATH"
		} else {
			envMap[envPath] = p + splitFlag + "$PATH"
		}
	}
}

// MergeJSON 返回合并后的 JSON 字符串，不写文件
func MergeJSON(orig map[string]interface{}, updates map[string]interface{}) (string, error) {
	// 如果原始 map 为 nil，创建一个空 map
	if orig == nil {
		orig = make(map[string]interface{})
	}

	// 合并更新字段
	for k, v := range updates {
		orig[k] = v
	}

	// 生成格式化 JSON 字符串
	data, err := json.MarshalIndent(orig, "", "    ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// EnsureBlockInFile 确保文件中存在指定 block（startMark ~ endMark之间的内容会被替换）
func EnsureBlockInFile(filePath, startMark, endMark, block string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(filePath, []byte(block+"\n"), 0o600)
		}
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	text := string(content)

	// 如果已包含 startMark 和 endMark，就替换成新的 block
	if strings.Contains(text, startMark) && strings.Contains(text, endMark) {
		re := regexp.MustCompile("(?s)" + regexp.QuoteMeta(startMark) + ".*?" + regexp.QuoteMeta(endMark))
		newText := re.ReplaceAllString(text, block)
		if newText != text {
			return os.WriteFile(filePath, []byte(newText), 0o600)
		}
		return nil
	}

	// 否则，直接追加
	newText := strings.TrimRight(text, "\n") + "\n" + block + "\n"
	return os.WriteFile(filePath, []byte(newText), 0o600)
}

// RemoveLinesInFile 删除文件中 startMark ~ endMark 的内容
func RemoveLinesInFile(filePath, startMark, endMark string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	text := string(content)
	re := regexp.MustCompile("(?s)" + regexp.QuoteMeta(startMark) + ".*?" + regexp.QuoteMeta(endMark))
	newText := re.ReplaceAllString(text, "")

	if newText == text {
		return nil
	}

	return os.WriteFile(filePath, []byte(strings.TrimRight(newText, "\n")+"\n"), 0o600)
}
