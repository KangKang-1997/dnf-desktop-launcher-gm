import hashlib
import json
import struct
import threading
import zlib
from datetime import datetime, timedelta

from fastapi import Depends, FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import FileResponse
from typing import Optional
from pathlib import Path

from .core.config import settings
from .core.audit import write_audit_log
from .core.db import execute, fetch_all, fetch_one, game_execute, game_fetch_all, game_fetch_one, mysql_cursor, transactional
from .services.accounts import authenticate_account, register_account
from .services.launcher import create_direct_launch
from .models import (
    AccountResolveRequest,
    AvatarHiddenRequest,
    AvatarQueryRequest,
    BanQueryRequest,
    BanSetRequest,
    CeraChargeRequest,
    CeraQueryRequest,
    ChangePasswordRequest,
    AdminChangePasswordRequest,
    CharacterJobRequest,
    CharacterLevelRequest,
    CharacterPvpGradeRequest,
    CharacterPvpPointRequest,
    CharacterVisibilityRequest,
    EventAddRequest,
    HomeSettingsRequest,
    InventoryClearRequest,
    InventoryDeleteRequest,
    InventoryQueryRequest,
    LoginRequest,
    MailDeleteRequest,
    MailMassSendRequest,
    MailSendRequest,
    PvfClientMd5Request,
    PvfRefreshRequest,
    RegisterRequest,
    SetPermissionsRequest,
)
from .services.pvf.cache_manager import (
    ensure_pvf_tables,
    get_active_pvf,
    get_exp_for_level,
    get_pvf_data,
    get_item_detail,
    refresh_pvf_cache,
    search_items,
)
from .core.permissions import (
    ALL_PERMISSIONS,
    change_admin_password as update_admin_password,
    ensure_permission_tables,
    get_setting,
    get_account_permissions,
    set_account_permissions,
    set_setting,
    verify_admin,
    visible_tools,
)
from .core.security import create_session_token, get_current_user, require_permission


def _is_ascii(value):
    try:
        value.encode("ascii")
        return True
    except UnicodeEncodeError:
        return False


def _require_admin(user, action="operate"):
    if user.get("user_type") != "admin":
        raise HTTPException(status_code=403, detail="Only admin can {0}".format(action))


PASSWORD_RESET_WINDOW = timedelta(minutes=10)
PASSWORD_RESET_IP_LIMIT = 20
PASSWORD_RESET_ACCOUNT_LIMIT = 5
_password_reset_attempts = {}
_password_reset_lock = threading.Lock()


def _is_digits(value):
    return bool(value) and value.isdigit()


def _check_password_reset_rate(request, account_name):
    now = datetime.utcnow()
    cutoff = now - PASSWORD_RESET_WINDOW
    ip = request.client.host if request.client else ""
    keys = (
        ("ip", ip, PASSWORD_RESET_IP_LIMIT),
        ("account", account_name, PASSWORD_RESET_ACCOUNT_LIMIT),
    )
    with _password_reset_lock:
        for kind, value, limit in keys:
            key = (kind, value)
            recent = [
                stamp
                for stamp in _password_reset_attempts.get(key, [])
                if stamp > cutoff
            ]
            if len(recent) >= limit:
                raise HTTPException(
                    status_code=429, detail="Too many password reset attempts"
                )
            recent.append(now)
            _password_reset_attempts[key] = recent


def _normalize_md5(value):
    md5 = (value or "").strip().upper()
    if not md5:
        return ""
    if len(md5) != 32 or any(char not in "0123456789ABCDEF" for char in md5):
        raise HTTPException(status_code=400, detail="PVF MD5 must be 32 hex characters")
    return md5


def _client_pvf_md5():
    configured = (get_setting("client_pvf_md5", "") or "").strip().upper()
    if configured:
        return configured
    return ""


def _resolve_gm_account(uid=None, account_name=""):
    account_name = (account_name or "").strip()
    if uid is not None:
        row = game_fetch_one(
            "SELECT uid, accountname FROM accounts WHERE uid=%s",
            (uid,),
        )
    elif account_name:
        row = game_fetch_one(
            "SELECT uid, accountname FROM accounts WHERE accountname=%s",
            (account_name,),
        )
    else:
        raise HTTPException(status_code=400, detail="Account name or UID is required")
    if not row:
        raise HTTPException(status_code=404, detail="Account not found")
    return {"uid": int(row["uid"]), "account_name": row["accountname"]}


def _resolve_accessible_account(user, uid=None, account_name=""):
    if user.get("user_type") == "admin":
        return _resolve_gm_account(uid, account_name)
    current_uid = int(user["uid"])
    current_name = str(user.get("account_name") or "")
    if uid is not None and int(uid) != current_uid:
        raise HTTPException(status_code=404, detail="Account not found")
    if account_name and account_name.strip() != current_name:
        raise HTTPException(status_code=404, detail="Account not found")
    return _resolve_gm_account(current_uid, "")


def _ensure_cera_rows(uid):
    cera = execute(
        """
        INSERT INTO cash_cera(account, cera, mod_date, reg_date)
        SELECT %s, 0, NOW(), NOW()
        FROM DUAL
        WHERE NOT EXISTS (SELECT 1 FROM cash_cera WHERE account=%s)
        """,
        (uid, uid),
        "taiwan_billing",
    )
    cera_point = execute(
        """
        INSERT INTO cash_cera_point(account, cera_point, mod_date, reg_date)
        SELECT %s, 0, NOW(), NOW()
        FROM DUAL
        WHERE NOT EXISTS (SELECT 1 FROM cash_cera_point WHERE account=%s)
        """,
        (uid, uid),
        "taiwan_billing",
    )
    return cera, cera_point


def _get_cera_balance(uid):
    _ensure_cera_rows(uid)
    cera = execute_fetch_cera(
        "SELECT cera FROM cash_cera WHERE account=%s",
        (uid,),
        "cera",
    )
    cera_point = execute_fetch_cera(
        "SELECT cera_point FROM cash_cera_point WHERE account=%s",
        (uid,),
        "cera_point",
    )
    return {"cera": cera, "cera_point": cera_point}


def _get_stack_limit(item_id):
    detail = get_item_detail(item_id)
    stack_limit = detail.get("[stack limit]")
    if isinstance(stack_limit, list) and stack_limit:
        try:
            return max(1, int(stack_limit[0]))
        except Exception:
            return 2147483647
    return 2147483647


def _insert_mail_message(cursor, charac_no, sender, message):
    reg_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    cursor.execute(
        """
        INSERT INTO taiwan_cain_2nd.letter
            (charac_no, send_charac_no, send_charac_name, letter_text, reg_date, stat)
        VALUES(%s, 0, %s, %s, %s, 1)
        """,
        (
            charac_no,
            sender.encode("utf-8"),
            message.encode("utf-8"),
            reg_time,
        ),
    )
    return int(cursor.lastrowid)


def _validate_mail_item_count(item_id, item_count, item_type=""):
    item_id = int(item_id or 0)
    if item_id <= 0:
        return

    is_equipment = (item_type or "").strip().lower() == "equipment"
    if is_equipment:
        return

    total_count = int(item_count or 0)
    if total_count <= 0:
        raise HTTPException(status_code=400, detail="Item count must be greater than 0")


def _insert_mail_postal(
    cursor,
    charac_no,
    letter_id,
    sender,
    message,
    item_id,
    item_count,
    gold,
    item_type="",
    item_grade=0,
    enhancement_level=0,
    forge_level=0,
    amplify_option=0,
    amplify_value=0,
):
    if not item_id and gold <= 0:
        return 0
    item_id = int(item_id or 0)
    is_equipment = item_id > 0 and (item_type or "").strip().lower() == "equipment"
    total_count = int(item_count or 0)
    if item_id > 0 and not is_equipment and total_count <= 0:
        raise HTTPException(status_code=400, detail="Item count must be greater than 0")
    if item_id <= 0:
        total_count = 1
    elif is_equipment:
        total_count = 1
    stack_limit = 1 if is_equipment else (_get_stack_limit(item_id) if item_id > 0 else 1)
    sent_count = 0
    postal_count = 0
    current_letter_id = letter_id
    current_gold = int(gold or 0)
    while sent_count < total_count:
        if postal_count and postal_count % 10 == 0:
            current_letter_id = _insert_mail_message(cursor, charac_no, sender, message)
        current_count = min(stack_limit, total_count - sent_count)
        add_info = 1 if is_equipment else current_count
        occ_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        cursor.execute(
            """
            INSERT INTO taiwan_cain_2nd.postal (
                occ_time, send_charac_name, receive_charac_no,
                amplify_option, amplify_value, seperate_upgrade, seal_flag,
                item_id, add_info, upgrade, gold, letter_id,
                avata_flag, creature_flag, endurance, unlimit_flag
            )
            VALUES(%s, %s, %s, %s, %s, %s, 0, %s, %s, %s, %s, %s, 0, 0, 0, 1)
            """,
            (
                occ_time,
                sender.encode("utf-8"),
                charac_no,
                int(amplify_option or 0) if is_equipment else 0,
                int(amplify_value or 0) if is_equipment else 0,
                int(forge_level or 0) if is_equipment else 0,
                item_id,
                add_info,
                int(enhancement_level or 0) if is_equipment else 0,
                current_gold,
                current_letter_id,
            ),
        )
        current_gold = 0
        sent_count += current_count
        postal_count += 1
    return postal_count


