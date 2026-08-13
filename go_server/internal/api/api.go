package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dnf-launcher-go/internal/accounts"
	"dnf-launcher-go/internal/apperror"
	"dnf-launcher-go/internal/audit"
	"dnf-launcher-go/internal/config"
	"dnf-launcher-go/internal/db"
	"dnf-launcher-go/internal/gm"
	"dnf-launcher-go/internal/launcher"
	"dnf-launcher-go/internal/permissions"
	"dnf-launcher-go/internal/pvf"
	"dnf-launcher-go/internal/security"
	appsettings "dnf-launcher-go/internal/settings"
)

type Handler struct {
	settings config.Settings
	store    *db.Store
}

type response map[string]any

type loginRequest struct {
	AccountName string `json:"account_name"`
	Password    string `json:"password"`
}

type registerRequest struct {
	AccountName string `json:"account_name"`
	Password    string `json:"password"`
	QQ          string `json:"qq"`
}

type changePasswordRequest struct {
	AccountName string `json:"account_name"`
	QQ          string `json:"qq"`
	NewPassword string `json:"new_password"`
}

type adminChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type accountResolveRequest struct {
	UID         *int   `json:"uid"`
	AccountName string `json:"account_name"`
}

type ceraQueryRequest struct {
	UID         *int   `json:"uid"`
	AccountName string `json:"account_name"`
}

type ceraChargeRequest struct {
	UID         *int   `json:"uid"`
	AccountName string `json:"account_name"`
	CeraType    string `json:"cera_type"`
	Action      string `json:"action"`
	Amount      int    `json:"amount"`
}

type banQueryRequest struct {
	UID         *int   `json:"uid"`
	AccountName string `json:"account_name"`
}

type banSetRequest struct {
	UID         *int   `json:"uid"`
	AccountName string `json:"account_name"`
	PunishType  int    `json:"punish_type"`
	Days        int    `json:"days"`
	Reason      string `json:"reason"`
}

type eventAddRequest struct {
	EventID    int  `json:"event_id"`
	Parameter1 *int `json:"parameter1"`
	Parameter2 *int `json:"parameter2"`
}

type mailSendRequest struct {
	CharacNo         int    `json:"charac_no"`
	Message          string `json:"message"`
	ItemID           *int   `json:"item_id"`
	ItemCount        int    `json:"item_count"`
	Gold             int    `json:"gold"`
	ItemType         string `json:"item_type"`
	ItemGrade        int    `json:"item_grade"`
	EnhancementLevel int    `json:"enhancement_level"`
	ForgeLevel       int    `json:"forge_level"`
	AmplifyOption    int    `json:"amplify_option"`
	AmplifyValue     int    `json:"amplify_value"`
}

type mailMassSendRequest struct {
	Message          string `json:"message"`
	ItemID           *int   `json:"item_id"`
	ItemCount        int    `json:"item_count"`
	Gold             int    `json:"gold"`
	ItemType         string `json:"item_type"`
	ItemGrade        int    `json:"item_grade"`
	EnhancementLevel int    `json:"enhancement_level"`
	ForgeLevel       int    `json:"forge_level"`
	AmplifyOption    int    `json:"amplify_option"`
	AmplifyValue     int    `json:"amplify_value"`
}

type mailDeleteRequest struct {
	CharacNo int `json:"charac_no"`
}

type characterLevelRequest struct {
	CharacNo int `json:"charac_no"`
	Level    int `json:"level"`
}

type characterJobRequest struct {
	CharacNo  int `json:"charac_no"`
	Job       int `json:"job"`
	GrowType  int `json:"grow_type"`
	WakeFlag  int `json:"wake_flag"`
	ExpertJob int `json:"expert_job"`
}

type characterPVPGradeRequest struct {
	CharacNo int `json:"charac_no"`
	PVPGrade int `json:"pvp_grade"`
}

type characterPVPPointRequest struct {
	CharacNo int `json:"charac_no"`
	PVPPoint int `json:"pvp_point"`
}

type characterVisibilityRequest struct {
	CharacNo int `json:"charac_no"`
}

type inventoryQueryRequest struct {
	CharacNo int    `json:"charac_no"`
	Scope    string `json:"scope"`
}

type inventoryDeleteRequest struct {
	CharacNo int    `json:"charac_no"`
	Scope    string `json:"scope"`
	Slot     int    `json:"slot"`
}

type inventoryClearRequest struct {
	CharacNo int    `json:"charac_no"`
	Scope    string `json:"scope"`
}

type avatarQueryRequest struct {
	CharacNo int `json:"charac_no"`
}

type avatarHiddenRequest struct {
	CharacNo     int   `json:"charac_no"`
	UIIDs        []int `json:"ui_ids"`
	HiddenOption int   `json:"hidden_option"`
}

type setPermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

type homeSettingsRequest struct {
	HomeTitle         string                     `json:"home_title"`
	HomeEyebrow       string                     `json:"home_eyebrow"`
	ClientDownloadURL string                     `json:"client_download_url"`
	Announcements     []appsettings.Announcement `json:"announcements"`
}

type pvfClientMD5Request struct {
	ClientPVFMD5 string `json:"client_pvf_md5"`
}

type pvfRefreshRequest struct {
	PVFPath string `json:"pvf_path"`
	Encode  string `json:"encode"`
}

