package gm

import (
	"bytes"
	"compress/zlib"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"dnf-launcher-go/internal/apperror"
	"dnf-launcher-go/internal/db"
	"dnf-launcher-go/internal/pvf"
	"dnf-launcher-go/internal/security"
)

var pvpRankList = append(
	segmentNames("级", 10, 1),
	append(
		segmentNames("段", 1, 10),
		[]string{"至尊1", "至尊2", "至尊3", "至尊4", "至尊5", "至尊6", "至尊7", "至尊8", "至尊9", "至尊10", "达人", "名人", "小霸王", "霸王", "斗神"}...,
	)...,
)

var expertJobMap = map[int]string{
	0: "无职业",
	1: "附魔师",
	2: "炼金术师",
	3: "分解师",
	4: "控偶师",
}

var itemTypeNames = map[byte]string{
	0x00: "已删除/空槽位",
	0x01: "装备",
	0x02: "消耗品",
	0x03: "材料",
	0x04: "任务材料",
	0x05: "宠物",
	0x06: "宠物装备",
	0x07: "宠物消耗品",
	0x0A: "副职业",
}

var increaseTypeNames = map[byte]string{
	0x00: "空",
	0x01: "异次元体力",
	0x02: "异次元精神",
	0x03: "异次元力量",
	0x04: "异次元智力",
}

var inventoryScopes = map[string]inventoryScope{
	"inventory": {
		database: "taiwan_cain_2nd",
		table:    "inventory",
		column:   "inventory",
		name:     "物品栏",
		where:    "charac_no",
	},
	"equipslot": {
		database: "taiwan_cain_2nd",
		table:    "inventory",
		column:   "equipslot",
		name:     "穿戴栏",
		where:    "charac_no",
	},
	"creature": {
		database: "taiwan_cain_2nd",
		table:    "inventory",
		column:   "creature",
		name:     "宠物栏",
		where:    "charac_no",
	},
	"cargo": {
		database: "taiwan_cain_2nd",
		table:    "charac_inven_expand",
		column:   "cargo",
		name:     "角色仓库",
		where:    "charac_no",
	},
	"account_cargo": {
		database: "taiwan_cain",
		table:    "account_cargo",
		column:   "cargo",
		name:     "账号仓库",
		where:    "m_id",
	},
}

var avatarHiddenMap = map[string]string{
	"physical attack":             "力量",
	"magical attack":              "智力",
	"physical defense":            "体力",
	"magical defense":             "精神",
	"HP MAX":                      "HP MAX",
	"MP MAX":                      "MP MAX",
	"HP regen speed":              "HP 恢复",
	"MP Regen speed":              "MP 恢复",
	"attack speed":                "攻击速度",
	"move speed":                  "移动速度",
	"cast speed":                  "施放速度",
	"inventory limit":             "负重上限",
	"stuck":                       "命中率",
	"stuck resistance":            "回避率",
	"all activestatus resistance": "异常抗性",
	"hit recovery":                "硬直",
	"equipment magical defence":   "魔法防御",
	"equipment physical defence":  "物理防御",
	"jump power":                  "跳跃力",
	"physical critical hit":       "物理暴击",
	"magical critical hit":        "魔法暴击",
	"":                            "",
}

var cp1252Encode = map[rune]byte{
	'\u20AC': 0x80,
	'\u201A': 0x82,
	'\u0192': 0x83,
	'\u201E': 0x84,
	'\u2026': 0x85,
	'\u2020': 0x86,
	'\u2021': 0x87,
	'\u02C6': 0x88,
	'\u2030': 0x89,
	'\u0160': 0x8A,
	'\u2039': 0x8B,
	'\u0152': 0x8C,
	'\u017D': 0x8E,
	'\u2018': 0x91,
	'\u2019': 0x92,
	'\u201C': 0x93,
	'\u201D': 0x94,
	'\u2022': 0x95,
	'\u2013': 0x96,
	'\u2014': 0x97,
	'\u02DC': 0x98,
	'\u2122': 0x99,
	'\u0161': 0x9A,
	'\u203A': 0x9B,
	'\u0153': 0x9C,
	'\u017E': 0x9E,
	'\u0178': 0x9F,
}

type Character struct {
	UID           int    `json:"uid"`
	CharacNo      int    `json:"charac_no"`
	CharacName    string `json:"charac_name"`
	Level         int    `json:"level"`
	Job           int    `json:"job"`
	GrowType      int    `json:"grow_type"`
	GrowTypeBase  int    `json:"grow_type_base"`
	WakeFlag      int    `json:"wake_flag"`
	JobName       string `json:"job_name"`
	DeleteFlag    int    `json:"delete_flag"`
	ExpertJob     int    `json:"expert_job"`
	ExpertJobName string `json:"expert_job_name"`
	PVPGrade      int    `json:"pvp_grade"`
	PVPGradeName  string `json:"pvp_grade_name"`
	PVPWin        int    `json:"pvp_win"`
	PVPPoint      int    `json:"pvp_point"`
	WinPoint      int    `json:"win_point"`
}

type Account struct {
	UID         int    `json:"uid"`
	AccountName string `json:"account_name"`
}

type CeraResult struct {
	UID         int    `json:"uid"`
	AccountName string `json:"account_name"`
	Cera        int    `json:"cera"`
	CeraPoint   int    `json:"cera_point"`
	CeraType    string `json:"cera_type,omitempty"`
	Action      string `json:"action,omitempty"`
}

