import hashlib

import pymysql
from fastapi import HTTPException

from ..core.config import settings
from ..core.db import execute, game_fetch_one, mysql_cursor, transactional


def _password_md5(password):
    return hashlib.md5(password.encode()).hexdigest()


def authenticate_account(account_name, password):
    row = game_fetch_one(
        "SELECT uid, accountname FROM accounts WHERE accountname=%s AND password=%s",
        (account_name, _password_md5(password)),
    )
    if not row:
        raise HTTPException(status_code=401, detail="Invalid account or password")
    return {"uid": int(row["uid"]), "account_name": row["accountname"]}


@transactional
def register_account(account_name, password, qq=""):
    password_md5 = _password_md5(password)
    existing = game_fetch_one(
        "SELECT uid, password FROM accounts WHERE accountname=%s",
        (account_name,),
    )
    if existing:
        if existing["password"] != password_md5:
            raise HTTPException(status_code=400, detail="Account already exists")
        uid = int(existing["uid"])
        _initialize_account(uid)
        return uid

    try:
        execute(
            "INSERT INTO accounts(accountname, password, qq) VALUES(%s, %s, %s)",
            (account_name, password_md5, qq),
            settings.game_db_name,
        )
    except pymysql.err.IntegrityError:
        raise HTTPException(status_code=400, detail="Account already exists")

    row = game_fetch_one(
        "SELECT uid FROM accounts WHERE accountname=%s AND password=%s",
        (account_name, password_md5),
    )
    if not row:
        raise HTTPException(status_code=500, detail="Register failed, please retry")
    uid = int(row["uid"])
    _initialize_account(uid)
    return uid


def _initialize_account(uid):
    lock_name = "launcher:account:init:{0}".format(int(uid))
    with mysql_cursor(settings.game_db_name) as lock_cursor:
        lock_cursor.execute("SELECT GET_LOCK(%s, 10) AS acquired", (lock_name,))
        row = lock_cursor.fetchone()
        if not row or int(row["acquired"] or 0) != 1:
            raise HTTPException(status_code=503, detail="Account initialization is busy")
        try:
            _initialize_account_unlocked(uid)
        finally:
            lock_cursor.execute("SELECT RELEASE_LOCK(%s)", (lock_name,))


def _initialize_account_unlocked(uid):
    statements = [
        (
            settings.game_db_name,
            "SELECT m_id FROM limit_create_character WHERE m_id=%s",
            (uid,),
            "INSERT INTO limit_create_character (m_id) VALUES (%s)",
            (uid,),
        ),
        (
            settings.game_db_name,
            "SELECT m_id FROM member_info WHERE m_id=%s",
            (uid,),
            "INSERT INTO member_info (m_id, user_id) VALUES (%s, %s)",
            (uid, uid),
        ),
        (
            settings.game_db_name,
            "SELECT m_id FROM member_join_info WHERE m_id=%s",
            (uid,),
            "INSERT INTO member_join_info (m_id) VALUES (%s)",
            (uid,),
        ),
        (
            settings.game_db_name,
            "SELECT m_id FROM member_miles WHERE m_id=%s",
            (uid,),
            "INSERT INTO member_miles (m_id) VALUES (%s)",
            (uid,),
        ),
        (
            settings.game_db_name,
            "SELECT m_id FROM member_white_account WHERE m_id=%s",
            (uid,),
            "INSERT INTO member_white_account (m_id) VALUES (%s)",
            (uid,),
        ),
        (
            "taiwan_login",
            "SELECT m_id FROM member_login WHERE m_id=%s",
            (uid,),
            "INSERT INTO member_login (m_id) VALUES (%s)",
            (uid,),
        ),
        (
            "taiwan_cain_2nd",
            "SELECT m_id FROM member_avatar_coin WHERE m_id=%s",
            (uid,),
            "INSERT INTO member_avatar_coin (m_id) VALUES (%s)",
            (uid,),
        ),
    ]
    for database, exists_sql, exists_args, insert_sql, insert_args in statements:
        with mysql_cursor(database) as cursor:
            cursor.execute(exists_sql, exists_args)
            if cursor.fetchone():
                continue
            cursor.execute(insert_sql, insert_args)