func RegisterRoutes(mux *http.ServeMux, settings config.Settings, store *db.Store) {
	h := Handler{settings: settings, store: store}

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /api/settings", h.settingsResponse)
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.HandleFunc("POST /api/auth/register", h.register)
	mux.HandleFunc("POST /api/auth/change-password", h.changePassword)
	mux.HandleFunc("POST /api/auth/admin/change-password", h.adminChangePassword)
	mux.HandleFunc("GET /api/pvf/status", h.pvfStatus)
	mux.HandleFunc("GET /api/pvf/items", h.pvfItems)
	mux.HandleFunc("GET /api/gm/characters", h.gmCharacters)
	mux.HandleFunc("POST /api/gm/account/resolve", h.resolveGMAccount)
	mux.HandleFunc("POST /api/gm/cera/query", h.queryCera)
	mux.HandleFunc("POST /api/gm/cera/charge", h.chargeCera)
	mux.HandleFunc("POST /api/gm/ban/query", h.queryBan)
	mux.HandleFunc("POST /api/gm/ban/set", h.setBan)
	mux.HandleFunc("POST /api/gm/ban/unban", h.unban)
	mux.HandleFunc("GET /api/gm/events", h.listEvents)
	mux.HandleFunc("POST /api/gm/events", h.addEvent)
	mux.HandleFunc("DELETE /api/gm/events/{log_id}", h.deleteEvent)
	mux.HandleFunc("POST /api/gm/mail/send", h.sendMail)
	mux.HandleFunc("POST /api/gm/mail/send-all", h.sendMailAll)
	mux.HandleFunc("POST /api/gm/mail/delete", h.deleteMail)
	mux.HandleFunc("POST /api/gm/mail/delete-all", h.deleteMailAll)
	mux.HandleFunc("GET /api/gm/character/job-options", h.characterJobOptions)
	mux.HandleFunc("POST /api/gm/character/level", h.setCharacterLevel)
	mux.HandleFunc("POST /api/gm/character/job", h.setCharacterJob)
	mux.HandleFunc("POST /api/gm/character/pvp-grade", h.setCharacterPVPGrade)
	mux.HandleFunc("POST /api/gm/character/pvp-point", h.setCharacterPVPPoint)
	mux.HandleFunc("POST /api/gm/character/delete", h.deleteCharacter)
	mux.HandleFunc("POST /api/gm/character/recover", h.recoverCharacter)
	mux.HandleFunc("POST /api/gm/inventory/query", h.queryInventory)
	mux.HandleFunc("POST /api/gm/inventory/delete", h.deleteInventoryItem)
	mux.HandleFunc("POST /api/gm/inventory/clear", h.clearInventoryScope)
	mux.HandleFunc("GET /api/gm/avatar/options", h.avatarOptions)
	mux.HandleFunc("POST /api/gm/avatar/query", h.queryAvatar)
	mux.HandleFunc("POST /api/gm/avatar/hidden", h.setAvatarHidden)
	mux.HandleFunc("POST /api/launcher/direct", h.directLaunch)
	mux.HandleFunc("GET /api/admin/permissions", h.listPermissions)
	mux.HandleFunc("GET /api/admin/accounts", h.listAccounts)
	mux.HandleFunc("PUT /api/admin/accounts/{uid}/permissions", h.updateAccountPermissions)
	mux.HandleFunc("GET /api/admin/logs", h.listOperationLogs)
	mux.HandleFunc("PUT /api/admin/settings/home", h.updateHomeSettings)
	mux.HandleFunc("POST /api/admin/pvf/refresh", h.refreshPVF)
	mux.HandleFunc("GET /api/admin/pvf/refresh/{job_id}", h.refreshPVFJob)
	mux.HandleFunc("PUT /api/admin/pvf/client-md5", h.updateClientPVFMD5)
	mux.HandleFunc("GET /api/posters/{filename}", h.posterImage)

	for _, route := range compatibilityRoutes() {
		mux.HandleFunc(route, h.notImplemented)
	}
}

