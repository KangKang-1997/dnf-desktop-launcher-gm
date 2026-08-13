package pvf

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"dnf-launcher-go/internal/apperror"
	"dnf-launcher-go/internal/db"
	appsettings "dnf-launcher-go/internal/settings"
)

type Meta struct {
	MD5            string
	PVFPath        string
	EncodeName     string
	FileSize       int64
	StackableCount int
	EquipmentCount int
	ExpLevelCount  int
	Active         int
	CreatedBy      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type JobMap map[int]map[int]string

type RefreshResult struct {
	MD5            string   `json:"md5"`
	Path           string   `json:"path"`
	FileSize       int64    `json:"file_size"`
	Encode         string   `json:"encode"`
	StackableCount int      `json:"stackable_count"`
	EquipmentCount int      `json:"equipment_count"`
	ExpLevelCount  int      `json:"exp_level_count"`
	Logs           []string `json:"logs"`
}

type RefreshTask struct {
	ID           int64          `json:"id"`
	Status       string         `json:"status"`
	PVFPath      string         `json:"pvf_path"`
	Encode       string         `json:"encode"`
	Message      string         `json:"message"`
	Result       map[string]any `json:"result"`
	Error        string         `json:"error"`
	StartedBy    int            `json:"started_by"`
	StartedAt    string         `json:"started_at"`
	FinishedAt   string         `json:"finished_at"`
	ClientPVFMD5 string         `json:"client_pvf_md5"`
}

func EnsureTables(toolDB *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS pvf_meta (
			md5 CHAR(32) PRIMARY KEY,
			pvf_path VARCHAR(512) NOT NULL,
			encode_name VARCHAR(32) NOT NULL,
			file_size BIGINT NOT NULL,
			stackable_count INT NOT NULL DEFAULT 0,
			equipment_count INT NOT NULL DEFAULT 0,
			exp_level_count INT NOT NULL DEFAULT 0,
			active TINYINT NOT NULL DEFAULT 0,
			created_by INT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
				ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS pvf_items (
			pvf_md5 CHAR(32) NOT NULL,
			item_id INT NOT NULL,
			item_type VARCHAR(32) NOT NULL,
			item_name VARCHAR(255) NOT NULL,
			detail_json LONGTEXT NULL,
			PRIMARY KEY (pvf_md5, item_id),
			INDEX idx_pvf_items_name (item_name(64)),
			INDEX idx_pvf_items_type (item_type)
		)`,
		`CREATE TABLE IF NOT EXISTS pvf_data (
			pvf_md5 CHAR(32) NOT NULL,
			data_key VARCHAR(64) NOT NULL,
			data_json LONGTEXT NOT NULL,
			PRIMARY KEY (pvf_md5, data_key)
		)`,
		`CREATE TABLE IF NOT EXISTS pvf_refresh_jobs (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			status VARCHAR(32) NOT NULL,
			pvf_path VARCHAR(512) NOT NULL,
			encode_name VARCHAR(32) NOT NULL,
			message VARCHAR(255) NOT NULL DEFAULT '',
			result_json LONGTEXT NULL,
			error_text LONGTEXT NULL,
			started_by INT NOT NULL,
			started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
				ON UPDATE CURRENT_TIMESTAMP,
			finished_at DATETIME NULL,
			INDEX idx_pvf_refresh_jobs_status (status),
			INDEX idx_pvf_refresh_jobs_started_at (started_at)
		)`,
	}
	for _, statement := range statements {
		if _, err := toolDB.Exec(statement); err != nil {
			return err
		}
	}
	if _, err := toolDB.Exec(`
		ALTER TABLE pvf_refresh_jobs
		ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
			ON UPDATE CURRENT_TIMESTAMP
	`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return err
	}
	return nil
}

func ExpireStaleRefreshJobs(store *db.Store) error {
	return expireStaleRefreshJobs(store)
}

func Active(store *db.Store) (*Meta, error) {
	row := store.Tool().QueryRow(`
		SELECT md5, pvf_path, encode_name, file_size, stackable_count,
			   equipment_count, exp_level_count, active, created_by,
			   created_at, updated_at
		FROM pvf_meta
		WHERE active=1
		ORDER BY updated_at DESC
		LIMIT 1
	`)
	var meta Meta
	err := row.Scan(
		&meta.MD5,
		&meta.PVFPath,
		&meta.EncodeName,
		&meta.FileSize,
		&meta.StackableCount,
		&meta.EquipmentCount,
		&meta.ExpLevelCount,
		&meta.Active,
		&meta.CreatedBy,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func StartRefresh(store *db.Store, pvfPath string, encodeName string, startedBy int) (RefreshTask, error) {
	if err := EnsureTables(store.Tool()); err != nil {
		return RefreshTask{}, err
	}
	encodeName, err := normalizePVFEncoding(encodeName)
	if err != nil {
		return RefreshTask{}, apperror.BadRequest(err.Error())
	}
	pvfPath = strings.TrimSpace(pvfPath)
	if pvfPath == "" {
		return RefreshTask{}, apperror.BadRequest("PVF path is required")
	}
	if err := expireStaleRefreshJobs(store); err != nil {
		return RefreshTask{}, err
	}
	running, err := runningRefreshJob(store)
	if err != nil {
		return RefreshTask{}, err
	}
	if running != nil {
		return *running, nil
	}
	result, err := store.Tool().Exec(`
		INSERT INTO pvf_refresh_jobs(status, pvf_path, encode_name, message, started_by)
		VALUES('queued', ?, ?, 'queued', ?)
	`, pvfPath, encodeName, startedBy)
	if err != nil {
		return RefreshTask{}, err
	}
	jobID, err := result.LastInsertId()
	if err != nil {
		return RefreshTask{}, err
	}
	go runRefreshJob(store, jobID, pvfPath, encodeName, startedBy)
	return RefreshJob(store, jobID)
}

func runningRefreshJob(store *db.Store) (*RefreshTask, error) {
	row := store.Tool().QueryRow(`
		SELECT id
		FROM pvf_refresh_jobs
		WHERE status IN ('queued', 'running')
		ORDER BY id DESC
		LIMIT 1
	`)
	var id int64
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	task, err := RefreshJob(store, id)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func RefreshJob(store *db.Store, jobID int64) (RefreshTask, error) {
	row := store.Tool().QueryRow(`
		SELECT id, status, pvf_path, encode_name, message, result_json, error_text,
			   started_by, started_at, finished_at
		FROM pvf_refresh_jobs
		WHERE id=?
	`, jobID)
	var task RefreshTask
	var resultJSON sql.NullString
	var errorText sql.NullString
	var startedAt time.Time
	var finishedAt sql.NullTime
	if err := row.Scan(
		&task.ID,
		&task.Status,
		&task.PVFPath,
		&task.Encode,
		&task.Message,
		&resultJSON,
		&errorText,
		&task.StartedBy,
		&startedAt,
		&finishedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return RefreshTask{}, apperror.New(404, "PVF refresh job not found")
		}
		return RefreshTask{}, err
	}
	task.StartedAt = startedAt.Format("2006-01-02 15:04:05")
	if finishedAt.Valid {
		task.FinishedAt = finishedAt.Time.Format("2006-01-02 15:04:05")
	}
	if errorText.Valid {
		task.Error = errorText.String
	}
	task.Result = map[string]any{}
	if resultJSON.Valid && resultJSON.String != "" {
		_ = json.Unmarshal([]byte(resultJSON.String), &task.Result)
	}
	task.ClientPVFMD5 = normalizeMD5(appsettings.ClientPVFMD5(store))
	return task, nil
}

func expireStaleRefreshJobs(store *db.Store) error {
	_, err := store.Tool().Exec(`
		UPDATE pvf_refresh_jobs
		SET status='failed',
			message='interrupted',
			error_text='PVF refresh was interrupted before completion',
			finished_at=NOW()
		WHERE status IN ('queued', 'running')
		  AND updated_at < DATE_SUB(NOW(), INTERVAL 10 MINUTE)
	`)
	return err
}

func LatestRefreshJob(store *db.Store) (map[string]any, error) {
	_ = expireStaleRefreshJobs(store)
	row := store.Tool().QueryRow("SELECT id FROM pvf_refresh_jobs ORDER BY id DESC LIMIT 1")
	var id int64
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	task, err := RefreshJob(store, id)
	if err != nil {
		return nil, err
	}
	return TaskPayload(task), nil
}

func TaskPayload(task RefreshTask) map[string]any {
	payload := map[string]any{
		"id":             task.ID,
		"status":         task.Status,
		"pvf_path":       task.PVFPath,
		"encode":         task.Encode,
		"message":        task.Message,
		"result":         task.Result,
		"error":          task.Error,
		"started_by":     task.StartedBy,
		"started_at":     task.StartedAt,
		"finished_at":    task.FinishedAt,
		"client_pvf_md5": task.ClientPVFMD5,
	}
	if task.Status == "completed" {
		for key, value := range task.Result {
			payload[key] = value
		}
		payload["loaded"] = true
	}
	return payload
}

func Refresh(store *db.Store, pvfPath string, encodeName string, createdBy int) (RefreshResult, error) {
	if err := EnsureTables(store.Tool()); err != nil {
		return RefreshResult{}, err
	}
	ctx := context.Background()
	tx, err := store.Tool().BeginTx(ctx, nil)
	if err != nil {
		return RefreshResult{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	var acquired int
	if err := tx.QueryRowContext(ctx, "SELECT GET_LOCK(?, 30) AS acquired", "launcher:pvf_refresh").Scan(&acquired); err != nil {
		return RefreshResult{}, err
	}
	defer tx.ExecContext(ctx, "SELECT RELEASE_LOCK(?)", "launcher:pvf_refresh")
	if acquired != 1 {
		return RefreshResult{}, apperror.Internal("PVF cache writer is busy, try again later")
	}
	cache, err := buildCache(pvfPath, encodeName)
	if err != nil {
		return RefreshResult{}, err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM pvf_items WHERE pvf_md5=?", cache.MD5); err != nil {
		return RefreshResult{}, err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM pvf_data WHERE pvf_md5=?", cache.MD5); err != nil {
		return RefreshResult{}, err
	}
	if err := insertItems(ctx, tx, cache); err != nil {
		return RefreshResult{}, err
	}
	dataStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO pvf_data(pvf_md5, data_key, data_json)
		VALUES(?, ?, ?)
	`)
	if err != nil {
		return RefreshResult{}, err
	}
	for key, value := range cache.Data {
		payload, err := json.Marshal(value)
		if err != nil {
			dataStmt.Close()
			return RefreshResult{}, err
		}
		if _, err := dataStmt.ExecContext(ctx, cache.MD5, key, string(payload)); err != nil {
			dataStmt.Close()
			return RefreshResult{}, err
		}
	}
	if err := dataStmt.Close(); err != nil {
		return RefreshResult{}, err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE pvf_meta SET active=0"); err != nil {
		return RefreshResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO pvf_meta(
			md5, pvf_path, encode_name, file_size, stackable_count,
			equipment_count, exp_level_count, active, created_by
		)
		VALUES(?, ?, ?, ?, ?, ?, ?, 1, ?)
		ON DUPLICATE KEY UPDATE
			pvf_path=VALUES(pvf_path),
			encode_name=VALUES(encode_name),
			file_size=VALUES(file_size),
			stackable_count=VALUES(stackable_count),
			equipment_count=VALUES(equipment_count),
			exp_level_count=VALUES(exp_level_count),
			active=1,
			created_by=VALUES(created_by)
	`, cache.MD5, cache.Path, cache.Encode, cache.FileSize, cache.StackableCount, cache.EquipmentCount, cache.ExpLevelCount, createdBy); err != nil {
		return RefreshResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return RefreshResult{}, err
	}
	committed = true
	logs := cache.Logs
	if len(logs) > 50 {
		logs = logs[len(logs)-50:]
	}
	return RefreshResult{
		MD5:            cache.MD5,
		Path:           cache.Path,
		FileSize:       cache.FileSize,
		Encode:         cache.Encode,
		StackableCount: cache.StackableCount,
		EquipmentCount: cache.EquipmentCount,
		ExpLevelCount:  cache.ExpLevelCount,
		Logs:           logs,
	}, nil
}

func insertItems(ctx context.Context, tx *sql.Tx, cache cacheBuild) error {
	const chunkSize = 500
	for start := 0; start < len(cache.Items); start += chunkSize {
		end := start + chunkSize
		if end > len(cache.Items) {
			end = len(cache.Items)
		}
		chunk := cache.Items[start:end]
		placeholders := make([]string, 0, len(chunk))
		args := make([]any, 0, len(chunk)*5)
		for _, item := range chunk {
			detail, err := detailJSON(item.Detail)
			if err != nil {
				return err
			}
			placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
			args = append(args, cache.MD5, item.ID, item.Type, item.Name, detail)
		}
		query := `
			INSERT INTO pvf_items
				(pvf_md5, item_id, item_type, item_name, detail_json)
			VALUES ` + strings.Join(placeholders, ",")
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}

func runRefreshJob(store *db.Store, jobID int64, pvfPath string, encodeName string, startedBy int) {
	updateJob(store, jobID, "running", "reading PVF file", nil, "")
	done := make(chan struct{})
	go refreshHeartbeat(store, jobID, done)
	result, err := Refresh(store, pvfPath, encodeName, startedBy)
	close(done)
	if err != nil {
		updateJob(store, jobID, "failed", "failed", nil, err.Error())
		return
	}
	payload := map[string]any{
		"md5":             result.MD5,
		"path":            result.Path,
		"file_size":       result.FileSize,
		"encode":          result.Encode,
		"stackable_count": result.StackableCount,
		"equipment_count": result.EquipmentCount,
		"exp_level_count": result.ExpLevelCount,
		"logs":            result.Logs,
		"client_pvf_md5":  normalizeMD5(appsettings.ClientPVFMD5(store)),
	}
	updateJob(store, jobID, "completed", "completed", payload, "")
}

func refreshHeartbeat(store *db.Store, jobID int64, done <-chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			touchJob(store, jobID)
		case <-done:
			return
		}
	}
}

func touchJob(store *db.Store, jobID int64) {
	_, _ = store.Tool().Exec(`
		UPDATE pvf_refresh_jobs
		SET updated_at=NOW()
		WHERE id=? AND status='running'
	`, jobID)
}

func updateJob(store *db.Store, jobID int64, status string, message string, result map[string]any, errorText string) {
	var resultJSON any
	if result != nil {
		payload, err := json.Marshal(result)
		if err == nil {
			resultJSON = string(payload)
		}
	}
	if status == "completed" || status == "failed" {
		_, _ = store.Tool().Exec(`
			UPDATE pvf_refresh_jobs
			SET status=?, message=?, result_json=?, error_text=?, finished_at=NOW()
			WHERE id=?
		`, status, message, resultJSON, errorText, jobID)
		return
	}
	_, _ = store.Tool().Exec(`
		UPDATE pvf_refresh_jobs
		SET status=?, message=?
		WHERE id=?
	`, status, message, jobID)
}

func Status(store *db.Store) (map[string]any, error) {
	active, err := Active(store)
	if err != nil {
		return nil, err
	}
	clientMD5 := normalizeMD5(appsettings.ClientPVFMD5(store))
	if active == nil {
		payload := map[string]any{
			"loaded":         false,
			"client_pvf_md5": clientMD5,
		}
		if job, err := LatestRefreshJob(store); err == nil && job != nil {
			payload["refresh_job"] = job
		}
		return payload, nil
	}
	payload := map[string]any{
		"loaded":          true,
		"md5":             active.MD5,
		"pvf_path":        active.PVFPath,
		"encode":          active.EncodeName,
		"file_size":       active.FileSize,
		"stackable_count": active.StackableCount,
		"equipment_count": active.EquipmentCount,
		"exp_level_count": active.ExpLevelCount,
		"client_pvf_md5":  clientMD5,
		"updated_at":      active.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if job, err := LatestRefreshJob(store); err == nil && job != nil {
		payload["refresh_job"] = job
	}
	return payload, nil
}

func Data(store *db.Store, dataKey string, target any) (bool, error) {
	active, err := Active(store)
	if err != nil {
		return false, err
	}
	if active == nil {
		return false, nil
	}
	var payload string
	err = store.Tool().QueryRow(
		"SELECT data_json FROM pvf_data WHERE pvf_md5=? AND data_key=?",
		active.MD5,
		dataKey,
	).Scan(&payload)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(payload), target); err != nil {
		return false, err
	}
	return true, nil
}

func GetExpForLevel(store *db.Store, level int) (int, error) {
	var expTable []any
	ok, err := Data(store, "exp_table", &expTable)
	if err != nil {
		return 0, err
	}
	if !ok || len(expTable) == 0 {
		return 0, apperror.Internal("PVF exp table is not loaded")
	}
	index := level - 2
	if index < -1 || index >= len(expTable) {
		return 0, apperror.BadRequest("Level is outside PVF exp table range")
	}
	if level <= 1 {
		return 1, nil
	}
	value, ok := intFromJSON(expTable[index])
	if !ok {
		return 0, apperror.Internal("PVF exp table contains invalid value")
	}
	return value + 1, nil
}

func SearchItems(store *db.Store, keyword string, itemID *int, itemType string, limit int, page int) (map[string]any, error) {
	active, err := Active(store)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return map[string]any{"items": []map[string]any{}, "page": 1, "limit": limit, "total": 0}, nil
	}
	limit = clamp(limit, 1, 100)
	page = clamp(page, 1, 2147483647)
	offset := (page - 1) * limit

	clauses := []string{"pvf_md5=?"}
	args := []any{active.MD5}
	if itemID != nil {
		clauses = append(clauses, "item_id=?")
		args = append(args, *itemID)
	}
	itemType = strings.TrimSpace(itemType)
	if itemType != "" {
		clauses = append(clauses, "item_type=?")
		args = append(args, itemType)
	}
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		clauses = append(clauses, "(item_name LIKE ? OR CAST(item_id AS CHAR)=?)")
		args = append(args, "%"+keyword+"%", keyword)
	}
	where := strings.Join(clauses, " AND ")

	var total int
	if err := store.Tool().QueryRow("SELECT COUNT(*) AS total FROM pvf_items WHERE "+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := store.Tool().Query(`
		SELECT item_id, item_type, item_name, detail_json
		FROM pvf_items
		WHERE `+where+`
		ORDER BY CHAR_LENGTH(item_name), item_id
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []map[string]any{}
	for rows.Next() {
		var id int
		var typ string
		var name string
		var detailJSON sql.NullString
		if err := rows.Scan(&id, &typ, &name, &detailJSON); err != nil {
			return nil, err
		}
		var stackLimit any
		if typ == "stackable" {
			stackLimit = stackLimitFromDetail(detailJSON.String)
		}
		items = append(items, map[string]any{
			"item_id":     id,
			"item_type":   typ,
			"item_name":   name,
			"stack_limit": stackLimit,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return map[string]any{
		"items": items,
		"page":  page,
		"limit": limit,
		"total": total,
	}, nil
}

func ItemDetail(store *db.Store, itemID int) (map[string]any, error) {
	active, err := Active(store)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return map[string]any{}, nil
	}
	var detailJSON sql.NullString
	err = store.Tool().QueryRow(
		"SELECT detail_json FROM pvf_items WHERE pvf_md5=? AND item_id=?",
		active.MD5,
		itemID,
	).Scan(&detailJSON)
	if err == sql.ErrNoRows {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	if !detailJSON.Valid || detailJSON.String == "" {
		return map[string]any{}, nil
	}
	var detail map[string]any
	if err := json.Unmarshal([]byte(detailJSON.String), &detail); err != nil {
		return nil, err
	}
	return detail, nil
}

func StackLimit(store *db.Store, itemID int) (int, error) {
	detail, err := ItemDetail(store, itemID)
	if err != nil {
		return 0, err
	}
	stackLimit, ok := detail["[stack limit]"].([]any)
	if !ok || len(stackLimit) == 0 {
		return 2147483647, nil
	}
	value, ok := intFromJSON(stackLimit[0])
	if !ok {
		return 2147483647, nil
	}
	if value < 1 {
		return 1, nil
	}
	return value, nil
}

func ItemNameMap(store *db.Store, itemIDs []int) (map[int]string, error) {
	unique := map[int]struct{}{}
	for _, itemID := range itemIDs {
		if itemID > 0 {
			unique[itemID] = struct{}{}
		}
	}
	if len(unique) == 0 {
		return map[int]string{}, nil
	}
	active, err := Active(store)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return map[int]string{}, nil
	}
	ids := sortedKeys(unique)
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, active.MD5)
	for _, itemID := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, itemID)
	}
	rows, err := store.Tool().Query(`
		SELECT item_id, item_name
		FROM pvf_items
		WHERE pvf_md5=? AND item_id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[int]string{}
	for rows.Next() {
		var itemID int
		var itemName string
		if err := rows.Scan(&itemID, &itemName); err != nil {
			return nil, err
		}
		result[itemID] = itemName
	}
	return result, rows.Err()
}

func Jobs(store *db.Store) (JobMap, error) {
	var raw map[string]map[string]any
	ok, err := Data(store, "job", &raw)
	if err != nil {
		return nil, err
	}
	result := JobMap{}
	if !ok {
		return result, nil
	}
	for jobKey, growMap := range raw {
		jobID, ok := parseInt(jobKey)
		if !ok {
			continue
		}
		result[jobID] = map[int]string{}
		for growKey, growName := range growMap {
			growID, ok := parseInt(growKey)
			if !ok {
				continue
			}
			result[jobID][growID] = fmt.Sprint(growName)
		}
	}
	return result, nil
}

func JobOptions(store *db.Store, expertJobName func(int) string, pvpRankNames []string) (map[string]any, error) {
	jobs, err := Jobs(store)
	if err != nil {
		return nil, err
	}
	jobIDs := sortedKeys(jobs)
	jobRows := []map[string]any{}
	for _, jobID := range jobIDs {
		growMap := jobs[jobID]
		growIDs := sortedKeys(growMap)
		growRows := []map[string]any{}
		for _, growID := range growIDs {
			growRows = append(growRows, map[string]any{
				"grow_type": growID,
				"name":      growMap[growID],
			})
		}
		name := growMap[0]
		if name == "" {
			name = fmt.Sprintf("职业 %d", jobID)
		}
		jobRows = append(jobRows, map[string]any{
			"job":        jobID,
			"name":       name,
			"grow_types": growRows,
		})
	}
	expertJobs := []map[string]any{}
	for i := 0; i <= 4; i++ {
		expertJobs = append(expertJobs, map[string]any{
			"expert_job": i,
			"name":       expertJobName(i),
		})
	}
	pvpRanks := []map[string]any{}
	for i, name := range pvpRankNames {
		pvpRanks = append(pvpRanks, map[string]any{
			"pvp_grade": i,
			"name":      name,
		})
	}
	return map[string]any{
		"jobs":        jobRows,
		"expert_jobs": expertJobs,
		"pvp_ranks":   pvpRanks,
	}, nil
}

func normalizeMD5(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func stackLimitFromDetail(payload string) any {
	if payload == "" {
		return nil
	}
	var detail map[string]any
	if err := json.Unmarshal([]byte(payload), &detail); err != nil {
		return nil
	}
	stackLimit, ok := detail["[stack limit]"].([]any)
	if !ok || len(stackLimit) == 0 {
		return nil
	}
	value, ok := intFromJSON(stackLimit[0])
	if !ok || value <= 0 {
		return nil
	}
	return value
}

func intFromJSON(value any) (int, bool) {
	switch v := value.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case string:
		return parseInt(v)
	default:
		return 0, false
	}
}

func parseInt(value string) (int, bool) {
	if value == "" {
		return 0, false
	}
	result := 0
	for _, c := range value {
		if c < '0' || c > '9' {
			return 0, false
		}
		result = result*10 + int(c-'0')
	}
	return result, true
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
