import hashlib
import json

from .config import settings
from .db import ensure_database, execute, fetch_one


ALL_PERMISSIONS = [
    "gm.mail",
    "gm.cera.charge",
    "gm.character.edit",
    "gm.inventory",
    "gm.event.manage",
    "gm.avatar.edit",
]

DEFAULT_PLAYER_PERMISSIONS = []

GM_TOOLS = [
    {"id": "mail", "name": "邮件", "permission": "gm.mail"},
    {"id": "cera_charge", "name": "充值点券", "permission": "gm.cera.charge"},
    {"id": "character_edit", "name": "角色修改", "permission": "gm.character.edit"},
    {"id": "inventory", "name": "背包", "permission": "gm.inventory"},
    {"id": "event_manage", "name": "活动管理", "permission": "gm.event.manage"},
    {"id": "avatar_edit", "name": "时装潜能", "permission": "gm.avatar.edit"},
]

def ensure_permission_tables() -> None:
    ensure_database(settings.tool_db_name)
    execute(
        """
        CREATE TABLE IF NOT EXISTS admins (
            id INT PRIMARY KEY AUTO_INCREMENT,
            username VARCHAR(64) NOT NULL UNIQUE,
            password_md5 CHAR(32) NOT NULL,
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                ON UPDATE CURRENT_TIMESTAMP
        )
        """
    )
    execute(
        """
        CREATE TABLE IF NOT EXISTS roles (
            id INT PRIMARY KEY AUTO_INCREMENT,
            name VARCHAR(64) NOT NULL UNIQUE,
            permissions TEXT NOT NULL,
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                ON UPDATE CURRENT_TIMESTAMP
        )
        """
    )
    execute(
        """
        CREATE TABLE IF NOT EXISTS account_roles (
            uid INT PRIMARY KEY,
            role_id INT NOT NULL,
            updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                ON UPDATE CURRENT_TIMESTAMP
        )
        """
    )
    execute(
        """
        CREATE TABLE IF NOT EXISTS operation_logs (
            id BIGINT PRIMARY KEY AUTO_INCREMENT,
            uid INT NOT NULL,
            action VARCHAR(128) NOT NULL,
            detail TEXT NULL,
            ip VARCHAR(64) NULL,
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        )
        """
    )
    execute(
        """
        CREATE TABLE IF NOT EXISTS settings (
            name VARCHAR(64) PRIMARY KEY,
            value TEXT NOT NULL,
            updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                ON UPDATE CURRENT_TIMESTAMP
        )
        """
    )
    _upsert_role("player", DEFAULT_PLAYER_PERMISSIONS)
    _upsert_role("admin", ALL_PERMISSIONS)
    _ensure_default_admin()


def _ensure_default_admin():
    row = fetch_one("SELECT id FROM admins WHERE username=%s", ("admin",))
    if row:
        return
    password_md5 = hashlib.md5("admin".encode()).hexdigest()
    execute(
        "INSERT INTO admins(username, password_md5) VALUES(%s, %s)",
        ("admin", password_md5),
    )


def verify_admin(username, password):
    row = fetch_one(
        "SELECT id, username FROM admins WHERE username=%s AND password_md5=%s",
        (username, hashlib.md5(password.encode()).hexdigest()),
    )
    if not row:
        return None
    return {"uid": -int(row["id"]), "accountname": row["username"]}


def change_admin_password(uid, current_password, new_password):
    admin_id = abs(int(uid))
    current_md5 = hashlib.md5(current_password.encode()).hexdigest()
    row = fetch_one(
        "SELECT id FROM admins WHERE id=%s AND password_md5=%s",
        (admin_id, current_md5),
    )
    if not row:
        return False
    new_md5 = hashlib.md5(new_password.encode()).hexdigest()
    execute(
        "UPDATE admins SET password_md5=%s WHERE id=%s",
        (new_md5, admin_id),
    )
    return True


def _ensure_setting(name, value):
    row = fetch_one("SELECT name FROM settings WHERE name=%s", (name,))
    if row:
        return
    execute(
        "INSERT INTO settings(name, value) VALUES(%s, %s)",
        (name, value),
    )


def get_setting(name, default=""):
    row = fetch_one("SELECT value FROM settings WHERE name=%s", (name,))
    if not row:
        return default
    return row["value"]


def get_client_pvf_md5():
    return (get_setting("client_pvf_md5", "") or "").strip().upper()


def set_setting(name, value):
    execute(
        """
        INSERT INTO settings(name, value) VALUES(%s, %s)
        ON DUPLICATE KEY UPDATE value=VALUES(value)
        """,
        (name, value),
    )
    return value


def _upsert_role(name, permissions):
    row = fetch_one("SELECT id FROM roles WHERE name=%s", (name,))
    payload = json.dumps(permissions, ensure_ascii=False)
    if row:
        execute("UPDATE roles SET permissions=%s WHERE id=%s", (payload, row["id"]))
        return int(row["id"])
    execute("INSERT INTO roles(name, permissions) VALUES(%s, %s)", (name, payload))
    row = fetch_one("SELECT id FROM roles WHERE name=%s", (name,))
    return int(row["id"])


def get_account_permissions(uid):
    row = fetch_one(
        """
        SELECT r.permissions
        FROM account_roles ar
        JOIN roles r ON r.id = ar.role_id
        WHERE ar.uid=%s
        """,
        (uid,),
    )
    if not row:
        return DEFAULT_PLAYER_PERMISSIONS[:]
    try:
        permissions = json.loads(row["permissions"])
    except Exception:
        return DEFAULT_PLAYER_PERMISSIONS[:]
    return [item for item in permissions if item in ALL_PERMISSIONS]


def set_account_permissions(uid, permissions):
    normalized = [item for item in permissions if item in ALL_PERMISSIONS]
    role_id = _upsert_role("account:{0}".format(uid), normalized)
    execute(
        """
        INSERT INTO account_roles(uid, role_id) VALUES(%s, %s)
        ON DUPLICATE KEY UPDATE role_id=VALUES(role_id)
        """,
        (uid, role_id),
    )
    return normalized


def visible_tools(permissions):
    allowed = set(permissions)
    return [tool for tool in GM_TOOLS if tool["permission"] in allowed]