func (h Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{
		"status":  "ok",
		"service": "dnf-launcher-go",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (h Handler) settingsResponse(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{"home": appsettings.Public(h.store)})
}

func (h Handler) login(w http.ResponseWriter, r *http.Request) {
	var payload loginRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.AccountName == "" || payload.Password == "" {
		writeError(w, apperror.BadRequest("Account name and password are required"))
		return
	}

	if uid, accountName, ok, err := permissions.VerifyAdmin(h.store.Tool(), payload.AccountName, payload.Password); err != nil {
		writeError(w, err)
		return
	} else if ok {
		token, err := security.CreateSessionToken(h.settings, uid, accountName, "admin")
		if err != nil {
			writeError(w, err)
			return
		}
		audit.Write(h.store.Tool(), r, uid, "auth.admin_login", "admin login")
		writeJSON(w, http.StatusOK, response{
			"access_token": token,
			"token_type":   "bearer",
			"uid":          uid,
			"account_name": accountName,
			"user_type":    "admin",
			"permissions":  permissions.All,
			"can_launch":   false,
			"tools":        permissions.VisibleTools(permissions.All),
		})
		return
	}

	account, err := accounts.Authenticate(h.store, payload.AccountName, payload.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	userPermissions, err := permissions.AccountPermissions(h.store.Tool(), account.UID)
	if err != nil {
		writeError(w, err)
		return
	}
	token, err := security.CreateSessionToken(h.settings, account.UID, account.AccountName, "game")
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(h.store.Tool(), r, account.UID, "auth.login", "desktop login")
	writeJSON(w, http.StatusOK, response{
		"access_token": token,
		"token_type":   "bearer",
		"uid":          account.UID,
		"account_name": account.AccountName,
		"user_type":    "game",
		"permissions":  userPermissions,
		"can_launch":   true,
		"tools":        permissions.VisibleTools(userPermissions),
	})
}

func (h Handler) register(w http.ResponseWriter, r *http.Request) {
	var payload registerRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if len(payload.AccountName) < 4 || len(payload.AccountName) > 16 {
		writeError(w, apperror.BadRequest("Account name must be 4-16 characters"))
		return
	}
	if !isASCII(payload.AccountName) {
		writeError(w, apperror.BadRequest("Account name must use ASCII characters"))
		return
	}
	if !isDigits(payload.AccountName) {
		writeError(w, apperror.BadRequest("Account name must contain digits only"))
		return
	}
	if payload.QQ != "" && !isDigits(payload.QQ) {
		writeError(w, apperror.BadRequest("QQ must contain digits only"))
		return
	}
	if len(payload.Password) < 6 || len(payload.Password) > 16 {
		writeError(w, apperror.BadRequest("Password must be 6-16 characters"))
		return
	}
	uid, err := accounts.Register(h.settings, h.store, payload.AccountName, payload.Password, payload.QQ)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{
		"ok":           true,
		"message":      "register success",
		"uid":          uid,
		"account_name": payload.AccountName,
	})
}

func (h Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	var payload changePasswordRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	accountName := strings.TrimSpace(payload.AccountName)
	if len(accountName) < 4 || len(accountName) > 16 {
		writeError(w, apperror.BadRequest("Account name must be 4-16 characters"))
		return
	}
	if !isDigits(accountName) {
		writeError(w, apperror.BadRequest("Account name must contain digits only"))
		return
	}
	if payload.QQ == "" || !isDigits(payload.QQ) {
		writeError(w, apperror.BadRequest("QQ must contain digits only"))
		return
	}
	if len(payload.NewPassword) < 6 || len(payload.NewPassword) > 16 {
		writeError(w, apperror.BadRequest("Password must be 6-16 characters"))
		return
	}
	uid, err := accounts.ChangePasswordByQQ(h.store, accountName, payload.QQ, payload.NewPassword)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(h.store.Tool(), r, uid, "auth.password.reset", "game account password reset by account and QQ")
	writeJSON(w, http.StatusOK, response{"ok": true})
}

func (h Handler) adminChangePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" {
		writeError(w, apperror.Forbidden("Only admin can change admin password"))
		return
	}
	var payload adminChangePasswordRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.CurrentPassword == "" {
		writeError(w, apperror.BadRequest("Current password is required"))
		return
	}
	if len(payload.NewPassword) < 6 || len(payload.NewPassword) > 128 {
		writeError(w, apperror.BadRequest("Password must be 6-128 characters"))
		return
	}
	changed, err := permissions.ChangeAdminPassword(h.store.Tool(), user.UID, payload.CurrentPassword, payload.NewPassword)
	if err != nil {
		writeError(w, err)
		return
	}
	if !changed {
		writeError(w, apperror.Unauthorized("Current password is incorrect"))
		return
	}
	audit.Write(h.store.Tool(), r, user.UID, "auth.admin_password.change", "admin password changed")
	writeJSON(w, http.StatusOK, response{"ok": true})
}

func (h Handler) pvfStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.currentUser(w, r); !ok {
		return
	}
	payload, err := pvf.Status(h.store)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (h Handler) pvfItems(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.currentUser(w, r); !ok {
		return
	}
	var itemID *int
	if rawID := strings.TrimSpace(r.URL.Query().Get("item_id")); rawID != "" {
		if !isDigits(rawID) {
			writeError(w, apperror.BadRequest("Invalid item_id"))
			return
		}
		value := parseIntDefault(rawID, 0)
		itemID = &value
	}
	payload, err := pvf.SearchItems(
		h.store,
		r.URL.Query().Get("keyword"),
		itemID,
		r.URL.Query().Get("item_type"),
		parseIntDefault(r.URL.Query().Get("limit"), 50),
		parseIntDefault(r.URL.Query().Get("page"), 1),
	)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (h Handler) gmCharacters(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	allowed := []string{"gm.mail", "gm.inventory", "gm.avatar.edit", "gm.character.edit"}
	if user.UserType != "admin" && !hasAnyPermission(user.Permissions, allowed) {
		writeError(w, apperror.Forbidden("Missing character permission"))
		return
	}
	includeDeleted := parseBool(r.URL.Query().Get("include_deleted"))
	if includeDeleted && user.UserType != "admin" && !permissions.IsAllowed("gm.character.edit", user.Permissions) {
		writeError(w, apperror.Forbidden("缺少权限：gm.character.edit"))
		return
	}
	var targetUID *int
	if rawUID := strings.TrimSpace(r.URL.Query().Get("uid")); rawUID != "" {
		value := parseIntDefault(rawUID, 0)
		targetUID = &value
	}
	payload, err := gm.Characters(
		h.store,
		user,
		r.URL.Query().Get("keyword"),
		parseIntDefault(r.URL.Query().Get("page"), 1),
		parseIntDefault(r.URL.Query().Get("limit"), 12),
		includeDeleted,
		targetUID,
	)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (h Handler) resolveGMAccount(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r, "resolve account") {
		return
	}
	var payload accountResolveRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	account, err := gm.ResolveAccount(h.store, payload.UID, payload.AccountName)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, account)
}

