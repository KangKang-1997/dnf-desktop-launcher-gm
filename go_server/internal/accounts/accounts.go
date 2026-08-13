package accounts

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"

	"dnf-launcher-go/internal/apperror"
	"dnf-launcher-go/internal/config"
	"dnf-launcher-go/internal/db"
)

type Account struct {
	UID         int
	AccountName string
}

func Authenticate(store *db.Store, accountName string, password string) (Account, error) {
	var account Account
	err := store.Game().QueryRow(
		"SELECT uid, accountname FROM accounts WHERE accountname=? AND password=?",
		accountName,
		md5Hex(password),
	).Scan(&account.UID, &account.AccountName)
	if err == sql.ErrNoRows {
		return Account{}, apperror.Unauthorized("Invalid account or password")
	}
	if err != nil {
		return Account{}, err
	}
	return account, nil
}

func Register(settings config.Settings, store *db.Store, accountName string, password string, qq string) (int, error) {
	passwordMD5 := md5Hex(password)
	var uid int
	var existingPassword string
	err := store.Game().QueryRow(
		"SELECT uid, password FROM accounts WHERE accountname=?",
		accountName,
	).Scan(&uid, &existingPassword)
	if err == nil {
		if existingPassword != passwordMD5 {
			return 0, apperror.BadRequest("Account already exists")
		}
		return uid, initializeAccount(settings, store, uid)
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	_, err = store.Game().Exec(
		"INSERT INTO accounts(accountname, password, qq) VALUES(?, ?, ?)",
		accountName,
		passwordMD5,
		qq,
	)
	if err != nil {
		return 0, apperror.BadRequest("Account already exists")
	}
	err = store.Game().QueryRow(
		"SELECT uid FROM accounts WHERE accountname=? AND password=?",
		accountName,
		passwordMD5,
	).Scan(&uid)
	if err != nil {
		return 0, apperror.Internal("Register failed, please retry")
	}
	return uid, initializeAccount(settings, store, uid)
}

func ChangePasswordByQQ(store *db.Store, accountName string, qq string, newPassword string) (int, error) {
	var uid int
	err := store.Game().QueryRow(
		"SELECT uid FROM accounts WHERE accountname=? AND qq=?",
		accountName,
		qq,
	).Scan(&uid)
	if err == sql.ErrNoRows {
		return 0, apperror.Unauthorized("Account or QQ is incorrect")
	}
	if err != nil {
		return 0, err
	}
	_, err = store.Game().Exec(
		"UPDATE accounts SET password=? WHERE uid=?",
		md5Hex(newPassword),
		uid,
	)
	if err != nil {
		return 0, err
	}
	return uid, nil
}

func initializeAccount(settings config.Settings, store *db.Store, uid int) error {
	statements := []struct {
		db     *sql.DB
		exists string
		insert string
		args   []any
	}{
		{store.Game(), "SELECT m_id FROM limit_create_character WHERE m_id=?", "INSERT INTO limit_create_character (m_id) VALUES (?)", []any{uid}},
		{store.Game(), "SELECT m_id FROM member_info WHERE m_id=?", "INSERT INTO member_info (m_id, user_id) VALUES (?, ?)", []any{uid, uid}},
		{store.Game(), "SELECT m_id FROM member_join_info WHERE m_id=?", "INSERT INTO member_join_info (m_id) VALUES (?)", []any{uid}},
		{store.Game(), "SELECT m_id FROM member_miles WHERE m_id=?", "INSERT INTO member_miles (m_id) VALUES (?)", []any{uid}},
		{store.Game(), "SELECT m_id FROM member_white_account WHERE m_id=?", "INSERT INTO member_white_account (m_id) VALUES (?)", []any{uid}},
	}
	for _, statement := range statements {
		if err := ensureRow(statement.db, statement.exists, statement.insert, uid, statement.args...); err != nil {
			return err
		}
	}
	return nil
}

func ensureRow(conn *sql.DB, existsSQL string, insertSQL string, uid int, insertArgs ...any) error {
	var mID int
	err := conn.QueryRow(existsSQL, uid).Scan(&mID)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}
	_, err = conn.Exec(insertSQL, insertArgs...)
	return err
}

func md5Hex(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}