def execute_fetch_cera(sql, args, field):
    row = fetch_one(sql, args, "taiwan_billing")
    if not row:
        return 0
    return int(row[field] or 0)


def _cera_response(account):
    balance = _get_cera_balance(account["uid"])
    return {
        "uid": account["uid"],
        "account_name": account["account_name"],
        "cera": balance["cera"],
        "cera_point": balance["cera_point"],
    }


def _normalize_cera_type(value):
    value = (value or "").strip()
    if value not in ("cera", "cera_point"):
        raise HTTPException(status_code=400, detail="Invalid cera type")
    return value


def _normalize_cera_action(value):
    value = (value or "").strip()
    if value not in ("add", "set"):
        raise HTTPException(status_code=400, detail="Invalid cera action")
    return value


def _normalize_punish_type(value):
    value = int(value)
    if value not in (1, 4):
        raise HTTPException(status_code=400, detail="Invalid punish type")
    return value


def _punish_type_name(value):
    return {1: "禁止登陆", 4: "限制交易", 11: "限制交易"}.get(int(value), str(value))


def _format_datetime(value):
    if not value:
        return ""
    if isinstance(value, str):
        return value
    return value.strftime("%Y-%m-%d %H:%M:%S")


def _ban_status(account):
    result = {
        "uid": account["uid"],
        "account_name": account["account_name"],
        "banned": False,
        "punish_type": None,
        "punish_type_name": "",
        "start_time": "",
        "end_time": "",
        "reason": "",
    }
    row = game_fetch_one(
        """
        SELECT m_id, punish_type, punish_value, apply_flag, start_time, end_time, reason
        FROM member_punish_info
        WHERE m_id=%s AND apply_flag!=0 AND end_time>NOW()
        ORDER BY end_time DESC
        LIMIT 1
        """,
        (account["uid"],),
    )
    if not row:
        return result
    result.update(
        {
            "banned": True,
            "punish_type": int(row["punish_type"]),
            "punish_type_name": _punish_type_name(row["punish_type"]),
            "start_time": _format_datetime(row["start_time"]),
            "end_time": _format_datetime(row["end_time"]),
            "reason": row["reason"] or "",
        }
    )
    return result


def _decode_charac_name_hex(value):
    if value is None:
        return ""
    if isinstance(value, bytes):
        value = value.decode("ascii")
    raw = bytes.fromhex(str(value))
    text = raw.decode("utf-8", errors="replace")
    s1 = text.encode("latin1", "replace")
    s2 = text.encode("cp1252", "replace")
    rebuilt = b""
    for index in range(len(s1)):
        if s1[index : index + 1] == b"?":
            rebuilt += s2[index : index + 1]
        else:
            rebuilt += s1[index : index + 1]
    return rebuilt.decode("utf-8", errors="replace")


EXPERT_JOB_MAP = {
    0: "无职业",
    1: "附魔师",
    2: "炼金术师",
    3: "分解师",
    4: "控偶师",
}


PVP_RANK_LIST = (
    ["{0}级".format(i) for i in range(10, 0, -1)]
    + ["{0}段".format(i) for i in range(1, 11)]
    + ["至尊{0}".format(i) for i in range(1, 11)]
    + ["达人", "名人", "小霸王", "霸王", "斗神"]
)


def _normalize_int(value, default=0):
    try:
        return int(value)
    except (TypeError, ValueError):
        return default


def _job_dict():
    raw = get_pvf_data("job", {}) or {}
    result = {}
    for job_key, grow_map in raw.items():
        try:
            job_id = int(job_key)
        except (TypeError, ValueError):
            continue
        if not isinstance(grow_map, dict):
            result[job_id] = {}
            continue
        result[job_id] = {}
        for grow_key, grow_name in grow_map.items():
            try:
                grow_id = int(grow_key)
            except (TypeError, ValueError):
                continue
            result[job_id][grow_id] = str(grow_name)
    return result


def _job_name(job, grow_type):
    job = _normalize_int(job)
    grow_type = _normalize_int(grow_type)
    grow_map = _job_dict().get(job)
    if isinstance(grow_map, dict):
        name = grow_map.get(grow_type % 16)
        if name:
            return name
    return "{0}/{1}".format(job, grow_type)


def _expert_job_name(expert_job):
    expert_job = _normalize_int(expert_job)
    return EXPERT_JOB_MAP.get(expert_job, str(expert_job))


def _pvp_rank_name(pvp_grade):
    pvp_grade = _normalize_int(pvp_grade)
    if 0 <= pvp_grade < len(PVP_RANK_LIST):
        return PVP_RANK_LIST[pvp_grade]
    return str(pvp_grade)


def _set_pvp_grade(charac_no, pvp_grade):
    if pvp_grade < 0 or pvp_grade >= len(PVP_RANK_LIST):
        raise HTTPException(status_code=400, detail="Invalid PVP grade")
    before = _get_character(charac_no)
    execute(
        "UPDATE pvp_result SET pvp_grade=%s WHERE charac_no=%s",
        (pvp_grade, charac_no),
        "taiwan_cain",
    )
    after = _get_character(charac_no)
    return before, after


def _set_pvp_point(charac_no, pvp_point):
    before = _get_character(charac_no)
    execute(
        "UPDATE pvp_result SET pvp_point=%s, win_point=%s WHERE charac_no=%s",
        (pvp_point, pvp_point, charac_no),
        "taiwan_cain",
    )
    after = _get_character(charac_no)
    return before, after


def _character_job_options():
    jobs = []
    for job_id, grow_map in sorted(_job_dict().items()):
        grow_types = [
            {"grow_type": grow_id, "name": grow_name}
            for grow_id, grow_name in sorted(grow_map.items())
        ]
        jobs.append(
            {
                "job": job_id,
                "name": grow_map.get(0) or "职业 {0}".format(job_id),
                "grow_types": grow_types,
            }
        )
    expert_jobs = [
        {"expert_job": expert_job, "name": name}
        for expert_job, name in sorted(EXPERT_JOB_MAP.items())
    ]
    pvp_ranks = [
        {"pvp_grade": index, "name": name}
        for index, name in enumerate(PVP_RANK_LIST)
    ]
    return {"jobs": jobs, "expert_jobs": expert_jobs, "pvp_ranks": pvp_ranks}


def _character_response_row(row):
    job = _normalize_int(row["job"])
    grow_type = _normalize_int(row["grow_type"])
    expert_job = _normalize_int(row["expert_job"])
    charac_name = _decode_charac_name_hex(row.get("charac_name_hex"))
    pvp_grade = _normalize_int(row["pvp_grade"])
    result = {
        "uid": int(row["m_id"]),
        "charac_no": int(row["charac_no"]),
        "charac_name": charac_name,
        "level": int(row["lev"] or 0),
        "job": job,
        "grow_type": grow_type,
        "grow_type_base": grow_type % 16,
        "wake_flag": grow_type // 16,
        "job_name": _job_name(job, grow_type),
        "delete_flag": int(row["delete_flag"] or 0),
        "expert_job": expert_job,
        "expert_job_name": _expert_job_name(expert_job),
        "pvp_grade": pvp_grade,
        "pvp_grade_name": _pvp_rank_name(pvp_grade),
        "pvp_win": _normalize_int(row["pvp_win"]),
        "pvp_point": _normalize_int(row["pvp_point"]),
        "win_point": _normalize_int(row["win_point"]),
    }
    return result


