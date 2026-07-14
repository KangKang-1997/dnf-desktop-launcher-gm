from contextlib import contextmanager
from contextvars import ContextVar
from functools import wraps

import pymysql
from pymysql.cursors import DictCursor

from .config import settings


_transaction_connection = ContextVar("transaction_connection", default=None)


@contextmanager
def mysql_cursor(database=None, commit=False, charset=None, use_unicode=None):
    active_conn = _transaction_connection.get()
    if active_conn is not None:
        if database:
            active_conn.select_db(database)
        cursor = active_conn.cursor()
        try:
            yield cursor
        finally:
            cursor.close()
        return

    connect_kwargs = dict(
        host=settings.db_host,
        port=settings.db_port,
        user=settings.db_user,
        password=settings.db_password,
        database=database,
        charset=charset or settings.db_charset,
        connect_timeout=5,
        autocommit=False,
        cursorclass=DictCursor,
    )
    if use_unicode is not None:
        connect_kwargs["use_unicode"] = use_unicode
    conn = pymysql.connect(**connect_kwargs)
    cursor = None
    try:
        cursor = conn.cursor()
        yield cursor
        if commit:
            conn.commit()
    except Exception:
        try:
            conn.rollback()
        except Exception:
            pass
        raise
    finally:
        if cursor is not None:
            cursor.close()
        conn.close()


@contextmanager
def mysql_transaction(database=None):
    if _transaction_connection.get() is not None:
        yield
        return

    connect_kwargs = dict(
        host=settings.db_host,
        port=settings.db_port,
        user=settings.db_user,
        password=settings.db_password,
        database=database or settings.tool_db_name,
        charset=settings.db_charset,
        connect_timeout=5,
        autocommit=False,
        cursorclass=DictCursor,
    )
    conn = pymysql.connect(**connect_kwargs)
    token = _transaction_connection.set(conn)
    try:
        yield
        conn.commit()
    except Exception:
        conn.rollback()
        raise
    finally:
        _transaction_connection.reset(token)
        conn.close()


def transactional(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        with mysql_transaction():
            return func(*args, **kwargs)

    return wrapper


def ensure_database(database):
    with mysql_cursor(None, commit=True) as cursor:
        cursor.execute(
            f"CREATE DATABASE IF NOT EXISTS `{database}` "
            f"DEFAULT CHARACTER SET {settings.db_charset}"
        )


def fetch_one(sql, args=(), database=None):
    with mysql_cursor(database or settings.tool_db_name) as cursor:
        cursor.execute(sql, args)
        return cursor.fetchone()


def fetch_all(sql, args=(), database=None):
    with mysql_cursor(database or settings.tool_db_name) as cursor:
        cursor.execute(sql, args)
        return list(cursor.fetchall())


def execute(sql, args=(), database=None):
    with mysql_cursor(database or settings.tool_db_name, commit=True) as cursor:
        return cursor.execute(sql, args)


def game_fetch_one(sql, args=()):
    return fetch_one(sql, args, settings.game_db_name)


def game_fetch_all(sql, args=()):
    return fetch_all(sql, args, settings.game_db_name)


def game_execute(sql, args=()):
    return execute(sql, args, settings.game_db_name)
