package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr       string
	DataDir          string
	DBPath           string
	StaticDir        string
	ConfigPath       string
	SeedCIDRs        []string
	DefaultScanPorts []int
	SeedDockerSocket bool
	AutoBootstrap    bool
	AdminToken       string
	AdminTokenFile   string
}

type fileConfig struct {
	Server struct {
		ListenAddr string `yaml:"listenAddr"`
	} `yaml:"server"`
	Storage struct {
		DataDir string `yaml:"dataDir"`
		DBPath  string `yaml:"dbPath"`
	} `yaml:"storage"`
	Frontend struct {
		StaticDir string `yaml:"staticDir"`
	} `yaml:"frontend"`
	Discovery struct {
		SeedCIDRs        []string `yaml:"seedCidrs"`
		DefaultScanPorts []int    `yaml:"defaultScanPorts"`
		SeedDockerSocket *bool    `yaml:"seedDockerSocket"`
	} `yaml:"discovery"`
	Bootstrap struct {
		AutoBootstrap  *bool  `yaml:"autoBootstrap"`
		AdminTokenFile string `yaml:"adminTokenFile"`
	} `yaml:"bootstrap"`
}

func Load() (Config, error) {
	adminTokenFileExplicit := false
	cfg := Config{
		ListenAddr:       ":8080",
		DataDir:          "./data",
		StaticDir:        "./web/dist",
		SeedDockerSocket: true,
		DefaultScanPorts: []int{22, 80, 443, 8080, 8443},
		AutoBootstrap:    true,
	}
	cfg.DBPath = filepath.Join(cfg.DataDir, "homelabwatch.db")
	cfg.AdminTokenFile = filepath.Join(cfg.DataDir, "admin-token")
	cfg.ConfigPath = envString("HOMELABWATCH_CONFIG", "")

	if cfg.ConfigPath != "" {
		explicit, err := loadYAML(&cfg, cfg.ConfigPath)
		if err != nil {
			return Config{}, err
		}
		adminTokenFileExplicit = adminTokenFileExplicit || explicit
	} else {
		for _, candidate := range []string{"./config.yaml", "./config.yml"} {
			if _, err := os.Stat(candidate); err == nil {
				explicit, err := loadYAML(&cfg, candidate)
				if err != nil {
					return Config{}, err
				}
				adminTokenFileExplicit = adminTokenFileExplicit || explicit
				cfg.ConfigPath = candidate
				break
			}
		}
	}

	cfg.ListenAddr = envString("HOMELABWATCH_LISTEN_ADDR", cfg.ListenAddr)
	cfg.DataDir = envString("HOMELABWATCH_DATA_DIR", cfg.DataDir)
	cfg.DBPath = envString("HOMELABWATCH_DB_PATH", cfg.DBPath)
	cfg.StaticDir = envString("HOMELABWATCH_STATIC_DIR", cfg.StaticDir)
	if cidrs := envCSV("HOMELABWATCH_SEED_CIDRS"); len(cidrs) > 0 {
		cfg.SeedCIDRs = cidrs
	}
	if ports := envCSVInts("HOMELABWATCH_DEFAULT_SCAN_PORTS"); len(ports) > 0 {
		cfg.DefaultScanPorts = ports
	}
	cfg.SeedDockerSocket = envBool("HOMELABWATCH_SEED_DOCKER_SOCKET", cfg.SeedDockerSocket)
	cfg.AutoBootstrap = envBool("HOMELABWATCH_AUTO_BOOTSTRAP", cfg.AutoBootstrap)
	cfg.AdminToken = envString("HOMELABWATCH_ADMIN_TOKEN", cfg.AdminToken)
	if tokenFile := envString("HOMELABWATCH_ADMIN_TOKEN_FILE", ""); tokenFile != "" {
		cfg.AdminTokenFile = tokenFile
		adminTokenFileExplicit = true
	}

	if strings.TrimSpace(cfg.DBPath) == "" {
		cfg.DBPath = filepath.Join(cfg.DataDir, "homelabwatch.db")
	}
	if strings.TrimSpace(cfg.DataDir) == "" {
		cfg.DataDir = filepath.Dir(cfg.DBPath)
	}
	if !adminTokenFileExplicit || strings.TrimSpace(cfg.AdminTokenFile) == "" {
		cfg.AdminTokenFile = filepath.Join(cfg.DataDir, "admin-token")
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return Config{}, err
	}
	if len(cfg.DefaultScanPorts) == 0 {
		return Config{}, errors.New("default scan ports cannot be empty")
	}
	return cfg, nil
}

func loadYAML(cfg *Config, path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	var fileCfg fileConfig
	if err := yaml.Unmarshal(content, &fileCfg); err != nil {
		return false, err
	}
	adminTokenFileExplicit := false
	if fileCfg.Server.ListenAddr != "" {
		cfg.ListenAddr = fileCfg.Server.ListenAddr
	}
	if fileCfg.Storage.DataDir != "" {
		cfg.DataDir = fileCfg.Storage.DataDir
	}
	if fileCfg.Storage.DBPath != "" {
		cfg.DBPath = fileCfg.Storage.DBPath
	}
	if fileCfg.Frontend.StaticDir != "" {
		cfg.StaticDir = fileCfg.Frontend.StaticDir
	}
	if len(fileCfg.Discovery.SeedCIDRs) > 0 {
		cfg.SeedCIDRs = append([]string(nil), fileCfg.Discovery.SeedCIDRs...)
	}
	if len(fileCfg.Discovery.DefaultScanPorts) > 0 {
		cfg.DefaultScanPorts = append([]int(nil), fileCfg.Discovery.DefaultScanPorts...)
	}
	if fileCfg.Discovery.SeedDockerSocket != nil {
		cfg.SeedDockerSocket = *fileCfg.Discovery.SeedDockerSocket
	}
	if fileCfg.Bootstrap.AutoBootstrap != nil {
		cfg.AutoBootstrap = *fileCfg.Bootstrap.AutoBootstrap
	}
	if fileCfg.Bootstrap.AdminTokenFile != "" {
		cfg.AdminTokenFile = fileCfg.Bootstrap.AdminTokenFile
		adminTokenFileExplicit = true
	}
	if strings.TrimSpace(cfg.DBPath) == "" {
		cfg.DBPath = filepath.Join(cfg.DataDir, "homelabwatch.db")
	}
	return adminTokenFileExplicit, nil
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envCSV(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func envCSVInts(key string) []int {
	parts := envCSV(key)
	values := make([]int, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value <= 0 {
			continue
		}
		values = append(values, value)
	}
	return values
}

func envBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}