func (h Handler) queryCera(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.cera.charge")
	if !ok {
		return
	}
	var payload ceraQueryRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	account, err := gm.ResolveAccessibleAccount(h.store, user, payload.UID, payload.AccountName)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := gm.CeraResponse(h.store, account)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) chargeCera(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.cera.charge")
	if !ok {
		return
	}
	var payload ceraChargeRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.Amount < 0 {
		writeError(w, apperror.BadRequest("Invalid amount"))
		return
	}
	account, err := gm.ResolveAccessibleAccount(h.store, user, payload.UID, payload.AccountName)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := gm.ChargeCera(h.store, account, payload.CeraType, payload.Action, payload.Amount)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.cera.charge",
		fmt.Sprintf("target_uid=%d; type=%s; action=%s; amount=%d; cera=%d; cera_point=%d", account.UID, result.CeraType, result.Action, payload.Amount, result.Cera, result.CeraPoint),
	)
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) queryBan(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r, "query account bans") {
		return
	}
	var payload banQueryRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	account, err := gm.ResolveAccount(h.store, payload.UID, payload.AccountName)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := gm.BanStatus(h.store, account)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) setBan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" {
		writeError(w, apperror.Forbidden("Only admin can set account bans"))
		return
	}
	var payload banSetRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.Days < 1 || payload.Days > 3650 {
		writeError(w, apperror.BadRequest("Invalid ban days"))
		return
	}
	account, err := gm.ResolveAccount(h.store, payload.UID, payload.AccountName)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := gm.SetBan(h.store, account, payload.PunishType, payload.Days, payload.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.ban.set",
		fmt.Sprintf("target_uid=%d; punish_type=%d; days=%d; end_time=%s; reason=%s", account.UID, result.PunishType, payload.Days, result.EndTime, strings.TrimSpace(payload.Reason)),
	)
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) unban(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" {
		writeError(w, apperror.Forbidden("Only admin can remove account bans"))
		return
	}
	var payload banQueryRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	account, err := gm.ResolveAccount(h.store, payload.UID, payload.AccountName)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := gm.Unban(h.store, account)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(h.store.Tool(), r, user.UID, "gm.ban.unban", fmt.Sprintf("target_uid=%d", account.UID))
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, "gm.event.manage"); !ok {
		return
	}
	payload, err := gm.Events(h.store)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (h Handler) addEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.event.manage")
	if !ok {
		return
	}
	var payload eventAddRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	parameter1 := 1
	if payload.Parameter1 != nil {
		parameter1 = *payload.Parameter1
	}
	parameter2 := 0
	if payload.Parameter2 != nil {
		parameter2 = *payload.Parameter2
	}
	running, err := gm.AddEvent(h.store, payload.EventID, parameter1, parameter2)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.event.add",
		fmt.Sprintf("event_id=%d; parameter1=%d; parameter2=%d", payload.EventID, parameter1, parameter2),
	)
	writeJSON(w, http.StatusOK, response{"running": running})
}

func (h Handler) deleteEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.event.manage")
	if !ok {
		return
	}
	logID := parseIntDefault(r.PathValue("log_id"), 0)
	if logID <= 0 {
		writeError(w, apperror.BadRequest("Invalid event log id"))
		return
	}
	running, err := gm.DeleteEvent(h.store, logID)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(h.store.Tool(), r, user.UID, "gm.event.delete", fmt.Sprintf("log_id=%d", logID))
	writeJSON(w, http.StatusOK, response{"running": running})
}

