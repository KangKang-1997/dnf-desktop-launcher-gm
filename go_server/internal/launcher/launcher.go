package launcher

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"dnf-launcher-go/internal/config"
	"dnf-launcher-go/internal/db"
	appsettings "dnf-launcher-go/internal/settings"
)

func Direct(settings config.Settings, store *db.Store, uid int) (map[string]any, error) {
	accountName := ""
	_ = store.Game().QueryRow("SELECT accountname FROM accounts WHERE uid=?", uid).Scan(&accountName)
	token, err := createDNFLoginToken(settings, uid)
	if err != nil {
		return nil, fmt.Errorf("Failed to create DNF login token: %w", err)
	}
	return map[string]any{
		"uid":            uid,
		"account_name":   accountName,
		"dnf_token":      token,
		"client_pvf_md5": appsettings.ClientPVFMD5(store),
	}, nil
}

func createDNFLoginToken(settings config.Settings, uid int) (string, error) {
	rawHex := fmt.Sprintf("%08x010101010101010101010101010101010101010101010101010101010101010155914510010403030101", uid)
	raw, err := hexBytes(rawHex)
	if err != nil {
		return "", err
	}
	keyPath, err := filepath.Abs(settings.LoginPrivateKeyPath)
	if err != nil {
		return "", err
	}
	cmd := exec.Command("openssl", "rsautl", "-sign", "-inkey", keyPath)
	cmd.Stdin = bytes.NewReader(raw)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(out), nil
}

func hexBytes(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if len(value)%2 != 0 {
		return nil, fmt.Errorf("hex string length must be even")
	}
	out := make([]byte, 0, len(value)/2)
	for i := 0; i < len(value); i += 2 {
		var b byte
		for _, c := range value[i : i+2] {
			b <<= 4
			switch {
			case c >= '0' && c <= '9':
				b |= byte(c - '0')
			case c >= 'a' && c <= 'f':
				b |= byte(c-'a') + 10
			case c >= 'A' && c <= 'F':
				b |= byte(c-'A') + 10
			default:
				return nil, fmt.Errorf("invalid hex character")
			}
		}
		out = append(out, b)
	}
	return out, nil
}

func ResolveAccount(conn *sql.DB, uid int) (string, error) {
	var accountName string
	err := conn.QueryRow("SELECT accountname FROM accounts WHERE uid=?", uid).Scan(&accountName)
	return accountName, err
}