type BanResult struct {
	UID            int    `json:"uid"`
	AccountName    string `json:"account_name"`
	Banned         bool   `json:"banned"`
	PunishType     any    `json:"punish_type"`
	PunishTypeName string `json:"punish_type_name"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	Reason         string `json:"reason"`
}

type EventOption struct {
	EventID      int    `json:"event_id"`
	EventName    string `json:"event_name"`
	EventExplain string `json:"event_explain"`
}

type RunningEvent struct {
	LogID        int64  `json:"log_id"`
	EventID      int    `json:"event_id"`
	EventName    string `json:"event_name"`
	EventExplain string `json:"event_explain"`
	Parameter1   int    `json:"parameter1"`
	Parameter2   int    `json:"parameter2"`
}

type AvatarHiddenOption struct {
	HiddenOption int    `json:"hidden_option"`
	Name         string `json:"name"`
}

type AvatarItem struct {
	UIID         int    `json:"ui_id"`
	ItemID       int    `json:"item_id"`
	ItemName     string `json:"item_name"`
	HiddenOption int    `json:"hidden_option"`
	HiddenName   string `json:"hidden_name"`
}

type InventoryItem struct {
	Slot             int    `json:"slot"`
	ItemID           int    `json:"item_id"`
	ItemName         string `json:"item_name"`
	ItemType         int    `json:"item_type"`
	ItemTypeName     string `json:"item_type_name"`
	IsSeal           int    `json:"is_seal"`
	EnhancementLevel int    `json:"enhancement_level"`
	CountOrGrade     int    `json:"count_or_grade"`
	Durability       int    `json:"durability"`
	Orb              int    `json:"orb"`
	IncreaseType     int    `json:"increase_type"`
	IncreaseTypeName string `json:"increase_type_name"`
	IncreaseValue    int    `json:"increase_value"`
	ForgeLevel       int    `json:"forge_level"`
}

type MailPayload struct {
	Message          string
	ItemID           int
	ItemCount        int
	Gold             int
	ItemType         string
	ItemGrade        int
	EnhancementLevel int
	ForgeLevel       int
	AmplifyOption    int
	AmplifyValue     int
}

type MailSendResult struct {
	OK          bool      `json:"ok"`
	LetterID    int64     `json:"letter_id"`
	PostalCount int       `json:"postal_count"`
	Character   Character `json:"character"`
}

type MailMassSendResult struct {
	OK            bool  `json:"ok"`
	TargetCount   int   `json:"target_count"`
	FirstLetterID int64 `json:"first_letter_id"`
	PostalCount   int   `json:"postal_count"`
}

type MailDeleteResult struct {
	OK           bool       `json:"ok"`
	DeletedCount int64      `json:"deleted_count"`
	Character    *Character `json:"character,omitempty"`
}

type inventoryScope struct {
	database string
	table    string
	column   string
	name     string
	where    string
}

type inventoryTarget struct {
	character Character
	scopeKey  string
	scope     inventoryScope
	targetID  int
	blob      []byte
}

func ResolveAccount(store *db.Store, uid *int, accountName string) (Account, error) {
	accountName = strings.TrimSpace(accountName)
	var account Account
	var err error
	if uid != nil {
		err = store.Game().QueryRow(
			"SELECT uid, accountname FROM accounts WHERE uid=?",
			*uid,
		).Scan(&account.UID, &account.AccountName)
	} else if accountName != "" {
		err = store.Game().QueryRow(
			"SELECT uid, accountname FROM accounts WHERE accountname=?",
			accountName,
		).Scan(&account.UID, &account.AccountName)
	} else {
		return Account{}, apperror.BadRequest("Account name or UID is required")
	}
	if err == sql.ErrNoRows {
		return Account{}, apperror.New(404, "Account not found")
	}
	if err != nil {
		return Account{}, err
	}
	return account, nil
}

func ResolveAccessibleAccount(store *db.Store, user security.User, uid *int, accountName string) (Account, error) {
	if user.UserType == "admin" {
		return ResolveAccount(store, uid, accountName)
	}
	currentUID := user.UID
	currentName := strings.TrimSpace(user.AccountName)
	if uid != nil && *uid != currentUID {
		return Account{}, apperror.New(404, "Account not found")
	}
	if strings.TrimSpace(accountName) != "" && strings.TrimSpace(accountName) != currentName {
		return Account{}, apperror.New(404, "Account not found")
	}
	return ResolveAccount(store, &currentUID, "")
}

func CeraResponse(store *db.Store, account Account) (CeraResult, error) {
	if err := ensureCeraRows(store, account.UID); err != nil {
		return CeraResult{}, err
	}
	cera, err := ceraValue(store, account.UID)
	if err != nil {
		return CeraResult{}, err
	}
	ceraPoint, err := ceraPointValue(store, account.UID)
	if err != nil {
		return CeraResult{}, err
	}
	return CeraResult{
		UID:         account.UID,
		AccountName: account.AccountName,
		Cera:        cera,
		CeraPoint:   ceraPoint,
	}, nil
}

func ChargeCera(store *db.Store, account Account, ceraType string, action string, amount int) (CeraResult, error) {
	ceraType, err := normalizeCeraType(ceraType)
	if err != nil {
		return CeraResult{}, err
	}
	action, err = normalizeCeraAction(action)
	if err != nil {
		return CeraResult{}, err
	}
	if action == "add" && amount <= 0 {
		return CeraResult{}, apperror.BadRequest("Amount must be greater than 0")
	}
	billing := storeBilling(store)
	tx, err := billing.Begin()
	if err != nil {
		return CeraResult{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := ensureCeraRowsTx(tx, account.UID); err != nil {
		return CeraResult{}, err
	}
	if ceraType == "cera" {
		if action == "add" {
			_, err = tx.Exec("UPDATE cash_cera SET cera=cera+?, mod_date=NOW() WHERE account=?", amount, account.UID)
		} else {
			_, err = tx.Exec("UPDATE cash_cera SET cera=?, mod_date=NOW() WHERE account=?", amount, account.UID)
		}
	} else {
		if action == "add" {
			_, err = tx.Exec("UPDATE cash_cera_point SET cera_point=cera_point+?, mod_date=NOW() WHERE account=?", amount, account.UID)
		} else {
			_, err = tx.Exec("UPDATE cash_cera_point SET cera_point=?, mod_date=NOW() WHERE account=?", amount, account.UID)
		}
	}
	if err != nil {
		return CeraResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return CeraResult{}, err
	}
	committed = true
	result, err := CeraResponse(store, account)
	if err != nil {
		return CeraResult{}, err
	}
	result.CeraType = ceraType
	result.Action = action
	return result, nil
}

func BanStatus(store *db.Store, account Account) (BanResult, error) {
	result := BanResult{
		UID:            account.UID,
		AccountName:    account.AccountName,
		Banned:         false,
		PunishType:     nil,
		PunishTypeName: "",
		StartTime:      "",
		EndTime:        "",
		Reason:         "",
	}
	row := store.Game().QueryRow(`
		SELECT punish_type, start_time, end_time, reason
		FROM member_punish_info
		WHERE m_id=? AND apply_flag!=0 AND end_time>NOW()
		ORDER BY end_time DESC
		LIMIT 1
	`, account.UID)
	var punishType int
	var startTime time.Time
	var endTime time.Time
	var reason sql.NullString
	if err := row.Scan(&punishType, &startTime, &endTime, &reason); err != nil {
		if err == sql.ErrNoRows {
			return result, nil
		}
		return BanResult{}, err
	}
	result.Banned = true
	result.PunishType = punishType
	result.PunishTypeName = punishTypeName(punishType)
	result.StartTime = startTime.Format("2006-01-02 15:04:05")
	result.EndTime = endTime.Format("2006-01-02 15:04:05")
	if reason.Valid {
		result.Reason = reason.String
	}
	return result, nil
}

func SetBan(store *db.Store, account Account, punishType int, days int, reason string) (BanResult, error) {
	if punishType != 1 && punishType != 4 {
		return BanResult{}, apperror.BadRequest("Invalid punish type")
	}
	if days < 1 || days > 3650 {
		return BanResult{}, apperror.BadRequest("Invalid ban days")
	}
	now := time.Now()
	endTime := now.Add(time.Duration(days) * 24 * time.Hour)
	game := store.Game()
	tx, err := game.Begin()
	if err != nil {
		return BanResult{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if _, err := tx.Exec("DELETE FROM member_punish_info WHERE m_id=?", account.UID); err != nil {
		return BanResult{}, err
	}
	if _, err := tx.Exec(`
		REPLACE INTO member_punish_info
			(m_id, punish_type, occ_time, punish_value, apply_flag, start_time, end_time, reason)
		VALUES(?, ?, ?, ?, 2, ?, ?, ?)
	`,
		account.UID,
		punishType,
		now.Format("2006-01-02 15:04:05"),
		101,
		now.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		strings.TrimSpace(reason),
	); err != nil {
		return BanResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return BanResult{}, err
	}
	committed = true
	return BanStatus(store, account)
}

func Unban(store *db.Store, account Account) (BanResult, error) {
	if _, err := store.Game().Exec("DELETE FROM member_punish_info WHERE m_id=?", account.UID); err != nil {
		return BanResult{}, err
	}
	return BanStatus(store, account)
}

func Events(store *db.Store) (map[string]any, error) {
	available, err := EventOptions(store)
	if err != nil {
		return nil, err
	}
	running, err := RunningEvents(store)
	if err != nil {
		return nil, err
	}
	return map[string]any{"available": available, "running": running}, nil
}

func EventOptions(store *db.Store) ([]EventOption, error) {
	rows, err := store.Game().Query(`
		SELECT event_id, event_name, event_explain
		FROM dnf_event_info
		ORDER BY event_id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []EventOption{}
	for rows.Next() {
		var item EventOption
		var name sql.NullString
		var explain sql.NullString
		if err := rows.Scan(&item.EventID, &name, &explain); err != nil {
			return nil, err
		}
		if name.Valid {
			item.EventName = name.String
		}
		if explain.Valid {
			item.EventExplain = explain.String
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func RunningEvents(store *db.Store) ([]RunningEvent, error) {
	options, err := EventOptions(store)
	if err != nil {
		return nil, err
	}
	optionMap := map[int]EventOption{}
	for _, item := range options {
		optionMap[item.EventID] = item
	}
	rows, err := store.Game().Query(`
		SELECT log_id, event_type, parameter1, parameter2
		FROM dnf_event_log
		ORDER BY log_id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []RunningEvent{}
	for rows.Next() {
		var item RunningEvent
		var parameter1 sql.NullInt64
		var parameter2 sql.NullInt64
		if err := rows.Scan(&item.LogID, &item.EventID, &parameter1, &parameter2); err != nil {
			return nil, err
		}
		option := optionMap[item.EventID]
		item.EventName = option.EventName
		item.EventExplain = option.EventExplain
		if item.EventExplain == "" {
			item.EventExplain = "explain"
		}
		if parameter1.Valid {
			item.Parameter1 = int(parameter1.Int64)
		}
		if parameter2.Valid {
			item.Parameter2 = int(parameter2.Int64)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func AddEvent(store *db.Store, eventID int, parameter1 int, parameter2 int) ([]RunningEvent, error) {
	_, err := store.Game().Exec(`
		INSERT INTO dnf_event_log(
			occ_time, event_type, parameter1, parameter2, server_id,
			event_flag, start_time, end_time
		)
		VALUES(0, ?, ?, ?, 0, 0, 0, 0)
	`, eventID, parameter1, parameter2)
	if err != nil {
		return nil, err
	}
	return RunningEvents(store)
}

func DeleteEvent(store *db.Store, logID int) ([]RunningEvent, error) {
	if _, err := store.Game().Exec("DELETE FROM dnf_event_log WHERE log_id=?", logID); err != nil {
		return nil, err
	}
	return RunningEvents(store)
}

func Characters(store *db.Store, user security.User, keyword string, page int, limit int, includeDeleted bool, targetUID *int) (map[string]any, error) {
	page = clamp(page, 1, 2147483647)
	limit = clamp(limit, 1, 500)
	offset := (page - 1) * limit
	keyword = strings.TrimSpace(keyword)
	if keyword != "" && !isDigits(keyword) {
		return map[string]any{"characters": []Character{}, "page": page, "limit": limit, "total": 0}, nil
	}

	clauses := []string{}
	args := []any{}
	if !includeDeleted {
		clauses = append(clauses, "c.delete_flag=0")
	}
	if user.UserType != "admin" {
		clauses = append(clauses, "c.m_id=?")
		args = append(args, user.UID)
	} else if targetUID != nil {
		clauses = append(clauses, "c.m_id=?")
		args = append(args, *targetUID)
	}
	if keyword != "" {
		clauses = append(clauses, "c.charac_no=?")
		args = append(args, parseInt(keyword, 0))
	}
	where := "1=1"
	if len(clauses) > 0 {
		where = strings.Join(clauses, " AND ")
	}

	var total int
	if err := storeGameCain(store).QueryRow(
		"SELECT COUNT(*) AS total FROM charac_info c JOIN pvp_result p ON p.charac_no=c.charac_no WHERE "+where,
		args...,
	).Scan(&total); err != nil {
		return nil, err
	}

	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := storeGameCain(store).Query(`
		SELECT
			c.m_id,
			c.charac_no,
			HEX(c.charac_name) AS charac_name_hex,
			c.lev,
			c.job,
			c.grow_type,
			c.delete_flag,
			c.expert_job,
			p.pvp_grade AS pvp_grade,
			p.win AS pvp_win,
			p.pvp_point AS pvp_point,
			p.win_point AS win_point
		FROM charac_info c
		JOIN pvp_result p ON p.charac_no=c.charac_no
		WHERE `+where+`
		ORDER BY c.charac_no DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobMap, err := pvf.Jobs(store)
	if err != nil {
		return nil, err
	}
	characters := []Character{}
	for rows.Next() {
		item, err := scanCharacter(rows, jobMap)
		if err != nil {
			return nil, err
		}
		characters = append(characters, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return map[string]any{
		"characters": characters,
		"page":       page,
		"limit":      limit,
		"total":      total,
	}, nil
}

func CharacterJobOptions(store *db.Store) (map[string]any, error) {
	return pvf.JobOptions(store, expertJobName, pvpRankList)
}

func AccessibleCharacter(store *db.Store, user security.User, characNo int, allowDeleted bool) (Character, error) {
	character, err := CharacterByNo(store, characNo)
	if err != nil {
		return Character{}, err
	}
	if !allowDeleted && character.DeleteFlag != 0 {
		return Character{}, apperror.New(404, "Character not found")
	}
	if user.UserType != "admin" && character.UID != user.UID {
		return Character{}, apperror.New(404, "Character not found")
	}
	return character, nil
}

func CharacterByNo(store *db.Store, characNo int) (Character, error) {
	row := storeGameCain(store).QueryRow(`
		SELECT
			c.m_id,
			c.charac_no,
			HEX(c.charac_name) AS charac_name_hex,
			c.lev,
			c.job,
			c.grow_type,
			c.delete_flag,
			c.expert_job,
			p.pvp_grade AS pvp_grade,
			p.win AS pvp_win,
			p.pvp_point AS pvp_point,
			p.win_point AS win_point
		FROM charac_info c
		JOIN pvp_result p ON p.charac_no=c.charac_no
		WHERE c.charac_no=?
	`, characNo)
	jobMap, err := pvf.Jobs(store)
	if err != nil {
		return Character{}, err
	}
	character, err := scanCharacter(row, jobMap)
	if err == sql.ErrNoRows {
		return Character{}, apperror.New(404, "Character not found")
	}
	if err != nil {
		return Character{}, err
	}
	return character, nil
}

func ActiveCharacterNos(store *db.Store) ([]int, error) {
	rows, err := storeGameCain(store).Query(`
		SELECT charac_no
		FROM charac_info
		WHERE delete_flag=0
		ORDER BY charac_no ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	characNos := []int{}
	for rows.Next() {
		var characNo int
		if err := rows.Scan(&characNo); err != nil {
			return nil, err
		}
		characNos = append(characNos, characNo)
	}
	return characNos, rows.Err()
}

func SetPVPGrade(store *db.Store, characNo int, pvpGrade int) (Character, Character, error) {
	if pvpGrade < 0 || pvpGrade >= len(pvpRankList) {
		return Character{}, Character{}, apperror.BadRequest("Invalid PVP grade")
	}
	before, err := CharacterByNo(store, characNo)
	if err != nil {
		return Character{}, Character{}, err
	}
	_, err = storeGameCain(store).Exec(
		"UPDATE pvp_result SET pvp_grade=? WHERE charac_no=?",
		pvpGrade,
		characNo,
	)
	if err != nil {
		return Character{}, Character{}, err
	}
	after, err := CharacterByNo(store, characNo)
	return before, after, err
}

func SetPVPPoint(store *db.Store, characNo int, pvpPoint int) (Character, Character, error) {
	before, err := CharacterByNo(store, characNo)
	if err != nil {
		return Character{}, Character{}, err
	}
	_, err = storeGameCain(store).Exec(
		"UPDATE pvp_result SET pvp_point=?, win_point=? WHERE charac_no=?",
		pvpPoint,
		pvpPoint,
		characNo,
	)
	if err != nil {
		return Character{}, Character{}, err
	}
	after, err := CharacterByNo(store, characNo)
	return before, after, err
}

func SetLevel(store *db.Store, characNo int, level int) (Character, Character, error) {
	before, err := CharacterByNo(store, characNo)
	if err != nil {
		return Character{}, Character{}, err
	}
	if before.Level != level {
		exp, err := pvf.GetExpForLevel(store, level)
		if err != nil {
			return Character{}, Character{}, err
		}
		if _, err := storeGameCain(store).Exec(
			"UPDATE charac_stat SET exp=? WHERE charac_no=?",
			exp,
			characNo,
		); err != nil {
			return Character{}, Character{}, err
		}
		if _, err := storeGameCain(store).Exec(
			"UPDATE charac_info SET lev=? WHERE charac_no=?",
			level,
			characNo,
		); err != nil {
			return Character{}, Character{}, err
		}
	}
	after, err := CharacterByNo(store, characNo)
	return before, after, err
}

func SetJob(store *db.Store, characNo int, job int, growType int, wakeFlag int, expertJob int) (Character, Character, error) {
	if growType < 0 || growType > 15 {
		return Character{}, Character{}, apperror.BadRequest("Invalid grow type")
	}
	if wakeFlag < 0 || wakeFlag > 2 {
		return Character{}, Character{}, apperror.BadRequest("Invalid wake flag")
	}
	if _, ok := expertJobMap[expertJob]; !ok {
		return Character{}, Character{}, apperror.BadRequest("Invalid expert job")
	}
	before, err := CharacterByNo(store, characNo)
	if err != nil {
		return Character{}, Character{}, err
	}
	jobMap, err := pvf.Jobs(store)
	if err != nil {
		return Character{}, Character{}, err
	}
	growMap, ok := jobMap[job]
	if len(jobMap) > 0 && !ok {
		return Character{}, Character{}, apperror.BadRequest("Invalid job")
	}
	if ok && len(growMap) > 0 {
		if _, exists := growMap[growType]; !exists {
			return Character{}, Character{}, apperror.BadRequest("Invalid grow type")
		}
	}
	storedGrowType := growType + wakeFlag*0x10
	_, err = storeGameCain(store).Exec(
		"UPDATE charac_info SET job=?, grow_type=?, expert_job=? WHERE charac_no=?",
		job,
		storedGrowType,
		expertJob,
		characNo,
	)
	if err != nil {
		return Character{}, Character{}, err
	}
	after, err := CharacterByNo(store, characNo)
	return before, after, err
}

func SetDeleteFlag(store *db.Store, characNo int, deleteFlag int) (Character, Character, error) {
	before, err := CharacterByNo(store, characNo)
	if err != nil {
		return Character{}, Character{}, err
	}
	_, err = storeGameCain(store).Exec(
		"UPDATE charac_info SET delete_flag=? WHERE charac_no=?",
		deleteFlag,
		characNo,
	)
	if err != nil {
		return Character{}, Character{}, err
	}
	after, err := CharacterByNo(store, characNo)
	return before, after, err
}

func SendMail(store *db.Store, character Character, sender string, payload MailPayload) (MailSendResult, error) {
	normalizeMailPayload(&payload)
	if err := validateMailItemCount(payload); err != nil {
		return MailSendResult{}, err
	}
	tx, err := store.MustGameNamed("taiwan_cain_2nd").Begin()
	if err != nil {
		return MailSendResult{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	letterID, postalCount, err := sendMailToCharacterTx(store, tx, character.CharacNo, sender, payload)
	if err != nil {
		return MailSendResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return MailSendResult{}, err
	}
	committed = true
	return MailSendResult{
		OK:          true,
		LetterID:    letterID,
		PostalCount: postalCount,
		Character:   character,
	}, nil
}

func SendMailAll(store *db.Store, sender string, payload MailPayload) (MailMassSendResult, error) {
	normalizeMailPayload(&payload)
	if payload.ItemID <= 0 && payload.Gold <= 0 && payload.Message == "" {
		return MailMassSendResult{}, apperror.BadRequest("Message, item, or gold is required")
	}
	if err := validateMailItemCount(payload); err != nil {
		return MailMassSendResult{}, err
	}
	characNos, err := ActiveCharacterNos(store)
	if err != nil {
		return MailMassSendResult{}, err
	}
	tx, err := store.MustGameNamed("taiwan_cain_2nd").Begin()
	if err != nil {
		return MailMassSendResult{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	result := MailMassSendResult{OK: true}
	for _, characNo := range characNos {
		letterID, postalCount, err := sendMailToCharacterTx(store, tx, characNo, sender, payload)
		if err != nil {
			return MailMassSendResult{}, err
		}
		if result.FirstLetterID == 0 {
			result.FirstLetterID = letterID
		}
		result.TargetCount++
		result.PostalCount += postalCount
	}
	if err := tx.Commit(); err != nil {
		return MailMassSendResult{}, err
	}
	committed = true
	return result, nil
}

func DeleteMailForCharacter(store *db.Store, character Character) (MailDeleteResult, error) {
	result, err := store.MustGameNamed("taiwan_cain_2nd").Exec(
		"UPDATE postal SET delete_flag=1 WHERE receive_charac_no=? AND delete_flag=0",
		character.CharacNo,
	)
	if err != nil {
		return MailDeleteResult{}, err
	}
	deletedCount, _ := result.RowsAffected()
	return MailDeleteResult{
		OK:           true,
		DeletedCount: deletedCount,
		Character:    &character,
	}, nil
}

func DeleteMailForAllCharacters(store *db.Store) (MailDeleteResult, error) {
	result, err := store.MustGameNamed("taiwan_cain_2nd").Exec(
		"UPDATE postal SET delete_flag=1 WHERE delete_flag=0",
	)
	if err != nil {
		return MailDeleteResult{}, err
	}
	deletedCount, _ := result.RowsAffected()
	return MailDeleteResult{
		OK:           true,
		DeletedCount: deletedCount,
	}, nil
}

func QueryInventory(store *db.Store, characNo int, scopeKey string) (map[string]any, error) {
	target, err := loadInventoryBlob(store, characNo, scopeKey)
	if err != nil {
		return nil, err
	}
	items, err := unpackInventoryBlob(store, target.blob)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"character":  target.character,
		"scope":      target.scopeKey,
		"scope_name": target.scope.name,
		"items":      items,
		"item_count": len(items),
	}, nil
}

func DeleteInventorySlots(store *db.Store, characNo int, scopeKey string, slots []int) (map[string]any, error) {
	target, err := loadInventoryBlob(store, characNo, scopeKey)
	if err != nil {
		return nil, err
	}
	newBlob, deletedSlots, err := buildDeletedInventoryBlob(slots, target.blob)
	if err != nil {
		return nil, err
	}
	if err := writeInventoryBlob(store, target.scope, target.targetID, newBlob); err != nil {
		return nil, err
	}
	items, err := unpackInventoryBlob(store, newBlob)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"character":     target.character,
		"scope":         target.scopeKey,
		"scope_name":    target.scope.name,
		"items":         items,
		"item_count":    len(items),
		"deleted_slots": deletedSlots,
	}, nil
}

func ClearInventoryScope(store *db.Store, characNo int, scopeKey string) (map[string]any, error) {
	target, err := loadInventoryBlob(store, characNo, scopeKey)
	if err != nil {
		return nil, err
	}
	slotCount, err := inventorySlotCount(target.blob)
	if err != nil {
		return nil, err
	}
	slots := make([]int, 0, slotCount)
	for slot := 0; slot < slotCount; slot++ {
		slots = append(slots, slot)
	}
	return DeleteInventorySlots(store, characNo, target.scopeKey, slots)
}

func AvatarHiddenOptions(store *db.Store) ([]AvatarHiddenOption, error) {
	var raw [][]string
	ok, err := pvf.Data(store, "avatar_hidden", &raw)
	if err != nil {
		return nil, err
	}
	options := []AvatarHiddenOption{{HiddenOption: 0, Name: "None"}}
	if !ok || len(raw) == 0 {
		return options, nil
	}
	for index, name := range raw[0] {
		displayName := name
		if mapped, ok := avatarHiddenMap[name]; ok {
			displayName = mapped
		}
		options = append(options, AvatarHiddenOption{
			HiddenOption: index + 1,
			Name:         displayName,
		})
	}
	return options, nil
}

func AvatarHiddenName(store *db.Store, value int) (string, error) {
	options, err := AvatarHiddenOptions(store)
	if err != nil {
		return "", err
	}
	for _, option := range options {
		if option.HiddenOption == value {
			return option.Name, nil
		}
	}
	return fmt.Sprint(value), nil
}

func QueryAvatarItems(store *db.Store, characNo int) (map[string]any, error) {
	character, err := CharacterByNo(store, characNo)
	if err != nil {
		return nil, err
	}
	rows, err := store.MustGameNamed("taiwan_cain_2nd").Query(`
		SELECT ui_id, it_id, hidden_option
		FROM user_items
		WHERE charac_no=?
		ORDER BY ui_id ASC
	`, characNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type rowItem struct {
		uiID         int
		itemID       int
		hiddenOption int
	}
	rawItems := []rowItem{}
	itemIDs := []int{}
	for rows.Next() {
		var item rowItem
		var itemID sql.NullInt64
		var hiddenOption sql.NullInt64
		if err := rows.Scan(&item.uiID, &itemID, &hiddenOption); err != nil {
			return nil, err
		}
		if itemID.Valid {
			item.itemID = int(itemID.Int64)
		}
		if hiddenOption.Valid {
			item.hiddenOption = int(hiddenOption.Int64)
		}
		rawItems = append(rawItems, item)
		itemIDs = append(itemIDs, item.itemID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	names, err := pvf.ItemNameMap(store, itemIDs)
	if err != nil {
		return nil, err
	}
	items := []AvatarItem{}
	for _, item := range rawItems {
		hiddenName, err := AvatarHiddenName(store, item.hiddenOption)
		if err != nil {
			return nil, err
		}
		items = append(items, AvatarItem{
			UIID:         item.uiID,
			ItemID:       item.itemID,
			ItemName:     names[item.itemID],
			HiddenOption: item.hiddenOption,
			HiddenName:   hiddenName,
		})
	}
	return map[string]any{
		"character":  character,
		"items":      items,
		"item_count": len(items),
	}, nil
}

func SetAvatarHidden(store *db.Store, characNo int, uiIDs []int, hiddenOption int) (map[string]any, error) {
	if len(uiIDs) == 0 {
		return nil, apperror.BadRequest("No avatar selected")
	}
	options, err := AvatarHiddenOptions(store)
	if err != nil {
		return nil, err
	}
	validOptions := map[int]struct{}{}
	for _, option := range options {
		validOptions[option.HiddenOption] = struct{}{}
	}
	if len(validOptions) > 0 {
		if _, ok := validOptions[hiddenOption]; !ok {
			return nil, apperror.BadRequest("Invalid hidden option")
		}
	}
	uniqueUIIDs := map[int]struct{}{}
	for _, uiID := range uiIDs {
		uniqueUIIDs[uiID] = struct{}{}
	}
	ids := sortedKeys(uniqueUIIDs)
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, characNo)
	for _, uiID := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, uiID)
	}

	game2 := store.MustGameNamed("taiwan_cain_2nd")
	tx, err := game2.Begin()
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	rows, err := tx.Query(
		"SELECT ui_id FROM user_items WHERE charac_no=? AND ui_id IN ("+strings.Join(placeholders, ",")+")",
		args...,
	)
	if err != nil {
		return nil, err
	}
	matched := map[int]struct{}{}
	for rows.Next() {
		var uiID int
		if err := rows.Scan(&uiID); err != nil {
			rows.Close()
			return nil, err
		}
		matched[uiID] = struct{}{}
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(matched) != len(uniqueUIIDs) {
		return nil, apperror.BadRequest("Selected avatar does not belong to character")
	}
	updateArgs := make([]any, 0, len(ids)+2)
	updateArgs = append(updateArgs, hiddenOption, characNo)
	for _, uiID := range ids {
		updateArgs = append(updateArgs, uiID)
	}
	if _, err := tx.Exec(
		"UPDATE user_items SET hidden_option=? WHERE charac_no=? AND ui_id IN ("+strings.Join(placeholders, ",")+")",
		updateArgs...,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	hiddenName, err := AvatarHiddenName(store, hiddenOption)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"updated":       len(matched),
		"hidden_option": hiddenOption,
		"hidden_name":   hiddenName,
	}, nil
}

func sendMailToCharacterTx(store *db.Store, tx *sql.Tx, characNo int, sender string, payload MailPayload) (int64, int, error) {
	if payload.ItemID <= 0 && payload.Gold <= 0 && payload.Message == "" {
		return 0, 0, apperror.BadRequest("Message, item, or gold is required")
	}
	letterID, err := insertMailMessage(tx, characNo, sender, payload.Message)
	if err != nil {
		return 0, 0, err
	}
	postalCount := 0
	if payload.ItemID > 0 || payload.Gold > 0 {
		postalCount, err = insertMailPostal(store, tx, characNo, letterID, sender, payload)
		if err != nil {
			return 0, 0, err
		}
	}
	return letterID, postalCount, nil
}

func insertMailMessage(tx *sql.Tx, characNo int, sender string, message string) (int64, error) {
	result, err := tx.Exec(`
		INSERT INTO letter
			(charac_no, send_charac_no, send_charac_name, letter_text, reg_date, stat)
		VALUES(?, 0, ?, ?, NOW(), 1)
	`, characNo, []byte(sender), []byte(message))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func insertMailPostal(store *db.Store, tx *sql.Tx, characNo int, letterID int64, sender string, payload MailPayload) (int, error) {
	if payload.ItemID <= 0 && payload.Gold <= 0 {
		return 0, nil
	}
	isEquipment := isEquipmentMailItem(payload)
	totalCount := payload.ItemCount
	if payload.ItemID > 0 && !isEquipment && totalCount <= 0 {
		return 0, apperror.BadRequest("Item count must be greater than 0")
	}
	if payload.ItemID <= 0 || isEquipment {
		totalCount = 1
	}
	stackLimit := 1
	if !isEquipment && payload.ItemID > 0 {
		limit, err := pvf.StackLimit(store, payload.ItemID)
		if err != nil {
			return 0, err
		}
		stackLimit = limit
	}
	sentCount := 0
	postalCount := 0
	currentLetterID := letterID
	currentGold := payload.Gold
	for sentCount < totalCount {
		if postalCount > 0 && postalCount%10 == 0 {
			var err error
			currentLetterID, err = insertMailMessage(tx, characNo, sender, payload.Message)
			if err != nil {
				return 0, err
			}
		}
		currentCount := min(stackLimit, totalCount-sentCount)
		addInfo := currentCount
		if isEquipment {
			addInfo = 1
		}
		amplifyOption := 0
		amplifyValue := 0
		forgeLevel := 0
		enhancementLevel := 0
		if isEquipment {
			amplifyOption = payload.AmplifyOption
			amplifyValue = payload.AmplifyValue
			forgeLevel = payload.ForgeLevel
			enhancementLevel = payload.EnhancementLevel
		}
		if _, err := tx.Exec(`
			INSERT INTO postal (
				occ_time, send_charac_name, receive_charac_no,
				amplify_option, amplify_value, seperate_upgrade, seal_flag,
				item_id, add_info, upgrade, gold, letter_id,
				avata_flag, creature_flag, endurance, unlimit_flag
			)
			VALUES(NOW(), ?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?, 0, 0, 0, 1)
		`,
			[]byte(sender),
			characNo,
			amplifyOption,
			amplifyValue,
			forgeLevel,
			payload.ItemID,
			addInfo,
			enhancementLevel,
			currentGold,
			currentLetterID,
		); err != nil {
			return 0, err
		}
		currentGold = 0
		sentCount += currentCount
		postalCount++
	}
	return postalCount, nil
}

func normalizeMailPayload(payload *MailPayload) {
	payload.Message = strings.TrimSpace(payload.Message)
	payload.ItemType = strings.TrimSpace(payload.ItemType)
}

func validateMailItemCount(payload MailPayload) error {
	if payload.ItemID <= 0 {
		return nil
	}
	if isEquipmentMailItem(payload) {
		return nil
	}
	if payload.ItemCount <= 0 {
		return apperror.BadRequest("Item count must be greater than 0")
	}
	return nil
}

func isEquipmentMailItem(payload MailPayload) bool {
	return strings.ToLower(strings.TrimSpace(payload.ItemType)) == "equipment"
}

func loadInventoryBlob(store *db.Store, characNo int, scopeKey string) (inventoryTarget, error) {
	scopeKey = strings.TrimSpace(scopeKey)
	if scopeKey == "" {
		scopeKey = "inventory"
	}
	scope, ok := inventoryScopes[scopeKey]
	if !ok {
		return inventoryTarget{}, apperror.BadRequest("Invalid inventory scope")
	}
	character, err := CharacterByNo(store, characNo)
	if err != nil {
		return inventoryTarget{}, err
	}
	targetID := characNo
	if scope.where == "m_id" {
		targetID = character.UID
	}
	var blob []byte
	query := fmt.Sprintf(
		"SELECT `%s` AS item_blob FROM `%s` WHERE `%s`=?",
		scope.column,
		scope.table,
		scope.where,
	)
	err = store.MustGameNamed(scope.database).QueryRow(query, targetID).Scan(&blob)
	if err == sql.ErrNoRows {
		return inventoryTarget{}, apperror.New(404, "Inventory row not found")
	}
	if err != nil {
		return inventoryTarget{}, err
	}
	return inventoryTarget{
		character: character,
		scopeKey:  scopeKey,
		scope:     scope,
		targetID:  targetID,
		blob:      blob,
	}, nil
}

func writeInventoryBlob(store *db.Store, scope inventoryScope, targetID int, blob []byte) error {
	query := fmt.Sprintf(
		"UPDATE `%s` SET `%s`=? WHERE `%s`=?",
		scope.table,
		scope.column,
		scope.where,
	)
	_, err := store.MustGameNamed(scope.database).Exec(query, blob, targetID)
	return err
}

func unpackInventoryBlob(store *db.Store, blob []byte) ([]InventoryItem, error) {
	if len(blob) == 0 {
		return []InventoryItem{}, nil
	}
	itemBytes, err := decompressInventoryItems(blob)
	if err != nil {
		return nil, err
	}
	items := []InventoryItem{}
	itemIDs := []int{}
	slotCount := len(itemBytes) / 61
	for slot := 0; slot < slotCount; slot++ {
		item := parseInventorySlot(slot, itemBytes[slot*61:(slot+1)*61])
		if item == nil {
			continue
		}
		items = append(items, *item)
		itemIDs = append(itemIDs, item.ItemID)
	}
	names, err := pvf.ItemNameMap(store, itemIDs)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].ItemName = names[items[index].ItemID]
	}
	return items, nil
}

func inventorySlotCount(blob []byte) (int, error) {
	if len(blob) == 0 {
		return 0, nil
	}
	itemBytes, err := decompressInventoryItems(blob)
	if err != nil {
		return 0, err
	}
	return len(itemBytes) / 61, nil
}

func buildDeletedInventoryBlob(deleteSlots []int, originBlob []byte) ([]byte, []int, error) {
	if len(originBlob) == 0 {
		return nil, nil, apperror.New(404, "Inventory blob is empty")
	}
	itemBytes, err := decompressInventoryItems(originBlob)
	if err != nil {
		return nil, nil, err
	}
	slotCount := len(itemBytes) / 61
	deletedSlots := []int{}
	seen := map[int]struct{}{}
	for _, slot := range deleteSlots {
		if slot < 0 || slot >= slotCount {
			return nil, nil, apperror.BadRequest("Slot is outside inventory range")
		}
		if _, ok := seen[slot]; ok {
			continue
		}
		seen[slot] = struct{}{}
		deletedSlots = append(deletedSlots, slot)
		for offset := slot * 61; offset < slot*61+61; offset++ {
			itemBytes[offset] = 0
		}
	}
	packed, err := compressInventoryItems(itemBytes)
	if err != nil {
		return nil, nil, err
	}
	prefixLen := 4
	if len(originBlob) < prefixLen {
		prefixLen = len(originBlob)
	}
	result := append([]byte{}, originBlob[:prefixLen]...)
	result = append(result, packed...)
	return result, deletedSlots, nil
}

func decompressInventoryItems(blob []byte) ([]byte, error) {
	if len(blob) <= 4 {
		return nil, apperror.New(500, "Inventory blob decompress failed")
	}
	reader, err := zlib.NewReader(bytes.NewReader(blob[4:]))
	if err != nil {
		return nil, apperror.New(500, "Inventory blob decompress failed")
	}
	defer reader.Close()
	payload, err := io.ReadAll(reader)
	if err != nil {
		return nil, apperror.New(500, "Inventory blob decompress failed")
	}
	return payload, nil
}

func compressInventoryItems(itemBytes []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := zlib.NewWriter(&buffer)
	if _, err := writer.Write(itemBytes); err != nil {
		writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func parseInventorySlot(slot int, itemBytes []byte) *InventoryItem {
	if len(itemBytes) < 61 {
		itemBytes = make([]byte, 61)
	}
	itemType := itemBytes[1]
	itemID := int(binary.LittleEndian.Uint32(itemBytes[2:6]))
	if itemType == 0 && itemID == 0 {
		return nil
	}
	itemTypeName, ok := itemTypeNames[itemType]
	if !ok {
		itemTypeName = "未知"
	}
	var countOrGrade int
	if itemTypeName == "装备" {
		countOrGrade = int(binary.BigEndian.Uint32(itemBytes[7:11]))
	} else {
		countOrGrade = int(binary.LittleEndian.Uint32(itemBytes[7:11]))
	}
	increaseType := itemBytes[17]
	increaseTypeName, ok := increaseTypeNames[increaseType]
	if !ok {
		increaseTypeName = fmt.Sprint(increaseType)
	}
	return &InventoryItem{
		Slot:             slot,
		ItemID:           itemID,
		ItemName:         "",
		ItemType:         int(itemType),
		ItemTypeName:     itemTypeName,
		IsSeal:           int(itemBytes[0]),
		EnhancementLevel: int(itemBytes[6] & 0x1F),
		CountOrGrade:     countOrGrade,
		Durability:       int(binary.LittleEndian.Uint16(itemBytes[11:13])),
		Orb:              int(binary.LittleEndian.Uint32(itemBytes[13:17])),
		IncreaseType:     int(increaseType),
		IncreaseTypeName: increaseTypeName,
		IncreaseValue:    int(binary.LittleEndian.Uint16(itemBytes[18:20])),
		ForgeLevel:       int(itemBytes[51]),
	}
}

func scanCharacter(rows scanner, jobMap pvf.JobMap) (Character, error) {
	var item Character
	var characNameHex sql.NullString
	if err := rows.Scan(
		&item.UID,
		&item.CharacNo,
		&characNameHex,
		&item.Level,
		&item.Job,
		&item.GrowType,
		&item.DeleteFlag,
		&item.ExpertJob,
		&item.PVPGrade,
		&item.PVPWin,
		&item.PVPPoint,
		&item.WinPoint,
	); err != nil {
		return Character{}, err
	}
	item.CharacName = decodeNameHex(characNameHex.String)
	item.GrowTypeBase = item.GrowType % 16
	item.WakeFlag = item.GrowType / 16
	item.JobName = jobName(item.Job, item.GrowType, jobMap)
	item.ExpertJobName = expertJobName(item.ExpertJob)
	item.PVPGradeName = pvpRankName(item.PVPGrade)
	return item, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func storeGameCain(store *db.Store) *sql.DB {
	return store.MustGameNamed("taiwan_cain")
}

func storeBilling(store *db.Store) *sql.DB {
	return store.MustGameNamed("taiwan_billing")
}

func ensureCeraRows(store *db.Store, uid int) error {
	billing := storeBilling(store)
	tx, err := billing.Begin()
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := ensureCeraRowsTx(tx, uid); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func ensureCeraRowsTx(tx *sql.Tx, uid int) error {
	if _, err := tx.Exec(`
		INSERT INTO cash_cera(account, cera, mod_date, reg_date)
		SELECT ?, 0, NOW(), NOW()
		FROM DUAL
		WHERE NOT EXISTS (SELECT 1 FROM cash_cera WHERE account=?)
	`, uid, uid); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO cash_cera_point(account, cera_point, mod_date, reg_date)
		SELECT ?, 0, NOW(), NOW()
		FROM DUAL
		WHERE NOT EXISTS (SELECT 1 FROM cash_cera_point WHERE account=?)
	`, uid, uid); err != nil {
		return err
	}
	return nil
}

func ceraValue(store *db.Store, uid int) (int, error) {
	var value sql.NullInt64
	err := storeBilling(store).QueryRow("SELECT cera FROM cash_cera WHERE account=?", uid).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int(value.Int64), nil
}

func ceraPointValue(store *db.Store, uid int) (int, error) {
	var value sql.NullInt64
	err := storeBilling(store).QueryRow("SELECT cera_point FROM cash_cera_point WHERE account=?", uid).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int(value.Int64), nil
}

func normalizeCeraType(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value != "cera" && value != "cera_point" {
		return "", apperror.BadRequest("Invalid cera type")
	}
	return value, nil
}

func normalizeCeraAction(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value != "add" && value != "set" {
		return "", apperror.BadRequest("Invalid cera action")
	}
	return value, nil
}

func punishTypeName(value int) string {
	switch value {
	case 1:
		return "禁止登陆"
	case 4, 11:
		return "限制交易"
	default:
		return fmt.Sprintf("%d", value)
	}
}

func decodeNameHex(value string) string {
	if value == "" {
		return ""
	}
	data, err := hex.DecodeString(value)
	if err != nil {
		return value
	}
	text := string(decodeUTF8Replace(data))
	latin1 := encodeLatin1Replace(text)
	cp1252 := encodeCP1252Replace(text)
	rebuilt := make([]byte, 0, len(latin1))
	for i := range latin1 {
		if latin1[i] == '?' && i < len(cp1252) {
			rebuilt = append(rebuilt, cp1252[i])
		} else {
			rebuilt = append(rebuilt, latin1[i])
		}
	}
	return string(decodeUTF8Replace(rebuilt))
}

func decodeUTF8Replace(data []byte) []rune {
	result := []rune{}
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			result = append(result, utf8.RuneError)
			data = data[1:]
			continue
		}
		result = append(result, r)
		data = data[size:]
	}
	return result
}

func encodeLatin1Replace(text string) []byte {
	result := make([]byte, 0, len(text))
	for _, r := range text {
		if r <= 0xFF {
			result = append(result, byte(r))
		} else {
			result = append(result, '?')
		}
	}
	return result
}

func encodeCP1252Replace(text string) []byte {
	result := make([]byte, 0, len(text))
	for _, r := range text {
		if r <= 0x7F || (r >= 0xA0 && r <= 0xFF) {
			result = append(result, byte(r))
			continue
		}
		if b, ok := cp1252Encode[r]; ok {
			result = append(result, b)
			continue
		}
		result = append(result, '?')
	}
	return result
}

func expertJobName(value int) string {
	if name, ok := expertJobMap[value]; ok {
		return name
	}
	return fmt.Sprintf("%d", value)
}

func jobName(job int, growType int, jobMap pvf.JobMap) string {
	if growMap, ok := jobMap[job]; ok {
		if name := growMap[growType%16]; name != "" {
			return name
		}
	}
	return fmt.Sprintf("%d/%d", job, growType)
}

func pvpRankName(value int) string {
	if value >= 0 && value < len(pvpRankList) {
		return pvpRankList[value]
	}
	return fmt.Sprintf("%d", value)
}

func segmentNames(suffix string, start int, end int) []string {
	result := []string{}
	if start <= end {
		for i := start; i <= end; i++ {
			result = append(result, fmt.Sprintf("%d%s", i, suffix))
		}
	} else {
		for i := start; i >= end; i-- {
			result = append(result, fmt.Sprintf("%d%s", i, suffix))
		}
	}
	return result
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, c := range value {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func parseInt(value string, fallback int) int {
	result := 0
	for _, c := range value {
		if c < '0' || c > '9' {
			return fallback
		}
		result = result*10 + int(c-'0')
	}
	return result
}

func sortedKeys[V any](items map[int]V) []int {
	keys := make([]int, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys
}

func clamp(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