func (h Handler) sendMail(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.mail")
	if !ok {
		return
	}
	var payload mailSendRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if err := validateMailRequest(payload.ItemCount, payload.Gold, payload.ItemGrade, payload.EnhancementLevel, payload.ForgeLevel, payload.AmplifyOption, payload.AmplifyValue); err != nil {
		writeError(w, err)
		return
	}
	character, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, false)
	if err != nil {
		writeError(w, err)
		return
	}
	mailPayload := buildMailPayload(
		payload.Message,
		payload.ItemID,
		payload.ItemCount,
		payload.Gold,
		payload.ItemType,
		payload.ItemGrade,
		payload.EnhancementLevel,
		payload.ForgeLevel,
		payload.AmplifyOption,
		payload.AmplifyValue,
	)
	result, err := gm.SendMail(h.store, character, user.AccountName, mailPayload)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.mail.send",
		fmt.Sprintf(
			"charac_no=%d; item_id=%d; item_count=%d; gold=%d; item_type=%s; item_grade=%d; enhancement_level=%d; forge_level=%d; amplify_option=%d; amplify_value=%d; letter_id=%d; postal_count=%d",
			character.CharacNo,
			mailPayload.ItemID,
			mailPayload.ItemCount,
			mailPayload.Gold,
			mailPayload.ItemType,
			mailAuditItemGrade(mailPayload),
			mailPayload.EnhancementLevel,
			mailPayload.ForgeLevel,
			mailPayload.AmplifyOption,
			mailPayload.AmplifyValue,
			result.LetterID,
			result.PostalCount,
		),
	)
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) sendMailAll(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" {
		writeError(w, apperror.Forbidden("Only admin can send mail to all characters"))
		return
	}
	var payload mailMassSendRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if err := validateMailRequest(payload.ItemCount, payload.Gold, payload.ItemGrade, payload.EnhancementLevel, payload.ForgeLevel, payload.AmplifyOption, payload.AmplifyValue); err != nil {
		writeError(w, err)
		return
	}
	mailPayload := buildMailPayload(
		payload.Message,
		payload.ItemID,
		payload.ItemCount,
		payload.Gold,
		payload.ItemType,
		payload.ItemGrade,
		payload.EnhancementLevel,
		payload.ForgeLevel,
		payload.AmplifyOption,
		payload.AmplifyValue,
	)
	result, err := gm.SendMailAll(h.store, user.AccountName, mailPayload)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.mail.send_all",
		fmt.Sprintf(
			"targets=%d; item_id=%d; item_count=%d; gold=%d; first_letter_id=%d; postal_count=%d",
			result.TargetCount,
			mailPayload.ItemID,
			mailPayload.ItemCount,
			mailPayload.Gold,
			result.FirstLetterID,
			result.PostalCount,
		),
	)
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) deleteMail(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.mail")
	if !ok {
		return
	}
	var payload mailDeleteRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	character, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, false)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := gm.DeleteMailForCharacter(h.store, character)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.mail.delete",
		fmt.Sprintf("charac_no=%d; deleted_count=%d", character.CharacNo, result.DeletedCount),
	)
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) deleteMailAll(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" {
		writeError(w, apperror.Forbidden("Only admin can delete mail for all characters"))
		return
	}
	result, err := gm.DeleteMailForAllCharacters(h.store)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(h.store.Tool(), r, user.UID, "gm.mail.delete_all", fmt.Sprintf("deleted_count=%d", result.DeletedCount))
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) characterJobOptions(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" && !permissions.IsAllowed("gm.character.edit", user.Permissions) {
		writeError(w, apperror.Forbidden("缺少权限：gm.character.edit"))
		return
	}
	payload, err := gm.CharacterJobOptions(h.store)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (h Handler) setCharacterLevel(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.character.edit")
	if !ok {
		return
	}
	var payload characterLevelRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.Level < 1 || payload.Level > 999 {
		writeError(w, apperror.BadRequest("Invalid level"))
		return
	}
	if _, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, false); err != nil {
		writeError(w, err)
		return
	}
	before, after, err := gm.SetLevel(h.store, payload.CharacNo, payload.Level)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.character.level",
		fmt.Sprintf("target_uid=%d; charac_no=%d; level=%d->%d", before.UID, payload.CharacNo, before.Level, after.Level),
	)
	writeJSON(w, http.StatusOK, response{"character": after})
}

func (h Handler) setCharacterJob(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.character.edit")
	if !ok {
		return
	}
	var payload characterJobRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.Job < 0 || payload.GrowType < 0 || payload.GrowType > 15 || payload.WakeFlag < 0 || payload.WakeFlag > 2 || payload.ExpertJob < 0 || payload.ExpertJob > 4 {
		writeError(w, apperror.BadRequest("Invalid job request"))
		return
	}
	if _, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, false); err != nil {
		writeError(w, err)
		return
	}
	before, after, err := gm.SetJob(h.store, payload.CharacNo, payload.Job, payload.GrowType, payload.WakeFlag, payload.ExpertJob)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.character.job",
		fmt.Sprintf("target_uid=%d; charac_no=%d; job=%d->%d; grow_type=%d->%d; expert_job=%d->%d", before.UID, payload.CharacNo, before.Job, after.Job, before.GrowType, after.GrowType, before.ExpertJob, after.ExpertJob),
	)
	writeJSON(w, http.StatusOK, response{"character": after})
}

func (h Handler) setCharacterPVPGrade(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.character.edit")
	if !ok {
		return
	}
	var payload characterPVPGradeRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if _, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, false); err != nil {
		writeError(w, err)
		return
	}
	before, after, err := gm.SetPVPGrade(h.store, payload.CharacNo, payload.PVPGrade)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.character.pvp_grade",
		fmt.Sprintf("target_uid=%d; charac_no=%d; pvp_grade=%d->%d", before.UID, payload.CharacNo, before.PVPGrade, after.PVPGrade),
	)
	writeJSON(w, http.StatusOK, response{"character": after})
}

func (h Handler) setCharacterPVPPoint(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.character.edit")
	if !ok {
		return
	}
	var payload characterPVPPointRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.PVPPoint < 0 {
		writeError(w, apperror.BadRequest("Invalid PVP point"))
		return
	}
	if _, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, false); err != nil {
		writeError(w, err)
		return
	}
	before, after, err := gm.SetPVPPoint(h.store, payload.CharacNo, payload.PVPPoint)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.character.pvp_point",
		fmt.Sprintf("target_uid=%d; charac_no=%d; pvp_point=%d->%d; win_point=%d->%d", before.UID, payload.CharacNo, before.PVPPoint, after.PVPPoint, before.WinPoint, after.WinPoint),
	)
	writeJSON(w, http.StatusOK, response{"character": after})
}

func (h Handler) deleteCharacter(w http.ResponseWriter, r *http.Request) {
	h.setCharacterVisibility(w, r, 1, "gm.character.delete")
}

func (h Handler) recoverCharacter(w http.ResponseWriter, r *http.Request) {
	h.setCharacterVisibility(w, r, 0, "gm.character.recover")
}

func (h Handler) setCharacterVisibility(w http.ResponseWriter, r *http.Request, deleteFlag int, action string) {
	user, ok := h.requirePermission(w, r, "gm.character.edit")
	if !ok {
		return
	}
	var payload characterVisibilityRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if _, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, true); err != nil {
		writeError(w, err)
		return
	}
	before, after, err := gm.SetDeleteFlag(h.store, payload.CharacNo, deleteFlag)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		action,
		fmt.Sprintf("target_uid=%d; charac_no=%d; delete_flag=%d->%d", before.UID, payload.CharacNo, before.DeleteFlag, after.DeleteFlag),
	)
	writeJSON(w, http.StatusOK, response{"character": after})
}

