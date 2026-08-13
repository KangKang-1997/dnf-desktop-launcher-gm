package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"dnf-launcher-go/internal/apperror"
	"dnf-launcher-go/internal/config"
)

type User struct {
	UID         int
	AccountName string
	UserType    string
	Permissions []string
	CanLaunch   bool
}

type tokenPayload struct {
	UID         int    `json:"uid"`
	AccountName string `json:"account_name"`
	UserType    string `json:"user_type"`
	Exp         int64  `json:"exp"`
}

func CreateSessionToken(settings config.Settings, uid int, accountName string, userType string) (string, error) {
	if userType == "" {
		userType = "game"
	}
	payload := tokenPayload{
		UID:         uid,
		AccountName: accountName,
		UserType:    userType,
		Exp:         time.Now().Unix() + int64(settings.SessionTTLSeconds),
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	body := b64url(bodyBytes)
	mac := hmac.New(sha256.New, []byte(settings.SessionSecret))
	_, _ = mac.Write([]byte(body))
	return body + "." + b64url(mac.Sum(nil)), nil
}

func ParseSessionToken(settings config.Settings, token string) (User, error) {
	body, sig, ok := strings.Cut(token, ".")
	if !ok || body == "" || sig == "" {
		return User{}, apperror.Unauthorized("Invalid session token")
	}
	expected := sign(settings.SessionSecret, body)
	actual, err := unb64url(sig)
	if err != nil || !hmac.Equal(actual, expected) {
		return User{}, apperror.Unauthorized("Invalid session token")
	}
	bodyBytes, err := unb64url(body)
	if err != nil {
		return User{}, apperror.Unauthorized("Invalid session token")
	}
	var payload tokenPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return User{}, apperror.Unauthorized("Invalid session token")
	}
	if payload.Exp < time.Now().Unix() {
		return User{}, apperror.Unauthorized("Session expired")
	}
	return User{
		UID:         payload.UID,
		AccountName: payload.AccountName,
		UserType:    payload.UserType,
		CanLaunch:   payload.UserType == "game",
	}, nil
}

func BearerToken(authorization string) (string, error) {
	if authorization == "" {
		return "", apperror.Unauthorized("Missing bearer token")
	}
	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", apperror.Unauthorized("Missing bearer token")
	}
	return strings.TrimSpace(parts[1]), nil
}

func UIDFromPath(path string, prefix string, suffix string) (int, error) {
	value := strings.TrimPrefix(path, prefix)
	value = strings.TrimSuffix(value, suffix)
	uid, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("invalid uid")
	}
	return uid, nil
}

func sign(secret string, body string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(body))
	return mac.Sum(nil)
}

func b64url(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func unb64url(value string) ([]byte, error) {
	padding := len(value) % 4
	if padding > 0 {
		value += strings.Repeat("=", 4-padding)
	}
	return base64.URLEncoding.DecodeString(value)
}
