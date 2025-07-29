package docker

import (
	"github.com/bookandmusic/dev-tools/internal/utils"
)

func (i *installer) detectVersions() error {
	// 内部工具函数：检测单个组件版本
	detect := func(name, repo string, target *string, trimV bool) error {
		// 如果版本已指定，跳过检测，但输出提示
		if *target != "" {
			i.ui.Info("%s version is already specified: %s, skip detection.", name, *target)
			return nil
		}

		i.ui.Info("Detecting latest %s version...", name)
		tag, err := utils.GetLatestReleaseTag(repo, i.cfg.Proxy)
		if err != nil {
			return err
		}

		// 去掉 v 前缀（如 v27.0.0 → 27.0.0）
		if trimV && len(tag) > 0 && tag[0] == 'v' {
			*target = tag[1:]
		} else {
			*target = tag
		}

		i.ui.Success("Detected %s version: %s", name, *target)
		return nil
	}

	// 顺序检测 Docker / Buildx / Compose
	if err := detect("Docker", "moby/moby", &i.cfg.DockerVersion, true); err != nil {
		return err
	}
	if err := detect("Buildx", "docker/buildx", &i.cfg.BuildxVersion, false); err != nil {
		return err
	}
	return detect("Compose", "docker/compose", &i.cfg.ComposeVersion, false)
}