func (h Handler) queryInventory(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.inventory")
	if !ok {
		return
	}
	var payload inventoryQueryRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if _, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, false); err != nil {
		writeError(w, err)
		return
	}
	result, err := gm.QueryInventory(h.store, payload.CharacNo, payload.Scope)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) deleteInventoryItem(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.inventory")
	if !ok {
		return
	}
	var payload inventoryDeleteRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.Slot < 0 {
		writeError(w, apperror.BadRequest("Slot is outside inventory range"))
		return
	}
	if _, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, false); err != nil {
		writeError(w, err)
		return
	}
	result, err := gm.DeleteInventorySlots(h.store, payload.CharacNo, payload.Scope, []int{payload.Slot})
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.inventory.delete",
		fmt.Sprintf("charac_no=%d; scope=%s; slot=%d", payload.CharacNo, result["scope"], payload.Slot),
	)
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) clearInventoryScope(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.inventory")
	if !ok {
		return
	}
	var payload inventoryClearRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if _, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, false); err != nil {
		writeError(w, err)
		return
	}
	result, err := gm.ClearInventoryScope(h.store, payload.CharacNo, payload.Scope)
	if err != nil {
		writeError(w, err)
		return
	}
	deletedCount := 0
	if deletedSlots, ok := result["deleted_slots"].([]int); ok {
		deletedCount = len(deletedSlots)
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.inventory.clear",
		fmt.Sprintf("charac_no=%d; scope=%s; deleted_slots=%d", payload.CharacNo, result["scope"], deletedCount),
	)
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) avatarOptions(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, "gm.avatar.edit"); !ok {
		return
	}
	options, err := gm.AvatarHiddenOptions(h.store)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{"options": options})
}

func (h Handler) queryAvatar(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.avatar.edit")
	if !ok {
		return
	}
	var payload avatarQueryRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if _, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, false); err != nil {
		writeError(w, err)
		return
	}
	result, err := gm.QueryAvatarItems(h.store, payload.CharacNo)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) setAvatarHidden(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, "gm.avatar.edit")
	if !ok {
		return
	}
	var payload avatarHiddenRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if _, err := gm.AccessibleCharacter(h.store, user, payload.CharacNo, false); err != nil {
		writeError(w, err)
		return
	}
	result, err := gm.SetAvatarHidden(h.store, payload.CharacNo, payload.UIIDs, payload.HiddenOption)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"gm.avatar.hidden",
		fmt.Sprintf("charac_no=%d; ui_ids=%v; hidden_option=%d", payload.CharacNo, payload.UIIDs, payload.HiddenOption),
	)
	writeJSON(w, http.StatusOK, result)
}

func (h Handler) directLaunch(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "game" {
		writeError(w, apperror.Forbidden("Admin cannot launch game directly"))
		return
	}
	payload, err := launcher.Direct(h.settings, h.store, user.UID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (h Handler) listPermissions(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r, "list permissions") {
		return
	}
	writeJSON(w, http.StatusOK, response{"permissions": permissions.All})
}

func (h Handler) listAccounts(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r, "list accounts") {
		return
	}
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	limit := clamp(parseIntDefault(r.URL.Query().Get("limit"), 50), 1, 100)
	rows, err := queryAccounts(h.store.Game(), keyword, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	accounts := []response{}
	for _, row := range rows {
		userPermissions, err := permissions.AccountPermissions(h.store.Tool(), row.UID)
		if err != nil {
			writeError(w, err)
			return
		}
		accounts = append(accounts, response{
			"uid":          row.UID,
			"account_name": row.AccountName,
			"permissions":  userPermissions,
			"tools":        permissions.VisibleTools(userPermissions),
		})
	}
	writeJSON(w, http.StatusOK, response{"accounts": accounts})
}

func (h Handler) updateAccountPermissions(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r, "update account permissions") {
		return
	}
	uid, err := security.UIDFromPath(r.URL.Path, "/api/admin/accounts/", "/permissions")
	if err != nil {
		writeError(w, apperror.BadRequest("Invalid account UID"))
		return
	}
	var payload setPermissionsRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	updated, err := permissions.SetAccountPermissions(h.store.Tool(), uid, payload.Permissions)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		0,
		"admin.permissions.update",
		fmt.Sprintf("target_uid=%d; permissions=%v", uid, updated),
	)
	writeJSON(w, http.StatusOK, response{
		"uid":         uid,
		"permissions": updated,
	})
}

