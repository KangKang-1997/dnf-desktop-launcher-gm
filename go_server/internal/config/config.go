package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

type Settings struct {
	ListenHost          string   `json:"listen_host"`
	ListenPort          int      `json:"listen_port"`
	ServerHost          string   `json:"server_host"`
	DBPort              int      `json:"db_port"`
	DBUser              string   `json:"db_user"`
	DBPassword          string   `json:"db_password"`
	GameDBName          string   `json:"game_db_name"`
	ToolDBName          string   `json:"tool_db_name"`
	DBCharset           string   `json:"db_charset"`
	SessionSecret       string   `json:"session_secret"`
	SessionTTLSeconds   int      `json:"session_ttl_seconds"`
	CORSOrigins         []string `json:"cors_origins"`
	LoginPrivateKeyPath string   `json:"login_private_key_path"`
}

func Default() Settings {
	return Settings{
		ListenHost:          "0.0.0.0",
		ListenPort:          8000,
		ServerHost:          "127.0.0.1",
		DBPort:              3306,
		DBUser:              "数据库用户名",
		DBPassword:          "数据库密码",
		GameDBName:          "d_taiwan",
		ToolDBName:          "dnf_launcher",
		DBCharset:           "utf8",
		SessionSecret:       "发布时改成随机长字符串",
		SessionTTLSeconds:   86400,
		CORSOrigins:         []string{"*"},
		LoginPrivateKeyPath: "/home/neople/game/privatekey.pem",
	}
}

func Load(path string) (Settings, error) {
	settings := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return settings, nil
		}
		return Settings{}, err
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}, err
	}
	if settings.ListenHost == "" {
		settings.ListenHost = "0.0.0.0"
	}
	if settings.ListenPort <= 0 {
		settings.ListenPort = 8000
	}
	if len(settings.CORSOrigins) == 0 {
		settings.CORSOrigins = []string{"*"}
	}
	return settings, nil
}

func (s Settings) ListenAddress() string {
	return s.ListenHost + ":" + strconv.Itoa(s.ListenPort)
}
