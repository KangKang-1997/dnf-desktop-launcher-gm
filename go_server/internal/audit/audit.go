package audit

import (
	"database/sql"
	"net"
	"net/http"
)

func Write(db *sql.DB, r *http.Request, uid int, action string, detail string) {
	ip := clientIP(r)
	_, _ = db.Exec(
		"INSERT INTO operation_logs(uid, action, detail, ip) VALUES(?, ?, ?, ?)",
		uid,
		action,
		detail,
		ip,
	)
}

func clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