func (h Handler) listOperationLogs(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r, "list operation logs") {
		return
	}
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	page := clamp(parseIntDefault(r.URL.Query().Get("page"), 1), 1, 2147483647)
	limit := clamp(parseIntDefault(r.URL.Query().Get("limit"), 50), 1, 200)
	total, rows, err := queryLogs(h.store.Tool(), keyword, page, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{
		"logs":  rows,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func (h Handler) updateHomeSettings(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" {
		writeError(w, apperror.Forbidden("Only admin can update home settings"))
		return
	}
	var payload homeSettingsRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	home, err := appsettings.UpdateHome(
		h.store,
		payload.HomeTitle,
		payload.HomeEyebrow,
		payload.ClientDownloadURL,
		payload.Announcements,
	)
	if err != nil {
		writeError(w, err)
		return
	}
	count := 0
	if announcements, ok := home["announcements"].([]appsettings.Announcement); ok {
		count = len(announcements)
	}
	state := "empty"
	if strings.TrimSpace(payload.ClientDownloadURL) != "" {
		state = "set"
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"admin.settings.home",
		fmt.Sprintf("home_title=%s; announcements=%d; client_download_url=%s", home["home_title"], count, state),
	)
	writeJSON(w, http.StatusOK, response{"home": home})
}

func (h Handler) updateClientPVFMD5(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" {
		writeError(w, apperror.Forbidden("Only admin can update PVF settings"))
		return
	}
	var payload pvfClientMD5Request
	if !decodeJSON(w, r, &payload) {
		return
	}
	clientPVFMD5, err := normalizeMD5(payload.ClientPVFMD5)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := appsettings.SetClientPVFMD5(h.store, clientPVFMD5); err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"admin.pvf.client_md5",
		fmt.Sprintf("client_pvf_md5=%s", clientPVFMD5),
	)
	writeJSON(w, http.StatusOK, response{"client_pvf_md5": clientPVFMD5})
}

func (h Handler) refreshPVF(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" {
		writeError(w, apperror.Forbidden("Only admin can refresh PVF"))
		return
	}
	var payload pvfRefreshRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	if strings.TrimSpace(payload.PVFPath) == "" {
		writeError(w, apperror.BadRequest("PVF path is required"))
		return
	}
	task, err := pvf.StartRefresh(h.store, payload.PVFPath, payload.Encode, user.UID)
	if err != nil {
		writeError(w, err)
		return
	}
	audit.Write(
		h.store.Tool(),
		r,
		user.UID,
		"admin.pvf.refresh",
		fmt.Sprintf("job_id=%d; path=%s; encode=%s", task.ID, task.PVFPath, task.Encode),
	)
	writeJSON(w, http.StatusAccepted, response{"job": pvf.TaskPayload(task)})
}

func (h Handler) refreshPVFJob(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" {
		writeError(w, apperror.Forbidden("Only admin can view PVF refresh jobs"))
		return
	}
	jobID := parseIntDefault(r.PathValue("job_id"), 0)
	if jobID <= 0 {
		writeError(w, apperror.BadRequest("Invalid PVF refresh job id"))
		return
	}
	task, err := pvf.RefreshJob(h.store, int64(jobID))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{"job": pvf.TaskPayload(task)})
}

func (h Handler) posterImage(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	path, ok := posterPath(filename)
	if !ok {
		writeError(w, apperror.New(http.StatusNotFound, "Poster not found"))
		return
	}
	if contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path))); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	http.ServeFile(w, r, path)
}

func (h Handler) notImplemented(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, response{
		"detail": "Go server route is registered but not implemented yet",
		"method": r.Method,
		"path":   r.URL.Path,
	})
}

func compatibilityRoutes() []string {
	return []string{}
}

func posterPath(filename string) (string, bool) {
	safeName := filepath.Base(filename)
	if safeName != filename {
		return "", false
	}
	extension := strings.ToLower(filepath.Ext(safeName))
	if extension != "" && !allowedPosterExtension(extension) {
		return "", false
	}
	posterDir, err := posterDirectory()
	if err != nil {
		return "", false
	}
	if extension != "" {
		path := filepath.Join(posterDir, safeName)
		if fileExists(path) {
			return path, true
		}
		return "", false
	}
	for _, candidateExtension := range posterExtensions() {
		path := filepath.Join(posterDir, safeName+candidateExtension)
		if fileExists(path) {
			return path, true
		}
	}
	path := filepath.Join(posterDir, safeName)
	if fileExists(path) {
		return path, true
	}
	return "", false
}

func posterDirectory() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(executable), "posters"), nil
}

func allowedPosterExtension(extension string) bool {
	for _, item := range posterExtensions() {
		if extension == item {
			return true
		}
	}
	return false
}

