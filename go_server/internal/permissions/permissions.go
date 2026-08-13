package permissions

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
)

var All = []string{
	"gm.mail",
	"gm.cera.charge",
	"gm.character.edit",
	"gm.inventory",
	"gm.event.manage",
	"gm.avatar.edit",
}

var DefaultPlayer = []string{}

var GMTools = []map[string]string{
	{"id": "mail", "name": "邮件", "permission": "gm.mail"},
	{"id": "cera_charge", "name": "充值点券", "permission": "gm.cera.charge"},
	{"id": "character_edit", "name": "角色修改", "permission": "gm.character.edit"},
	{"id": "inventory", "name": "背包", "permission": "gm.inventory"},
	{"id": "event_manage", "name": "活动管理", "permission": "gm.event.manage"},
	{"id": "avatar_edit", "name": "时装潜能", "permission": "gm.avatar.edit"},
}

func EnsureTables(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS admins (
			id INT PRIMARY KEY AUTO_INCREMENT,
			username VARCHAR(64) NOT NULL UNIQUE,
			password_md5 CHAR(32) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS roles (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(64) NOT NULL UNIQUE,
			permissions TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS account_roles (
			uid INT PRIMARY KEY,
			role_id INT NOT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS operation_logs (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uid INT NOT NULL,
			action VARCHAR(128) NOT NULL,
			detail TEXT NULL,
			ip VARCHAR(64) NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			name VARCHAR(64) PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	if _, err := upsertRole(db, "player", DefaultPlayer); err != nil {
		return err
	}
	if _, err := upsertRole(db, "admin", All); err != nil {
		return err
	}
	return ensureDefaultAdmin(db)
}

func VerifyAdmin(db *sql.DB, username string, password string) (int, string, bool, error) {
	var id int
	var account string
	err := db.QueryRow(
		"SELECT id, username FROM admins WHERE username=? AND password_md5=?",
		username,
		md5Hex(password),
	).Scan(&id, &account)
	if err == sql.ErrNoRows {
		return 0, "", false, nil
	}
	if err != nil {
		return 0, "", false, err
	}
	return -id, account, true, nil
}

func ChangeAdminPassword(db *sql.DB, uid int, currentPassword string, newPassword string) (bool, error) {
	adminID := uid
	if adminID < 0 {
		adminID = -adminID
	}
	var id int
	err := db.QueryRow(
		"SELECT id FROM admins WHERE id=? AND password_md5=?",
		adminID,
		md5Hex(currentPassword),
	).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	_, err = db.Exec(
		"UPDATE admins SET password_md5=? WHERE id=?",
		md5Hex(newPassword),
		adminID,
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func AccountPermissions(db *sql.DB, uid int) ([]string, error) {
	var payload string
	err := db.QueryRow(`
		SELECT r.permissions
		FROM account_roles ar
		JOIN roles r ON r.id = ar.role_id
		WHERE ar.uid=?
	`, uid).Scan(&payload)
	if err == sql.ErrNoRows {
		return append([]string{}, DefaultPlayer...), nil
	}
	if err != nil {
		return nil, err
	}
	var permissions []string
	if err := json.Unmarshal([]byte(payload), &permissions); err != nil {
		return append([]string{}, DefaultPlayer...), nil
	}
	return normalize(permissions), nil
}

func SetAccountPermissions(db *sql.DB, uid int, values []string) ([]string, error) {
	normalized := normalize(values)
	roleID, err := upsertRole(db, "account:"+itoa(uid), normalized)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		INSERT INTO account_roles(uid, role_id) VALUES(?, ?)
		ON DUPLICATE KEY UPDATE role_id=VALUES(role_id)
	`, uid, roleID)
	if err != nil {
		return nil, err
	}
	return normalized, nil
}

func VisibleTools(userPermissions []string) []map[string]string {
	allowed := map[string]bool{}
	for _, item := range userPermissions {
		allowed[item] = true
	}
	tools := []map[string]string{}
	for _, tool := range GMTools {
		if allowed[tool["permission"]] {
			tools = append(tools, tool)
		}
	}
	return tools
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

func IsAllowed(permission string, permissions []string) bool {
	for _, item := range permissions {
		if item == permission {
			return true
		}
	}
	return false
}

func ensureDefaultAdmin(db *sql.DB) error {
	var id int
	err := db.QueryRow("SELECT id FROM admins WHERE username=?", "admin").Scan(&id)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}
	_, err = db.Exec(
		"INSERT INTO admins(username, password_md5) VALUES(?, ?)",
		"admin",
		md5Hex("admin"),
	)
	return err
}

func upsertRole(db *sql.DB, name string, values []string) (int, error) {
	payload, err := json.Marshal(values)
	if err != nil {
		return 0, err
	}
	var id int
	err = db.QueryRow("SELECT id FROM roles WHERE name=?", name).Scan(&id)
	if err == nil {
		_, err = db.Exec("UPDATE roles SET permissions=? WHERE id=?", string(payload), id)
		return id, err
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	if _, err := db.Exec("INSERT INTO roles(name, permissions) VALUES(?, ?)", name, string(payload)); err != nil {
		return 0, err
	}
	err = db.QueryRow("SELECT id FROM roles WHERE name=?", name).Scan(&id)
	return id, err
}

func normalize(values []string) []string {
	allowed := map[string]bool{}
	for _, item := range All {
		allowed[item] = true
	}
	result := []string{}
	for _, item := range values {
		if allowed[item] {
			result = append(result, item)
		}
	}
	return result
}

func md5Hex(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}
