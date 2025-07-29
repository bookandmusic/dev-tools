package docker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bookandmusic/dev-tools/internal/utils"
)

func (i *installer) installDocker() error {
	i.ui.Info("Installing Docker %s ...", i.cfg.DockerVersion)

	tmpDir := "./tmp"
	defer os.RemoveAll(tmpDir)

	tarFile := fmt.Sprintf("docker-%s.tgz", i.cfg.DockerVersion)
	url := fmt.Sprintf("https://%s/docker-ce/linux/static/stable/%s/%s", i.cfg.Domain, i.cfg.Arch, tarFile)
	tarFilePath := filepath.Join(tmpDir, tarFile)

	// 1. 下载 Docker 包
	i.ui.Info("Downloading Docker from %s ...", url)
	if err := utils.DownloadFileWithProgress(url, tarFilePath, i.ui); err != nil {
		i.ui.Error("Failed to download Docker: %v", err)
		return err
	}
	i.ui.Success("Docker downloaded to %s", tarFilePath)

	// 2. 解压
	i.ui.Info("Extracting Docker archive...")
	if err := utils.ExtractTarGz(tarFilePath, tmpDir); err != nil {
		i.ui.Error("Failed to extract Docker archive: %v", err)
		return err
	}
	i.ui.Success("Docker archive extracted successfully.")

	// 3. 安装二进制文件
	i.ui.Info("Installing Docker binaries to %s...", i.cfg.BinDir)
	if err := utils.CreateDirIfNotExists(i.cfg.BinDir); err != nil {
		i.ui.Error("Failed to create directory %s: %v", i.cfg.BinDir, err)
		return err
	}

	dockerTmpDir := filepath.Join(tmpDir, "docker")

	// 4. 增加执行权限
	i.ui.Info("Adding execute permission to Docker binaries...")
	entries, err := os.ReadDir(dockerTmpDir)
	if err != nil {
		i.ui.Error("Failed to read directory %s: %v", dockerTmpDir, err)
		return err
	}
	for _, entry := range entries {
		if err := utils.AddExecPermission(filepath.Join(dockerTmpDir, entry.Name())); err != nil {
			i.ui.Error("Failed to add exec permission to %s: %v", entry.Name(), err)
			return err
		}
	}

	// 5. 拷贝到安装目录
	i.ui.Info("Copying Docker binaries to %s...", i.cfg.BinDir)
	if err := utils.CopyDir(dockerTmpDir, i.cfg.BinDir); err != nil {
		i.ui.Error("Failed to copy Docker binaries: %v", err)
		return err
	}

	// 6. 添加 PATH
	i.ui.Info("Adding %s to user PATH...", i.cfg.BinDir)
	if err := utils.AddPathToUserEnv(i.cfg.BinDir); err != nil {
		i.ui.Error("Failed to add path to user environment: %v", err)
		return err
	}
	return nil
}

func (i *installer) ensureDaemonConfig() error {
	configPath := "/etc/docker/daemon.json"

	i.ui.Info("Ensuring Docker daemon config at %s...", configPath)
	err := utils.UpdateJSONConfigFile(configPath, func(config map[string]interface{}) error {
		config["bip"] = "172.18.0.1/16"

		if val, ok := config["insecure-registries"]; !ok || val == nil {
			config["insecure-registries"] = []interface{}{}
		}

		existingSet := make(map[string]struct{})
		if arr, ok := config["registry-mirrors"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					existingSet[s] = struct{}{}
				}
			}
		}
		merged := make([]string, 0, len(existingSet))
		for m := range existingSet {
			merged = append(merged, m)
		}
		config["registry-mirrors"] = merged
		return nil
	})
	if err != nil {
		// 非致命问题，记录警告
		i.ui.Warning("Failed to update daemon config: %v", err)
		return err
	}

	i.ui.Success("Docker daemon configuration updated successfully.")
	return nil
}
