package settings

import (
	"database/sql"
	"encoding/json"
	"strings"

	"dnf-launcher-go/internal/apperror"
	"dnf-launcher-go/internal/db"
)

const DefaultHomeTitle = "欢迎回来，勇士"
const DefaultHomeEyebrow = "冒险准备完成"

type Announcement struct {
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Content   string `json:"content"`
	PosterURL string `json:"poster_url"`
}

func Public(store *db.Store) map[string]any {
	return map[string]any{
		"home_title":          get(store, "home_title", DefaultHomeTitle),
		"home_eyebrow":        get(store, "home_eyebrow", DefaultHomeEyebrow),
		"client_download_url": get(store, "client_download_url", ""),
		"announcements":       announcements(store),
	}
}

func UpdateHome(store *db.Store, homeTitle string, homeEyebrow string, clientDownloadURL string, rows []Announcement) (map[string]any, error) {
	homeTitle = strings.TrimSpace(homeTitle)
	if homeTitle == "" {
		homeTitle = DefaultHomeTitle
	}
	homeEyebrow = strings.TrimSpace(homeEyebrow)
	if homeEyebrow == "" {
		homeEyebrow = DefaultHomeEyebrow
	}
	clientDownloadURL = strings.TrimSpace(clientDownloadURL)
	if clientDownloadURL != "" && !(strings.HasPrefix(clientDownloadURL, "http://") || strings.HasPrefix(clientDownloadURL, "https://")) {
		return nil, apperror.BadRequest("Client download URL must use http:// or https://")
	}

	normalized := []Announcement{}
	for _, row := range rows {
		if len(normalized) >= 8 {
			break
		}
		title := strings.TrimSpace(row.Title)
		if title == "" {
			continue
		}
		posterURL := strings.TrimSpace(row.PosterURL)
		if posterURL != "" &&
			!(strings.HasPrefix(posterURL, "http://") ||
				strings.HasPrefix(posterURL, "https://") ||
				strings.HasPrefix(posterURL, "/api/posters/")) {
			return nil, apperror.BadRequest("Poster URL must use http://, https://, or /api/posters/")
		}
		normalized = append(normalized, Announcement{
			Title:     title,
			Summary:   strings.TrimSpace(row.Summary),
			Content:   strings.TrimSpace(row.Content),
			PosterURL: posterURL,
		})
	}
	if len(normalized) == 0 {
		normalized = defaultAnnouncements()
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	if err := set(store, "home_title", homeTitle); err != nil {
		return nil, err
	}
	if err := set(store, "home_eyebrow", homeEyebrow); err != nil {
		return nil, err
	}
	if err := set(store, "client_download_url", clientDownloadURL); err != nil {
		return nil, err
	}
	if err := set(store, "home_announcements", string(payload)); err != nil {
		return nil, err
	}
	return Public(store), nil
}

func ClientPVFMD5(store *db.Store) string {
	return get(store, "client_pvf_md5", "")
}

func SetClientPVFMD5(store *db.Store, value string) error {
	return set(store, "client_pvf_md5", value)
}

func get(store *db.Store, name string, fallback string) string {
	var value string
	err := store.Tool().QueryRow("SELECT value FROM settings WHERE name=?", name).Scan(&value)
	if err == sql.ErrNoRows || err != nil {
		return fallback
	}
	return value
}

func announcements(store *db.Store) []Announcement {
	payload := get(store, "home_announcements", "")
	if payload == "" {
		return defaultAnnouncements()
	}
	var rows []Announcement
	if err := json.Unmarshal([]byte(payload), &rows); err != nil {
		return defaultAnnouncements()
	}
	normalized := []Announcement{}
	for _, row := range rows {
		title := strings.TrimSpace(row.Title)
		if title == "" {
			continue
		}
		normalized = append(normalized, Announcement{
			Title:     title,
			Summary:   strings.TrimSpace(row.Summary),
			Content:   strings.TrimSpace(row.Content),
			PosterURL: strings.TrimSpace(row.PosterURL),
		})
	}
	if len(normalized) == 0 {
		return defaultAnnouncements()
	}
	return normalized
}

func set(store *db.Store, name string, value string) error {
	_, err := store.Tool().Exec(`
		INSERT INTO settings(name, value) VALUES(?, ?)
		ON DUPLICATE KEY UPDATE value=VALUES(value)
	`, name, value)
	return err
}

func defaultAnnouncements() []Announcement {
	return []Announcement{
		{
			Title:     "版本更新公告标题占位",
			Summary:   "后续可替换为具体版本更新内容",
			Content:   "这里用于展示版本更新全文，管理员可在权限管理页修改。",
			PosterURL: "/api/posters/sample-1",
		},
		{
			Title:     "客户端下载说明占位",
			Summary:   "后续可替换为客户端下载说明",
			Content:   "这里用于展示客户端下载说明全文，管理员可在权限管理页修改。",
			PosterURL: "/api/posters/sample-2",
		},
		{
			Title:     "活动与维护通知占位",
			Summary:   "后续可替换为活动、维护或补偿通知",
			Content:   "这里用于展示活动、维护或补偿通知全文，管理员可在权限管理页修改。",
			PosterURL: "/api/posters/sample-3",
		},
	}
}
