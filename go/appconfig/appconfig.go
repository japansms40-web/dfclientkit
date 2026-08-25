// Package appconfig 提供"读 JSON 配置、按系统目录存、缺省兜底"这套通用持久化机制，
// 供各客户端项目的运行参数结构体复用。
package appconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultDir 返回 appName 对应应用在当前系统上的默认状态目录
// （macOS: ~/Library/Application Support/<appName>，Windows: %AppData%/<appName>）。
func DefaultDir(appName string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName), nil
}

// LoadFile 从 path 指向的 JSON 文件读取配置；文件不存在或内容无法解析时返回 defaults。
func LoadFile[T any](path string, defaults T) T {
	b, err := os.ReadFile(path)
	if err != nil {
		return defaults
	}
	cfg := defaults
	if err := json.Unmarshal(b, &cfg); err != nil {
		return defaults
	}
	return cfg
}

// SaveFile 把 cfg 写入 path 指向的 JSON 文件，按需创建父目录。
func SaveFile[T any](path string, cfg T) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// Load 从 appName 对应默认目录下的 config.json 读取配置；目录/文件不存在或内容
// 无法解析时返回 defaults。
func Load[T any](appName string, defaults T) T {
	dir, err := DefaultDir(appName)
	if err != nil {
		return defaults
	}
	return LoadFile(filepath.Join(dir, "config.json"), defaults)
}

// Save 把 cfg 写入 appName 对应默认目录下的 config.json，按需创建父目录。
func Save[T any](appName string, cfg T) error {
	dir, err := DefaultDir(appName)
	if err != nil {
		return err
	}
	return SaveFile(filepath.Join(dir, "config.json"), cfg)
}
