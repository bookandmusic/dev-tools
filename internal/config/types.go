package config

type AnsibleConfig struct {
	BaseDir    string `yaml:"base-dir"`
	PythonDir  string `yaml:"python-dir"`
	AnsibleDir string `yaml:"ansible-dir"`
}

type LangConfig struct {
	BaseDir  string   `yaml:"base-dir"`
	Versions []string `yaml:"versions"`
	Global   string   `yaml:"global"`
}

type GoConfig struct {
	BaseDir  string   `yaml:"base-dir"`
	Versions []string `yaml:"versions"`
	Global   string   `yaml:"global"`
}

type OhMyzshPlugin struct {
	Name string `yaml:"name"`
	Repo string `yaml:"repo"`
}

type OhMyzshConfig struct {
	InstallDir string           `yaml:"install-dir"`
	Theme      string           `yaml:"theme"`
	Plugins    []*OhMyzshPlugin `yaml:"plugins"`
}

type CommonConfig struct {
	Debug       bool   `yaml:"debug"`
	RootDir     string `yaml:"root-dir"`
	WorkDir     string `yaml:"work-dir"`
	CacheDir    string `yaml:"cache-dir"`
	GithubProxy string `yaml:"github-proxy"`
	HttpProxy   string `yaml:"http-proxy"`
}

type DockerConfig struct {
	InstallDir      string   `yaml:"install-dir"`
	Version         string   `yaml:"version" default:"26.1.0"`
	HttpProxy       string   `yaml:"http-proxy"`
	RegistryMirrors []string `yaml:"registry-mirrors"`
}

type GlobalConfig struct {
	Common  *CommonConfig  `yaml:"common"`
	Ansible *AnsibleConfig `yaml:"ansible"`
	Python  *LangConfig    `yaml:"python"`
	Go      *LangConfig    `yaml:"go"`
	OhMyzsh *OhMyzshConfig `yaml:"oh-my-zsh"`
	Docker  *DockerConfig  `yaml:"docker"`
}