func posterExtensions() []string {
	return []string{".jpg", ".jpeg", ".png", ".webp"}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func buildMailPayload(
	message string,
	itemID *int,
	itemCount int,
	gold int,
	itemType string,
	itemGrade int,
	enhancementLevel int,
	forgeLevel int,
	amplifyOption int,
	amplifyValue int,
) gm.MailPayload {
	resolvedItemID := 0
	if itemID != nil {
		resolvedItemID = *itemID
	}
	return gm.MailPayload{
		Message:          message,
		ItemID:           resolvedItemID,
		ItemCount:        itemCount,
		Gold:             gold,
		ItemType:         itemType,
		ItemGrade:        itemGrade,
		EnhancementLevel: enhancementLevel,
		ForgeLevel:       forgeLevel,
		AmplifyOption:    amplifyOption,
		AmplifyValue:     amplifyValue,
	}
}

func validateMailRequest(itemCount int, gold int, itemGrade int, enhancementLevel int, forgeLevel int, amplifyOption int, amplifyValue int) error {
	if itemCount < 0 || gold < 0 {
		return apperror.BadRequest("Invalid mail request")
	}
	if itemGrade < 0 || itemGrade > 4294967295 {
		return apperror.BadRequest("Invalid mail request")
	}
	if enhancementLevel < 0 || enhancementLevel > 31 || forgeLevel < 0 || forgeLevel > 31 {
		return apperror.BadRequest("Invalid mail request")
	}
	if amplifyOption < 0 || amplifyOption > 4 || amplifyValue < 0 || amplifyValue > 65535 {
		return apperror.BadRequest("Invalid mail request")
	}
	return nil
}

func mailAuditItemGrade(payload gm.MailPayload) int {
	if strings.ToLower(strings.TrimSpace(payload.ItemType)) == "equipment" {
		return 1
	}
	return payload.ItemGrade
}

type accountRow struct {
	UID         int
	AccountName string
}

func queryAccounts(gameDB *sql.DB, keyword string, limit int) ([]accountRow, error) {
	var rows *sql.Rows
	var err error
	if keyword != "" {
		rows, err = gameDB.Query(`
			SELECT uid, accountname
			FROM accounts
			WHERE accountname LIKE ? OR CAST(uid AS CHAR)=?
			ORDER BY uid DESC
			LIMIT ?
		`, "%"+keyword+"%", keyword, limit)
	} else {
		rows, err = gameDB.Query(`
			SELECT uid, accountname
			FROM accounts
			ORDER BY uid DESC
			LIMIT ?
		`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []accountRow{}
	for rows.Next() {
		var item accountRow
		if err := rows.Scan(&item.UID, &item.AccountName); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func queryLogs(toolDB *sql.DB, keyword string, page int, limit int) (int, []response, error) {
	offset := (page - 1) * limit
	where := "1=1"
	args := []any{}
	if keyword != "" {
		where = "(action LIKE ? OR detail LIKE ? OR ip LIKE ? OR CAST(uid AS CHAR)=?)"
		like := "%" + keyword + "%"
		args = append(args, like, like, like, keyword)
	}
	var total int
	countArgs := append([]any{}, args...)
	if err := toolDB.QueryRow("SELECT COUNT(*) AS total FROM operation_logs WHERE "+where, countArgs...).Scan(&total); err != nil {
		return 0, nil, err
	}
	queryArgs := append(args, limit, offset)
	rows, err := toolDB.Query(`
		SELECT id, uid, action, detail, ip, created_at
		FROM operation_logs
		WHERE `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()
	result := []response{}
	for rows.Next() {
		var id int64
		var uid int
		var action string
		var detail sql.NullString
		var ip sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&id, &uid, &action, &detail, &ip, &createdAt); err != nil {
			return 0, nil, err
		}
		result = append(result, response{
			"id":         id,
			"uid":        uid,
			"action":     action,
			"detail":     nullString(detail),
			"ip":         nullString(ip),
			"created_at": createdAt.Format("2006-01-02 15:04:05"),
		})
	}
	return total, result, rows.Err()
}

func (h Handler) currentUser(w http.ResponseWriter, r *http.Request) (security.User, bool) {
	token, err := security.BearerToken(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, err)
		return security.User{}, false
	}
	user, err := security.ParseSessionToken(h.settings, token)
	if err != nil {
		writeError(w, err)
		return security.User{}, false
	}
	if user.UserType == "admin" {
		user.Permissions = permissions.All
	} else {
		userPermissions, err := permissions.AccountPermissions(h.store.Tool(), user.UID)
		if err != nil {
			writeError(w, err)
			return security.User{}, false
		}
		user.Permissions = userPermissions
	}
	return user, true
}

func (h Handler) requireAdmin(w http.ResponseWriter, r *http.Request, action string) bool {
	user, ok := h.currentUser(w, r)
	if !ok {
		return false
	}
	if user.UserType != "admin" {
		writeError(w, apperror.Forbidden("Only admin can "+action))
		return false
	}
	return true
}

func (h Handler) requirePermission(w http.ResponseWriter, r *http.Request, permission string) (security.User, bool) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return security.User{}, false
	}
	if user.UserType != "admin" && !permissions.IsAllowed(permission, user.Permissions) {
		writeError(w, apperror.Forbidden("缺少权限："+permission))
		return security.User{}, false
	}
	return user, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(w, apperror.BadRequest("Invalid JSON body"))
		return false
	}
	return true
}

func parseIntDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	result := 0
	for _, c := range value {
		if c < '0' || c > '9' {
			return fallback
		}
		result = result*10 + int(c-'0')
	}
	return result
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

func parseBool(value string) bool {
	return value == "1" || strings.EqualFold(value, "true") || strings.EqualFold(value, "yes")
}

func hasAnyPermission(userPermissions []string, required []string) bool {
	for _, item := range required {
		if permissions.IsAllowed(item, userPermissions) {
			return true
		}
	}
	return false
}

func isASCII(value string) bool {
	for _, c := range value {
		if c > 127 {
			return false
		}
	}
	return true
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

func normalizeMD5(value string) (string, error) {
	md5 := strings.ToUpper(strings.TrimSpace(value))
	if md5 == "" {
		return "", nil
	}
	if len(md5) != 32 {
		return "", apperror.BadRequest("PVF MD5 must be 32 hex characters")
	}
	for _, c := range md5 {
		if (c < '0' || c > '9') && (c < 'A' || c > 'F') {
			return "", apperror.BadRequest("PVF MD5 must be 32 hex characters")
		}
	}
	return md5, nil
}

func nullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func writeError(w http.ResponseWriter, err error) {
	var appErr apperror.Error
	if errors.As(err, &appErr) {
		writeJSON(w, appErr.Status, response{"detail": appErr.Detail})
		return
	}
	log.Printf("internal api error: %v", err)
	writeJSON(w, http.StatusInternalServerError, response{"detail": "Internal server error"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, strings.TrimSpace(err.Error()), http.StatusInternalServerError)
	}
}