def _get_character(charac_no):
    with mysql_cursor("taiwan_cain", charset="latin1") as cursor:
        cursor.execute(
            """
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
            WHERE c.charac_no=%s
            """,
            (charac_no,),
        )
        rows = list(cursor.fetchall())
    row = rows[0] if rows else None
    if not row:
        raise HTTPException(status_code=404, detail="Character not found")
    return _character_response_row(row)


def _get_accessible_character(user, charac_no, allow_deleted=False):
    character = _get_character(charac_no)
    if not allow_deleted and character["delete_flag"] != 0:
        raise HTTPException(status_code=404, detail="Character not found")
    if user.get("user_type") != "admin" and character["uid"] != int(user["uid"]):
        raise HTTPException(status_code=404, detail="Character not found")
    return character


def _set_character_delete_flag(charac_no, delete_flag):
    before = _get_character(charac_no)
    execute(
        "UPDATE charac_info SET delete_flag=%s WHERE charac_no=%s",
        (delete_flag, charac_no),
        "taiwan_cain",
    )
    after = _get_character(charac_no)
    return before, after


def _set_character_job(payload):
    before = _get_character(payload.charac_no)
    job_map = _job_dict()
    grow_map = job_map.get(payload.job)
    if job_map and not isinstance(grow_map, dict):
        raise HTTPException(status_code=400, detail="Invalid job")
    if isinstance(grow_map, dict) and grow_map and payload.grow_type not in grow_map:
        raise HTTPException(status_code=400, detail="Invalid grow type")
    if payload.expert_job not in EXPERT_JOB_MAP:
        raise HTTPException(status_code=400, detail="Invalid expert job")
    grow_type = payload.grow_type + payload.wake_flag * 0x10
    execute(
        "UPDATE charac_info SET job=%s, grow_type=%s, expert_job=%s WHERE charac_no=%s",
        (payload.job, grow_type, payload.expert_job, payload.charac_no),
        "taiwan_cain",
    )
    after = _get_character(payload.charac_no)
    return before, after


ITEM_TYPE_NAMES = {
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

INCREASE_TYPE_NAMES = {
    0x00: "空",
    0x01: "异次元体力",
    0x02: "异次元精神",
    0x03: "异次元力量",
    0x04: "异次元智力",
}

INVENTORY_SCOPES = {
    "inventory": {
        "database": "taiwan_cain_2nd",
        "table": "inventory",
        "column": "inventory",
        "name": "物品栏",
        "where": "charac_no",
    },
    "equipslot": {
        "database": "taiwan_cain_2nd",
        "table": "inventory",
        "column": "equipslot",
        "name": "穿戴栏",
        "where": "charac_no",
    },
    "creature": {
        "database": "taiwan_cain_2nd",
        "table": "inventory",
        "column": "creature",
        "name": "宠物栏",
        "where": "charac_no",
    },
    "cargo": {
        "database": "taiwan_cain_2nd",
        "table": "charac_inven_expand",
        "column": "cargo",
        "name": "角色仓库",
        "where": "charac_no",
    },
    "account_cargo": {
        "database": "taiwan_cain",
        "table": "account_cargo",
        "column": "cargo",
        "name": "账号仓库",
        "where": "m_id",
    },
}


AVATAR_HIDDEN_MAP = {
    "physical attack": "力量",
    "magical attack": "智力",
    "physical defense": "体力",
    "magical defense": "精神",
    "HP MAX": "HP MAX",
    "MP MAX": "MP MAX",
    "HP regen speed": "HP 恢复",
    "MP Regen speed": "MP 恢复",
    "attack speed": "攻击速度",
    "move speed": "移动速度",
    "cast speed": "施放速度",
    "inventory limit": "负重上限",
    "stuck": "命中率",
    "stuck resistance": "回避率",
    "all activestatus resistance": "异常抗性",
    "hit recovery": "硬直",
    "equipment magical defence": "魔法防御",
    "equipment physical defence": "物理防御",
    "jump power": "跳跃力",
    "physical critical hit": "物理暴击",
    "magical critical hit": "魔法暴击",
    "": "",
}


def _item_name_map(item_ids):
    item_ids = sorted({int(item_id) for item_id in item_ids if item_id})
    if not item_ids:
        return {}
    active = get_active_pvf()
    if not active:
        return {}
    placeholders = ", ".join(["%s"] * len(item_ids))
    rows = fetch_all(
        "SELECT item_id, item_name FROM pvf_items WHERE pvf_md5=%s AND item_id IN ({0})".format(
            placeholders
        ),
        tuple([active["md5"]] + item_ids),
    )
    return {int(row["item_id"]): row["item_name"] for row in rows}


def _parse_item_slot(index, item_bytes):
    if len(item_bytes) < 61:
        item_bytes = b"\x00" * 61
    item_type = item_bytes[1]
    item_id = struct.unpack("I", item_bytes[2:6])[0]
    if item_type == 0 and item_id == 0:
        return None
    item_type_name = ITEM_TYPE_NAMES.get(item_type, "未知")
    if item_type_name == "装备":
        count_or_grade = struct.unpack("!I", item_bytes[7:11])[0]
    else:
        count_or_grade = struct.unpack("I", item_bytes[7:11])[0]
    increase_type = item_bytes[17]
    return {
        "slot": index,
        "item_id": int(item_id),
        "item_name": "",
        "item_type": int(item_type),
        "item_type_name": item_type_name,
        "is_seal": int(item_bytes[0]),
        "enhancement_level": int(item_bytes[6] & 0x1F),
        "count_or_grade": int(count_or_grade),
        "durability": int(struct.unpack("H", item_bytes[11:13])[0]),
        "orb": int(struct.unpack("I", item_bytes[13:17])[0]),
        "increase_type": int(increase_type),
        "increase_type_name": INCREASE_TYPE_NAMES.get(increase_type, str(increase_type)),
        "increase_value": int(struct.unpack("H", item_bytes[18:20])[0]),
        "forge_level": int(item_bytes[51]),
    }


def _unpack_inventory_blob(blob):
    if not blob:
        return []
    if isinstance(blob, str):
        blob = blob.encode("latin1", "replace")
    try:
        items_bytes = zlib.decompress(blob[4:])
    except Exception as exc:
        raise HTTPException(status_code=500, detail="Inventory blob decompress failed") from exc
    items = []
    count = len(items_bytes) // 61
    for index in range(count):
        item = _parse_item_slot(index, items_bytes[index * 61 : (index + 1) * 61])
        if item:
            items.append(item)
    names = _item_name_map([item["item_id"] for item in items])
    for item in items:
        item["item_name"] = names.get(item["item_id"], "")
    return items


def _inventory_slot_count(blob):
    if not blob:
        return 0
    if isinstance(blob, str):
        blob = blob.encode("latin1", "replace")
    try:
        return len(zlib.decompress(blob[4:])) // 61
    except Exception as exc:
        raise HTTPException(status_code=500, detail="Inventory blob decompress failed") from exc


def _inventory_target(charac_no, scope):
    scope = (scope or "inventory").strip()
    scope_info = INVENTORY_SCOPES.get(scope)
    if not scope_info:
        raise HTTPException(status_code=400, detail="Invalid inventory scope")
    character = _get_character(charac_no)
    target_id = character["uid"] if scope_info["where"] == "m_id" else charac_no
    return character, scope, scope_info, target_id


def _load_inventory_blob(charac_no, scope):
    character, scope, scope_info, target_id = _inventory_target(charac_no, scope)
    sql = "SELECT `{0}` AS item_blob FROM `{1}` WHERE `{2}`=%s".format(
        scope_info["column"], scope_info["table"], scope_info["where"]
    )
    row = fetch_one(sql, (target_id,), scope_info["database"])
    if not row:
        raise HTTPException(status_code=404, detail="Inventory row not found")
    return character, scope, scope_info, target_id, row["item_blob"]


def _write_inventory_blob(scope_info, target_id, blob):
    sql = "UPDATE `{0}` SET `{1}`=%s WHERE `{2}`=%s".format(
        scope_info["table"], scope_info["column"], scope_info["where"]
    )
    execute(sql, (blob, target_id), scope_info["database"])


def _build_deleted_inventory_blob(delete_slots, origin_blob):
    if not origin_blob:
        raise HTTPException(status_code=404, detail="Inventory blob is empty")
    if isinstance(origin_blob, str):
        origin_blob = origin_blob.encode("latin1", "replace")
    try:
        items_bytes = bytearray(zlib.decompress(origin_blob[4:]))
    except Exception as exc:
        raise HTTPException(status_code=500, detail="Inventory blob decompress failed") from exc
    slot_count = len(items_bytes) // 61
    for slot in delete_slots:
        slot = int(slot)
        if slot < 0 or slot >= slot_count:
            raise HTTPException(status_code=400, detail="Slot is outside inventory range")
        items_bytes[slot * 61 : slot * 61 + 61] = bytearray(b"\x00" * 61)
    return origin_blob[:4] + zlib.compress(bytes(items_bytes))


def _delete_inventory_slots(charac_no, scope, slots):
    character, scope, scope_info, target_id, blob = _load_inventory_blob(charac_no, scope)
    new_blob = _build_deleted_inventory_blob(slots, blob)
    _write_inventory_blob(scope_info, target_id, new_blob)
    items = _unpack_inventory_blob(new_blob)
    return {
        "character": character,
        "scope": scope,
        "scope_name": scope_info["name"],
        "items": items,
        "item_count": len(items),
        "deleted_slots": [int(slot) for slot in slots],
    }


def _clear_inventory_scope(charac_no, scope):
    _character, _scope, _scope_info, _target_id, blob = _load_inventory_blob(charac_no, scope)
    slot_count = _inventory_slot_count(blob)
    return _delete_inventory_slots(charac_no, scope, range(slot_count))


def _query_inventory(payload):
    scope = (payload.scope or "inventory").strip()
    character, scope, scope_info, _target_id, blob = _load_inventory_blob(payload.charac_no, scope)
    items = _unpack_inventory_blob(blob)
    return {
        "character": character,
        "scope": scope,
        "scope_name": scope_info["name"],
        "items": items,
        "item_count": len(items),
    }


def _avatar_hidden_options():
    raw = get_pvf_data("avatar_hidden", [[], []]) or [[], []]
    upper = raw[0] if isinstance(raw, list) and raw else []
    options = [{"hidden_option": 0, "name": "None"}]
    for index, name in enumerate(upper):
        name = str(name)
        options.append(
            {
                "hidden_option": index + 1,
                "name": AVATAR_HIDDEN_MAP.get(name, name),
            }
        )
    return options


def _avatar_hidden_name(value):
    value = _normalize_int(value)
    for option in _avatar_hidden_options():
        if option["hidden_option"] == value:
            return option["name"]
    return str(value)


def _query_avatar_items(charac_no):
    character = _get_character(charac_no)
    rows = fetch_all(
        """
        SELECT ui_id, it_id, hidden_option
        FROM user_items
        WHERE charac_no=%s
        ORDER BY ui_id ASC
        """,
        (charac_no,),
        "taiwan_cain_2nd",
    )
    names = _item_name_map([row["it_id"] for row in rows])
    items = []
    for row in rows:
        hidden_option = _normalize_int(row["hidden_option"])
        item_id = int(row["it_id"] or 0)
        items.append(
            {
                "ui_id": int(row["ui_id"]),
                "item_id": item_id,
                "item_name": names.get(item_id, ""),
                "hidden_option": hidden_option,
                "hidden_name": _avatar_hidden_name(hidden_option),
            }
        )
    return {"character": character, "items": items, "item_count": len(items)}


def _set_avatar_hidden(charac_no, ui_ids, hidden_option):
    if not ui_ids:
        raise HTTPException(status_code=400, detail="No avatar selected")
    valid_options = {option["hidden_option"] for option in _avatar_hidden_options()}
    if valid_options and hidden_option not in valid_options:
        raise HTTPException(status_code=400, detail="Invalid hidden option")
    placeholders = ", ".join(["%s"] * len(ui_ids))
    ui_id_values = [int(ui_id) for ui_id in ui_ids]
    rows = fetch_all(
        "SELECT ui_id FROM user_items WHERE charac_no=%s AND ui_id IN ({0})".format(placeholders),
        tuple([int(charac_no)] + ui_id_values),
        "taiwan_cain_2nd",
    )
    matched_ui_ids = {int(row["ui_id"]) for row in rows}
    if matched_ui_ids != set(ui_id_values):
        raise HTTPException(status_code=400, detail="Selected avatar does not belong to character")
    args = [int(hidden_option), int(charac_no)] + ui_id_values
    execute(
        "UPDATE user_items SET hidden_option=%s WHERE charac_no=%s AND ui_id IN ({0})".format(placeholders),
        tuple(args),
        "taiwan_cain_2nd",
    )
    return {"updated": len(matched_ui_ids), "hidden_option": hidden_option, "hidden_name": _avatar_hidden_name(hidden_option)}


def _event_options():
    rows = game_fetch_all(
        """
        SELECT event_id, event_name, event_explain
        FROM dnf_event_info
        ORDER BY event_id ASC
        """
    )
    return [
        {
            "event_id": int(row["event_id"]),
            "event_name": str(row["event_name"] or ""),
            "event_explain": str(row["event_explain"] or ""),
        }
        for row in rows
    ]


def _running_events():
    rows = game_fetch_all(
        """
        SELECT log_id, event_type, parameter1, parameter2
        FROM dnf_event_log
        ORDER BY log_id ASC
        """
    )
    options = {item["event_id"]: item for item in _event_options()}
    result = []
    for row in rows:
        event_id = int(row["event_type"])
        option = options.get(event_id, {})
        result.append(
            {
                "log_id": int(row["log_id"]),
                "event_id": event_id,
                "event_name": option.get("event_name", ""),
                "event_explain": option.get("event_explain", "explain"),
                "parameter1": int(row["parameter1"] or 0),
                "parameter2": int(row["parameter2"] or 0),
            }
        )
    return result


def _add_event(payload):
    game_execute(
        """
        INSERT INTO dnf_event_log(
            occ_time, event_type, parameter1, parameter2, server_id,
            event_flag, start_time, end_time
        )
        VALUES(0, %s, %s, %s, 0, 0, 0, 0)
        """,
        (payload.event_id, payload.parameter1, payload.parameter2),
    )
    return _running_events()


def _delete_event(log_id):
    game_execute("DELETE FROM dnf_event_log WHERE log_id=%s", (log_id,))
    return _running_events()


def _mail_character_rows(user, keyword="", page=1, limit=12, include_deleted=False, target_uid=None):
    page = max(1, int(page or 1))
    limit = max(1, min(int(limit or 12), 500))
    offset = (page - 1) * limit
    keyword = (keyword or "").strip()
    if keyword and not keyword.isdigit():
        return {
            "characters": [],
            "page": page,
            "limit": limit,
            "total": 0,
        }

    clauses = []
    args = []
    count_args = []
    if not include_deleted:
        clauses.append("c.delete_flag=0")
    if user.get("user_type") != "admin":
        clauses.append("c.m_id=%s")
        args.append(int(user["uid"]))
        count_args.append(int(user["uid"]))
    elif target_uid is not None:
        clauses.append("c.m_id=%s")
        args.append(int(target_uid))
        count_args.append(int(target_uid))
    if keyword:
        clauses.append("c.charac_no=%s")
        args.append(int(keyword))
        count_args.append(int(keyword))

    where_sql = " AND ".join(clauses) if clauses else "1=1"
    with mysql_cursor("taiwan_cain", charset="latin1") as cursor:
        cursor.execute(
            "SELECT COUNT(*) AS total FROM charac_info c JOIN pvp_result p ON p.charac_no=c.charac_no WHERE {0}".format(where_sql),
            tuple(count_args),
        )
        count_row = cursor.fetchone()
        cursor.execute(
            """
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
            WHERE {0}
            ORDER BY c.charac_no DESC
            LIMIT %s OFFSET %s
            """.format(where_sql),
            tuple(args + [limit, offset]),
        )
        rows = list(cursor.fetchall())
    characters = [_character_response_row(row) for row in rows]
    return {
        "characters": characters,
        "page": page,
        "limit": limit,
        "total": int(count_row["total"] or 0) if count_row else 0,
    }


def _active_character_nos():
    with mysql_cursor("taiwan_cain", charset="latin1") as cursor:
        cursor.execute(
            """
            SELECT charac_no
            FROM charac_info
            WHERE delete_flag=0
            ORDER BY charac_no ASC
            """
        )
        return [int(row["charac_no"]) for row in cursor.fetchall()]


def _send_mail_to_character_cursor(
    cursor,
    charac_no,
    sender,
    message,
    item_id,
    item_count,
    gold,
    item_type="",
    item_grade=0,
    enhancement_level=0,
    forge_level=0,
    amplify_option=0,
    amplify_value=0,
):
    if item_id <= 0 and gold <= 0 and not message:
        raise HTTPException(status_code=400, detail="Message, item, or gold is required")
    letter_id = _insert_mail_message(cursor, charac_no, sender, message)
    postal_count = 0
    if item_id > 0 or gold > 0:
        postal_count = _insert_mail_postal(
            cursor,
            charac_no,
            letter_id,
            sender,
            message,
            item_id,
            item_count,
            gold,
            item_type,
            item_grade,
            enhancement_level,
            forge_level,
            amplify_option,
            amplify_value,
        )
    return letter_id, postal_count


def _send_mail_to_character(
    charac_no,
    sender,
    message,
    item_id,
    item_count,
    gold,
    item_type="",
    item_grade=0,
    enhancement_level=0,
    forge_level=0,
    amplify_option=0,
    amplify_value=0,
):
    with mysql_cursor("taiwan_cain_2nd", commit=True, charset="latin1") as cursor:
        return _send_mail_to_character_cursor(
            cursor,
            charac_no,
            sender,
            message,
            item_id,
            item_count,
            gold,
            item_type,
            item_grade,
            enhancement_level,
            forge_level,
            amplify_option,
            amplify_value,
        )


def _delete_mail_for_character(charac_no):
    with mysql_cursor("taiwan_cain_2nd", commit=True) as cursor:
        return cursor.execute(
            "UPDATE taiwan_cain_2nd.postal SET delete_flag=1 WHERE receive_charac_no=%s AND delete_flag=0",
            (charac_no,),
        )


def _delete_mail_for_all_characters():
    with mysql_cursor("taiwan_cain_2nd", commit=True) as cursor:
        return cursor.execute("UPDATE taiwan_cain_2nd.postal SET delete_flag=1 WHERE delete_flag=0")


app = FastAPI(title="DNF Desktop Launcher API")
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origin_list,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.on_event("startup")
def startup() -> None:
    ensure_permission_tables()
    ensure_pvf_tables()


@app.get("/health")
def health() -> dict:
    return {
        "ok": True,
        "server_host": settings.server_host,
        "game_db_name": settings.game_db_name,
        "tool_db_name": settings.tool_db_name,
    }


DEFAULT_HOME_TITLE = "欢迎回来，勇士"
DEFAULT_HOME_EYEBROW = "冒险准备完成"
DEFAULT_HOME_ANNOUNCEMENTS = [
    {
        "title": "版本更新公告标题占位",
        "summary": "后续可替换为具体版本更新内容",
        "content": "这里用于展示版本更新全文，管理员可在权限管理页修改。",
        "poster_url": "/api/posters/sample-1",
    },
    {
        "title": "客户端下载说明占位",
        "summary": "后续可替换为客户端下载说明",
        "content": "这里用于展示客户端下载说明全文，管理员可在权限管理页修改。",
        "poster_url": "/api/posters/sample-2",
    },
    {
        "title": "活动与维护通知占位",
        "summary": "后续可替换为活动、维护或补偿通知",
        "content": "这里用于展示活动、维护或补偿通知全文，管理员可在权限管理页修改。",
        "poster_url": "/api/posters/sample-3",
    },
]


def _home_announcements():
    raw = get_setting("home_announcements", "")
    if not raw:
        return DEFAULT_HOME_ANNOUNCEMENTS[:]
    try:
        data = json.loads(raw)
    except Exception:
        return DEFAULT_HOME_ANNOUNCEMENTS[:]
    if not isinstance(data, list):
        return DEFAULT_HOME_ANNOUNCEMENTS[:]
    announcements = []
    for item in data[:8]:
        if not isinstance(item, dict):
            continue
        title = str(item.get("title") or "").strip()
        if not title:
            continue
        announcements.append(
            {
                "title": title,
                "summary": str(item.get("summary") or "").strip(),
                "content": str(item.get("content") or "").strip(),
                "poster_url": str(item.get("poster_url") or "").strip(),
            }
        )
    return announcements or DEFAULT_HOME_ANNOUNCEMENTS[:]


def _home_settings():
    return {
        "home_title": get_setting("home_title", DEFAULT_HOME_TITLE) or DEFAULT_HOME_TITLE,
        "home_eyebrow": get_setting("home_eyebrow", DEFAULT_HOME_EYEBROW) or DEFAULT_HOME_EYEBROW,
        "client_download_url": get_setting("client_download_url", "") or "",
        "announcements": _home_announcements(),
    }


@app.get("/api/settings")
def public_settings() -> dict:
    return {"home": _home_settings()}


POSTER_DIRECTORY = Path(__file__).resolve().parent.parent / "data" / "posters"
POSTER_EXTENSIONS = (".jpg", ".jpeg", ".png", ".webp")


@app.get("/api/posters/{filename}")
def poster_image(filename: str):
    safe_name = Path(filename).name
    suffix = Path(safe_name).suffix.lower()
    if safe_name != filename or (suffix and suffix not in POSTER_EXTENSIONS):
        raise HTTPException(status_code=404, detail="Poster not found")
    if suffix:
        path = POSTER_DIRECTORY / safe_name
    else:
        path = next(
            (
                POSTER_DIRECTORY / f"{safe_name}{extension}"
                for extension in POSTER_EXTENSIONS
                if (POSTER_DIRECTORY / f"{safe_name}{extension}").is_file()
            ),
            POSTER_DIRECTORY / safe_name,
        )
    if not path.is_file():
        raise HTTPException(status_code=404, detail="Poster not found")
    return FileResponse(str(path))


@app.post("/api/auth/login")
@transactional
def login(payload: LoginRequest, request: Request) -> dict:
    admin = verify_admin(payload.account_name, payload.password)
    if admin:
        write_audit_log(request, admin["uid"], "auth.admin_login", "admin login")
        return {
            "access_token": create_session_token(
                admin["uid"], admin["accountname"], "admin"
            ),
            "token_type": "bearer",
            "uid": admin["uid"],
            "account_name": admin["accountname"],
            "user_type": "admin",
            "permissions": ALL_PERMISSIONS,
            "can_launch": False,
            "tools": visible_tools(ALL_PERMISSIONS),
        }

    account = authenticate_account(payload.account_name, payload.password)
    uid = account["uid"]
    write_audit_log(request, uid, "auth.login", "desktop login")
    permissions = get_account_permissions(uid)
    return {
        "access_token": create_session_token(uid, account["account_name"]),
        "token_type": "bearer",
        "uid": uid,
        "account_name": account["account_name"],
        "user_type": "game",
        "permissions": permissions,
        "can_launch": True,
        "tools": visible_tools(permissions),
    }


@app.post("/api/auth/register")
def register(payload: RegisterRequest) -> dict:
    if not _is_ascii(payload.account_name):
        raise HTTPException(status_code=400, detail="Account name must use ASCII characters")
    if not _is_digits(payload.account_name):
        raise HTTPException(status_code=400, detail="Account name must contain digits only")
    if payload.qq and not payload.qq.isdigit():
        raise HTTPException(status_code=400, detail="QQ must contain digits only")
    register_account(payload.account_name, payload.password, payload.qq)
    return {"ok": True, "message": "register success", "account_name": payload.account_name}


@app.post("/api/auth/change-password")
@transactional
def change_password(payload: ChangePasswordRequest, request: Request) -> dict:
    account_name = payload.account_name.strip()
    if not _is_digits(account_name):
        raise HTTPException(status_code=400, detail="Account name must contain digits only")
    if not payload.qq or not payload.qq.isdigit():
        raise HTTPException(status_code=400, detail="QQ must contain digits only")
    _check_password_reset_rate(request, account_name)

    row = game_fetch_one(
        "SELECT uid FROM accounts WHERE accountname=%s AND qq=%s",
        (account_name, payload.qq),
    )
    if not row:
        raise HTTPException(status_code=401, detail="Account or QQ is incorrect")
    uid = int(row["uid"])
    game_execute(
        "UPDATE accounts SET password=%s WHERE uid=%s",
        (hashlib.md5(payload.new_password.encode()).hexdigest(), uid),
    )

    write_audit_log(
        request,
        uid,
        "auth.password.reset",
        "game account password reset by account and QQ",
    )
    return {"ok": True}


@app.post("/api/auth/admin/change-password")
@transactional
def change_admin_password_api(
    payload: AdminChangePasswordRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    _require_admin(user, "change admin password")
    if not update_admin_password(
        user["uid"], payload.current_password, payload.new_password
    ):
        raise HTTPException(status_code=401, detail="Current password is incorrect")
    write_audit_log(
        request, user["uid"], "auth.admin_password.change", "admin password changed"
    )
    return {"ok": True}


@app.get("/api/pvf/status")
def pvf_status(user: dict = Depends(get_current_user)) -> dict:
    active = get_active_pvf()
    if not active:
        return {"loaded": False, "client_pvf_md5": _client_pvf_md5()}
    return {
        "loaded": True,
        "md5": active["md5"],
        "pvf_path": active["pvf_path"],
        "encode": active["encode_name"],
        "file_size": int(active["file_size"]),
        "stackable_count": int(active["stackable_count"]),
        "equipment_count": int(active["equipment_count"]),
        "exp_level_count": int(active["exp_level_count"]),
        "client_pvf_md5": _client_pvf_md5(),
        "updated_at": _format_datetime(active["updated_at"]),
    }


@app.get("/api/pvf/items")
def pvf_items(
    keyword: str = "",
    item_id: Optional[int] = None,
    item_type: str = "",
    limit: int = 50,
    page: int = 1,
    user: dict = Depends(get_current_user),
) -> dict:
    return search_items(keyword, item_id, item_type, limit, page)


@app.get("/api/gm/characters")
def gm_characters(
    keyword: str = "",
    page: int = 1,
    limit: int = 12,
    include_deleted: bool = False,
    uid: Optional[int] = None,
    user: dict = Depends(get_current_user),
) -> dict:
    allowed = {"gm.mail", "gm.inventory", "gm.avatar.edit", "gm.character.edit"}
    if user.get("user_type") != "admin" and not (allowed & set(user.get("permissions", []))):
        raise HTTPException(status_code=403, detail="Missing character permission")
    if include_deleted and user.get("user_type") != "admin":
        require_permission(user, "gm.character.edit")
    return _mail_character_rows(user, keyword, page, limit, include_deleted, uid)


@app.post("/api/gm/mail/send")
@transactional
def send_mail(
    payload: MailSendRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.mail")
    character = _get_accessible_character(user, payload.charac_no)
    sender = user["account_name"]
    message = payload.message.strip()
    item_id = payload.item_id or 0
    _validate_mail_item_count(
        item_id,
        payload.item_count,
        payload.item_type,
    )
    letter_id, postal_count = _send_mail_to_character(
        character["charac_no"],
        sender,
        message,
        item_id,
        payload.item_count,
        payload.gold,
        payload.item_type,
        payload.item_grade,
        payload.enhancement_level,
        payload.forge_level,
        payload.amplify_option,
        payload.amplify_value,
    )
    write_audit_log(
        request,
        user["uid"],
        "gm.mail.send",
        "charac_no={0}; item_id={1}; item_count={2}; gold={3}; item_type={4}; item_grade={5}; enhancement_level={6}; forge_level={7}; amplify_option={8}; amplify_value={9}; letter_id={10}; postal_count={11}".format(
            character["charac_no"],
            item_id,
            payload.item_count,
            payload.gold,
            payload.item_type,
            1 if (payload.item_type or "").strip().lower() == "equipment" else payload.item_grade,
            payload.enhancement_level,
            payload.forge_level,
            payload.amplify_option,
            payload.amplify_value,
            letter_id,
            postal_count,
        ),
    )
    return {
        "ok": True,
        "letter_id": letter_id,
        "postal_count": postal_count,
        "character": character,
    }


@app.post("/api/gm/mail/send-all")
@transactional
def send_mail_all(
    payload: MailMassSendRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    _require_admin(user, "send mail to all characters")
    sender = user["account_name"]
    message = payload.message.strip()
    item_id = payload.item_id or 0
    if item_id <= 0 and payload.gold <= 0 and not message:
        raise HTTPException(status_code=400, detail="Message, item, or gold is required")
    charac_nos = _active_character_nos()
    _validate_mail_item_count(
        item_id,
        payload.item_count,
        payload.item_type,
    )
    sent_count = 0
    postal_count = 0
    first_letter_id = 0
    with mysql_cursor("taiwan_cain_2nd", commit=True, charset="latin1") as cursor:
        for charac_no in charac_nos:
            letter_id, current_postal_count = _send_mail_to_character_cursor(
                cursor,
                charac_no,
                sender,
                message,
                item_id,
                payload.item_count,
                payload.gold,
                payload.item_type,
                payload.item_grade,
                payload.enhancement_level,
                payload.forge_level,
                payload.amplify_option,
                payload.amplify_value,
            )
            if not first_letter_id:
                first_letter_id = letter_id
            sent_count += 1
            postal_count += current_postal_count
    write_audit_log(
        request,
        user["uid"],
        "gm.mail.send_all",
        "targets={0}; item_id={1}; item_count={2}; gold={3}; first_letter_id={4}; postal_count={5}".format(
            sent_count,
            item_id,
            payload.item_count,
            payload.gold,
            first_letter_id,
            postal_count,
        ),
    )
    return {
        "ok": True,
        "target_count": sent_count,
        "first_letter_id": first_letter_id,
        "postal_count": postal_count,
    }


@app.post("/api/gm/mail/delete")
@transactional
def delete_mail(
    payload: MailDeleteRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.mail")
    character = _get_accessible_character(user, payload.charac_no)
    deleted_count = _delete_mail_for_character(character["charac_no"])
    write_audit_log(
        request,
        user["uid"],
        "gm.mail.delete",
        "charac_no={0}; deleted_count={1}".format(
            character["charac_no"],
            deleted_count,
        ),
    )
    return {"ok": True, "deleted_count": deleted_count, "character": character}


@app.post("/api/gm/mail/delete-all")
@transactional
def delete_mail_all(request: Request, user: dict = Depends(get_current_user)) -> dict:
    _require_admin(user, "delete mail for all characters")
    deleted_count = _delete_mail_for_all_characters()
    write_audit_log(
        request, user["uid"], "gm.mail.delete_all", "deleted_count={0}".format(deleted_count)
    )
    return {"ok": True, "deleted_count": deleted_count}


@app.post("/api/admin/pvf/refresh")
@transactional
def admin_refresh_pvf(
    payload: PvfRefreshRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    if user.get("user_type") != "admin":
        raise HTTPException(status_code=403, detail="Only admin can refresh PVF")
    try:
        result = refresh_pvf_cache(payload.pvf_path, payload.encode, user["uid"])
    except FileNotFoundError as exc:
        raise HTTPException(status_code=404, detail=str(exc))
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc))
    except Exception as exc:
        detail = str(exc).strip() or repr(exc)
        raise HTTPException(
            status_code=500,
            detail="PVF refresh failed: {0}: {1}".format(exc.__class__.__name__, detail),
        )
    result["client_pvf_md5"] = _client_pvf_md5()
    write_audit_log(
        request,
        user["uid"],
        "admin.pvf.refresh",
        "md5={0}; path={1}; stackable={2}; equipment={3}; exp_levels={4}".format(
            result["md5"],
            result["path"],
            result["stackable_count"],
            result["equipment_count"],
            result["exp_level_count"],
        ),
    )
    return result


@app.put("/api/admin/pvf/client-md5")
@transactional
def update_client_pvf_md5(
    payload: PvfClientMd5Request,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    if user.get("user_type") != "admin":
        raise HTTPException(status_code=403, detail="Only admin can update PVF settings")
    client_pvf_md5 = _normalize_md5(payload.client_pvf_md5)
    set_setting("client_pvf_md5", client_pvf_md5)
    write_audit_log(
        request,
        user["uid"],
        "admin.pvf.client_md5",
        "client_pvf_md5={0}".format(client_pvf_md5),
    )
    return {"client_pvf_md5": client_pvf_md5}


@app.post("/api/gm/cera/query")
def query_cera(payload: CeraQueryRequest, user: dict = Depends(get_current_user)) -> dict:
    require_permission(user, "gm.cera.charge")
    account = _resolve_accessible_account(user, payload.uid, payload.account_name)
    return _cera_response(account)


@app.post("/api/gm/cera/charge")
@transactional
def charge_cera(
    payload: CeraChargeRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.cera.charge")
    cera_type = _normalize_cera_type(payload.cera_type)
    action = _normalize_cera_action(payload.action)
    if action == "add" and payload.amount <= 0:
        raise HTTPException(status_code=400, detail="Amount must be greater than 0")

    account = _resolve_accessible_account(user, payload.uid, payload.account_name)
    _ensure_cera_rows(account["uid"])
    if cera_type == "cera":
        if action == "add":
            execute(
                "UPDATE cash_cera SET cera=cera+%s, mod_date=NOW() WHERE account=%s",
                (payload.amount, account["uid"]),
                "taiwan_billing",
            )
        else:
            execute(
                "UPDATE cash_cera SET cera=%s, mod_date=NOW() WHERE account=%s",
                (payload.amount, account["uid"]),
                "taiwan_billing",
            )
    else:
        if action == "add":
            execute(
                "UPDATE cash_cera_point SET cera_point=cera_point+%s, mod_date=NOW() WHERE account=%s",
                (payload.amount, account["uid"]),
                "taiwan_billing",
            )
        else:
            execute(
                "UPDATE cash_cera_point SET cera_point=%s, mod_date=NOW() WHERE account=%s",
                (payload.amount, account["uid"]),
                "taiwan_billing",
            )

    result = _cera_response(account)
    write_audit_log(
        request,
        user["uid"],
        "gm.cera.charge",
        "target_uid={0}; type={1}; action={2}; amount={3}; cera={4}; cera_point={5}".format(
            account["uid"],
            cera_type,
            action,
            payload.amount,
            result["cera"],
            result["cera_point"],
        ),
    )
    return result


@app.post("/api/gm/account/resolve")
def resolve_gm_account(payload: AccountResolveRequest, user: dict = Depends(get_current_user)) -> dict:
    _require_admin(user, "resolve account")
    return _resolve_gm_account(payload.uid, payload.account_name)


@app.get("/api/gm/character/job-options")
def character_job_options(user: dict = Depends(get_current_user)) -> dict:
    require_permission(user, "gm.character.edit")
    return _character_job_options()


@app.post("/api/gm/character/level")
@transactional
def set_character_level(
    payload: CharacterLevelRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.character.edit")
    before = _get_accessible_character(user, payload.charac_no)
    if before["level"] != payload.level:
        try:
            exp = get_exp_for_level(payload.level)
        except ValueError as exc:
            raise HTTPException(status_code=400, detail=str(exc))
        except RuntimeError as exc:
            raise HTTPException(status_code=500, detail=str(exc))
        execute(
            "UPDATE charac_stat SET exp=%s WHERE charac_no=%s",
            (exp, payload.charac_no),
            "taiwan_cain",
        )
        execute(
            "UPDATE charac_info SET lev=%s WHERE charac_no=%s",
            (payload.level, payload.charac_no),
            "taiwan_cain",
        )
    after = _get_character(payload.charac_no)
    write_audit_log(
        request,
        user["uid"],
        "gm.character.level",
        "target_uid={0}; charac_no={1}; level={2}->{3}".format(
            before["uid"],
            payload.charac_no,
            before["level"],
            after["level"],
        ),
    )
    return {"character": after}


@app.post("/api/gm/character/pvp-grade")
@transactional
def set_character_pvp_grade(
    payload: CharacterPvpGradeRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.character.edit")
    _get_accessible_character(user, payload.charac_no)
    before, after = _set_pvp_grade(payload.charac_no, payload.pvp_grade)
    write_audit_log(
        request,
        user["uid"],
        "gm.character.pvp_grade",
        "target_uid={0}; charac_no={1}; pvp_grade={2}->{3}".format(
            before["uid"],
            payload.charac_no,
            before["pvp_grade"],
            after["pvp_grade"],
        ),
    )
    return {"character": after}


@app.post("/api/gm/character/pvp-point")
@transactional
def set_character_pvp_point(
    payload: CharacterPvpPointRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.character.edit")
    _get_accessible_character(user, payload.charac_no)
    before, after = _set_pvp_point(payload.charac_no, payload.pvp_point)
    write_audit_log(
        request,
        user["uid"],
        "gm.character.pvp_point",
        "target_uid={0}; charac_no={1}; pvp_point={2}->{3}; win_point={4}->{5}".format(
            before["uid"],
            payload.charac_no,
            before["pvp_point"],
            after["pvp_point"],
            before["win_point"],
            after["win_point"],
        ),
    )
    return {"character": after}


@app.post("/api/gm/character/job")
@transactional
def set_character_job(
    payload: CharacterJobRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.character.edit")
    _get_accessible_character(user, payload.charac_no)
    before, after = _set_character_job(payload)
    write_audit_log(
        request,
        user["uid"],
        "gm.character.job",
        "target_uid={0}; charac_no={1}; job={2}->{3}; grow_type={4}->{5}; expert_job={6}->{7}".format(
            before["uid"],
            payload.charac_no,
            before["job"],
            after["job"],
            before["grow_type"],
            after["grow_type"],
            before["expert_job"],
            after["expert_job"],
        ),
    )
    return {"character": after}


@app.post("/api/gm/character/delete")
@transactional
def delete_character(
    payload: CharacterVisibilityRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.character.edit")
    _get_accessible_character(user, payload.charac_no, allow_deleted=True)
    before, after = _set_character_delete_flag(payload.charac_no, 1)
    write_audit_log(
        request,
        user["uid"],
        "gm.character.delete",
        "target_uid={0}; charac_no={1}; delete_flag={2}->{3}".format(
            before["uid"],
            payload.charac_no,
            before["delete_flag"],
            after["delete_flag"],
        ),
    )
    return {"character": after}


@app.post("/api/gm/character/recover")
@transactional
def recover_character(
    payload: CharacterVisibilityRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.character.edit")
    _get_accessible_character(user, payload.charac_no, allow_deleted=True)
    before, after = _set_character_delete_flag(payload.charac_no, 0)
    write_audit_log(
        request,
        user["uid"],
        "gm.character.recover",
        "target_uid={0}; charac_no={1}; delete_flag={2}->{3}".format(
            before["uid"],
            payload.charac_no,
            before["delete_flag"],
            after["delete_flag"],
        ),
    )
    return {"character": after}


@app.post("/api/gm/inventory/query")
def query_inventory(payload: InventoryQueryRequest, user: dict = Depends(get_current_user)) -> dict:
    require_permission(user, "gm.inventory")
    _get_accessible_character(user, payload.charac_no)
    return _query_inventory(payload)


@app.post("/api/gm/inventory/delete")
@transactional
def delete_inventory_item(
    payload: InventoryDeleteRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.inventory")
    _get_accessible_character(user, payload.charac_no)
    result = _delete_inventory_slots(payload.charac_no, payload.scope, [payload.slot])
    write_audit_log(
        request,
        user["uid"],
        "gm.inventory.delete",
        "charac_no={0}; scope={1}; slot={2}".format(
            payload.charac_no, result["scope"], payload.slot
        ),
    )
    return result


@app.post("/api/gm/inventory/clear")
@transactional
def clear_inventory_scope(
    payload: InventoryClearRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.inventory")
    _get_accessible_character(user, payload.charac_no)
    result = _clear_inventory_scope(payload.charac_no, payload.scope)
    write_audit_log(
        request,
        user["uid"],
        "gm.inventory.clear",
        "charac_no={0}; scope={1}; deleted_slots={2}".format(
            payload.charac_no, result["scope"], len(result["deleted_slots"])
        ),
    )
    return result


@app.get("/api/gm/avatar/options")
def avatar_options(user: dict = Depends(get_current_user)) -> dict:
    require_permission(user, "gm.avatar.edit")
    return {"options": _avatar_hidden_options()}


@app.post("/api/gm/avatar/query")
def query_avatar(payload: AvatarQueryRequest, user: dict = Depends(get_current_user)) -> dict:
    require_permission(user, "gm.avatar.edit")
    _get_accessible_character(user, payload.charac_no)
    return _query_avatar_items(payload.charac_no)


@app.post("/api/gm/avatar/hidden")
@transactional
def set_avatar_hidden(
    payload: AvatarHiddenRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.avatar.edit")
    _get_accessible_character(user, payload.charac_no)
    result = _set_avatar_hidden(payload.charac_no, payload.ui_ids, payload.hidden_option)
    write_audit_log(
        request,
        user["uid"],
        "gm.avatar.hidden",
        "charac_no={0}; ui_ids={1}; hidden_option={2}".format(
            payload.charac_no,
            ",".join(str(ui_id) for ui_id in payload.ui_ids),
            payload.hidden_option,
        ),
    )
    return result


@app.get("/api/gm/events")
def list_events(user: dict = Depends(get_current_user)) -> dict:
    require_permission(user, "gm.event.manage")
    return {"available": _event_options(), "running": _running_events()}


@app.post("/api/gm/events")
@transactional
def add_event(
    payload: EventAddRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.event.manage")
    running = _add_event(payload)
    write_audit_log(
        request,
        user["uid"],
        "gm.event.add",
        "event_id={0}; parameter1={1}; parameter2={2}".format(
            payload.event_id,
            payload.parameter1,
            payload.parameter2,
        ),
    )
    return {"running": running}


@app.delete("/api/gm/events/{log_id}")
@transactional
def delete_event(
    log_id: int,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    require_permission(user, "gm.event.manage")
    running = _delete_event(log_id)
    write_audit_log(
        request, user["uid"], "gm.event.delete", "log_id={0}".format(log_id)
    )
    return {"running": running}


@app.post("/api/gm/ban/query")
def query_ban(payload: BanQueryRequest, user: dict = Depends(get_current_user)) -> dict:
    _require_admin(user, "query account bans")
    account = _resolve_gm_account(payload.uid, payload.account_name)
    return _ban_status(account)


@app.post("/api/gm/ban/set")
@transactional
def set_ban(
    payload: BanSetRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    _require_admin(user, "set account bans")
    account = _resolve_gm_account(payload.uid, payload.account_name)
    punish_type = _normalize_punish_type(payload.punish_type)
    now = datetime.now()
    start_time = now.strftime("%Y-%m-%d %H:%M:%S")
    end_time = (now + timedelta(days=payload.days)).strftime("%Y-%m-%d %H:%M:%S")
    game_execute("DELETE FROM member_punish_info WHERE m_id=%s", (account["uid"],))
    game_execute(
        """
        REPLACE INTO member_punish_info
            (m_id, punish_type, occ_time, punish_value, apply_flag, start_time, end_time, reason)
        VALUES(%s, %s, %s, %s, 2, %s, %s, %s)
        """,
        (
            account["uid"],
            punish_type,
            start_time,
            101,
            start_time,
            end_time,
            payload.reason.strip(),
        ),
    )
    result = _ban_status(account)
    write_audit_log(
        request,
        user["uid"],
        "gm.ban.set",
        "target_uid={0}; punish_type={1}; days={2}; end_time={3}; reason={4}".format(
            account["uid"],
            punish_type,
            payload.days,
            end_time,
            payload.reason.strip(),
        ),
    )
    return result


@app.post("/api/gm/ban/unban")
@transactional
def unset_ban(
    payload: BanQueryRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    _require_admin(user, "remove account bans")
    account = _resolve_gm_account(payload.uid, payload.account_name)
    game_execute("DELETE FROM member_punish_info WHERE m_id=%s", (account["uid"],))
    result = _ban_status(account)
    write_audit_log(
        request, user["uid"], "gm.ban.unban", "target_uid={0}".format(account["uid"])
    )
    return result


@app.get("/api/admin/permissions")
def list_permissions(user: dict = Depends(get_current_user)) -> dict:
    _require_admin(user, "list permissions")
    return {"permissions": ALL_PERMISSIONS}


@app.get("/api/admin/accounts")
def list_accounts(keyword: str = "", limit: int = 50, user: dict = Depends(get_current_user)) -> dict:
    _require_admin(user, "list accounts")
    if limit < 1:
        limit = 1
    if limit > 100:
        limit = 100
    keyword = (keyword or "").strip()
    if keyword:
        rows = game_fetch_all(
            """
            SELECT uid, accountname
            FROM accounts
            WHERE accountname LIKE %s OR CAST(uid AS CHAR)=%s
            ORDER BY uid DESC
            LIMIT %s
            """,
            ("%{0}%".format(keyword), keyword, limit),
        )
    else:
        rows = game_fetch_all(
            """
            SELECT uid, accountname
            FROM accounts
            ORDER BY uid DESC
            LIMIT %s
            """,
            (limit,),
        )
    accounts = []
    for row in rows:
        uid = int(row["uid"])
        permissions = get_account_permissions(uid)
        accounts.append(
            {
                "uid": uid,
                "account_name": row["accountname"],
                "permissions": permissions,
                "tools": visible_tools(permissions),
            }
        )
    return {"accounts": accounts}


@app.put("/api/admin/accounts/{uid}/permissions")
@transactional
def update_permissions(
    uid: int,
    payload: SetPermissionsRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    _require_admin(user, "update account permissions")
    permissions = set_account_permissions(uid, payload.permissions)
    write_audit_log(
        request,
        user["uid"],
        "admin.permissions.update",
        f"target_uid={uid}; permissions={permissions}",
    )
    return {"uid": uid, "permissions": permissions}


@app.get("/api/admin/logs")
def list_operation_logs(
    keyword: str = "",
    page: int = 1,
    limit: int = 50,
    user: dict = Depends(get_current_user),
) -> dict:
    _require_admin(user, "list operation logs")
    page = max(1, int(page or 1))
    limit = max(1, min(int(limit or 50), 200))
    offset = (page - 1) * limit
    keyword = (keyword or "").strip()
    where_sql = "1=1"
    args = []
    if keyword:
        where_sql = "(action LIKE %s OR detail LIKE %s OR ip LIKE %s OR CAST(uid AS CHAR)=%s)"
        like_keyword = "%{0}%".format(keyword)
        args = [like_keyword, like_keyword, like_keyword, keyword]
    count_row = fetch_one(
        "SELECT COUNT(*) AS total FROM operation_logs WHERE {0}".format(where_sql),
        tuple(args),
    )
    rows = fetch_all(
        """
        SELECT id, uid, action, detail, ip, created_at
        FROM operation_logs
        WHERE {0}
        ORDER BY id DESC
        LIMIT %s OFFSET %s
        """.format(where_sql),
        tuple(args + [limit, offset]),
    )
    return {
        "logs": [
            {
                "id": int(row["id"]),
                "uid": int(row["uid"]),
                "action": row["action"] or "",
                "detail": row["detail"] or "",
                "ip": row["ip"] or "",
                "created_at": _format_datetime(row["created_at"]),
            }
            for row in rows
        ],
        "page": page,
        "limit": limit,
        "total": int(count_row["total"] or 0) if count_row else 0,
    }


@app.put("/api/admin/settings/home")
@transactional
def update_home_settings(
    payload: HomeSettingsRequest,
    request: Request,
    user: dict = Depends(get_current_user),
) -> dict:
    _require_admin(user, "update home settings")
    home_title = payload.home_title.strip() or DEFAULT_HOME_TITLE
    home_eyebrow = payload.home_eyebrow.strip() or DEFAULT_HOME_EYEBROW
    client_download_url = payload.client_download_url.strip()
    if client_download_url and not (
        client_download_url.startswith("http://")
        or client_download_url.startswith("https://")
    ):
        raise HTTPException(
            status_code=400,
            detail="Client download URL must use http:// or https://",
        )
    announcements = []
    for item in payload.announcements[:8]:
        title = item.title.strip()
        if not title:
            continue
        poster_url = item.poster_url.strip()
        if poster_url and not (
            poster_url.startswith("http://")
            or poster_url.startswith("https://")
            or poster_url.startswith("/api/posters/")
        ):
            raise HTTPException(
                status_code=400,
                detail="Poster URL must use http://, https://, or /api/posters/",
            )
        announcements.append(
            {
                "title": title,
                "summary": item.summary.strip(),
                "content": item.content.strip(),
                "poster_url": poster_url,
            }
        )
    if not announcements:
        announcements = DEFAULT_HOME_ANNOUNCEMENTS[:]
    set_setting("home_title", home_title)
    set_setting("home_eyebrow", home_eyebrow)
    set_setting("client_download_url", client_download_url)
    set_setting("home_announcements", json.dumps(announcements, ensure_ascii=False))
    write_audit_log(
        request,
        user["uid"],
        "admin.settings.home",
        "home_title={0}; announcements={1}; client_download_url={2}".format(
            home_title,
            len(announcements),
            "set" if client_download_url else "empty",
        ),
    )
    return {"home": _home_settings()}


@app.post("/api/launcher/direct")
def direct_launch(user: dict = Depends(get_current_user)) -> dict:
    if user.get("user_type") != "game":
        raise HTTPException(status_code=403, detail="Admin cannot launch game directly")
    return create_direct_launch(int(user["uid"]))
