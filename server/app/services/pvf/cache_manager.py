import hashlib
import json
import threading
from pathlib import Path

from ...core.config import settings
from ...core.db import execute, fetch_all, fetch_one, mysql_cursor

PVF_DATA_KEYS = [
    "job",
    "avatar_hidden",
    "exp_table",
]

PVF_ENCODINGS = ("big5", "gbk")
_PVF_REFRESH_LOCK = threading.Lock()


def normalize_pvf_encoding(value):
    value = (value or "big5").strip().lower()
    if value == "none":
        value = "big5"
    if value not in PVF_ENCODINGS:
        raise ValueError("PVF encoding must be one of: {0}".format(", ".join(PVF_ENCODINGS)))
    return value


def _load_pvf_reader():
    try:
        from . import pvf_reader
    except ModuleNotFoundError as exc:
        if exc.name == "zhconv":
            raise RuntimeError(
                "Missing dependency zhconv. Install server dependencies once with: "
                "python3 -m pip install -r requirements.txt"
            )
        raise
    return pvf_reader


def ensure_pvf_tables():
    execute(
        """
        CREATE TABLE IF NOT EXISTS pvf_meta (
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
        )
        """
    )
    execute(
        """
        CREATE TABLE IF NOT EXISTS pvf_items (
            pvf_md5 CHAR(32) NOT NULL,
            item_id INT NOT NULL,
            item_type VARCHAR(32) NOT NULL,
            item_name VARCHAR(255) NOT NULL,
            detail_json LONGTEXT NULL,
            PRIMARY KEY (pvf_md5, item_id),
            INDEX idx_pvf_items_name (item_name(64)),
            INDEX idx_pvf_items_type (item_type)
        )
        """
    )
    execute(
        """
        CREATE TABLE IF NOT EXISTS pvf_data (
            pvf_md5 CHAR(32) NOT NULL,
            data_key VARCHAR(64) NOT NULL,
            data_json LONGTEXT NOT NULL,
            PRIMARY KEY (pvf_md5, data_key)
        )
        """
    )


def _json_dumps(value):
    return json.dumps(value, ensure_ascii=False, default=str, separators=(",", ":"))


def _chunked(items, chunk_size):
    for index in range(0, len(items), chunk_size):
        yield items[index : index + chunk_size]


def _hash_file(path):
    digest = hashlib.md5()
    with path.open("rb") as fp:
        while True:
            chunk = fp.read(1024 * 1024)
            if not chunk:
                break
            digest.update(chunk)
    return digest.hexdigest().upper()


def _item_rows(md5, names, details, item_type):
    rows = []
    for raw_id, raw_name in names.items():
        item_id = int(raw_id)
        name = str(raw_name or "").strip()
        rows.append(
            (
                md5,
                item_id,
                item_type,
                name[:255],
                _json_dumps(details.get(raw_id) or details.get(item_id) or {}),
            )
        )
    return rows


def _stack_limit_from_detail(detail):
    stack_limit = detail.get("[stack limit]")
    if isinstance(stack_limit, list) and stack_limit:
        try:
            value = int(stack_limit[0])
        except Exception:
            return None
        return value if value > 0 else None
    return None


def _exception_text(exc):
    text = str(exc).strip()
    if not text:
        text = repr(exc)
    return "{0}: {1}".format(exc.__class__.__name__, text)


def _read_required_pvf_stage(stage_name, func, logs):
    try:
        return func()
    except Exception as exc:
        message = "{0} failed: {1}".format(stage_name, _exception_text(exc))
        logs.append(message)
        raise RuntimeError(message)


def _read_optional_pvf_stage(stage_name, default_value, func, logs):
    try:
        return func()
    except Exception as exc:
        logs.append("{0} skipped: {1}".format(stage_name, _exception_text(exc)))
        return default_value


def _read_pvf_data(pvf_reader, pvf, logs):
    pvf.load_Leafs(["stackable", "character", "etc", "equipment"])

    job, _job_tag = _read_required_pvf_stage(
        "job",
        lambda: pvf_reader.get_Job_Dict2(pvf),
        logs,
    )
    exp_table = _read_required_pvf_stage(
        "exp table",
        lambda: pvf_reader.get_exp_table2(pvf),
        logs,
    )
    equipment_structured, equipment, equipment_detail = _read_required_pvf_stage(
        "equipment",
        lambda: pvf_reader.get_Equipment_Dict3(pvf),
        logs,
    )
    stackable, stackable_detail = _read_required_pvf_stage(
        "stackable",
        lambda: pvf_reader.get_Stackable_dict3(pvf),
        logs,
    )
    avatar_hidden = _read_optional_pvf_stage(
        "avatar hidden",
        [[], []],
        lambda: pvf_reader.get_Hidden_Avatar_List2(pvf),
        logs,
    )

    return {
        "jobDict": job,
        "expTable": exp_table,
        "equipment": equipment,
        "equipmentStructuredDict": equipment_structured,
        "equipment_detail": equipment_detail,
        "stackable": stackable,
        "stackable_detail": stackable_detail,
        "avatarHidden": avatar_hidden,
    }


def _build_pvf_cache(pvf_path, encode_name, logs):
    pvf_reader = _load_pvf_reader()
    encode_name = normalize_pvf_encoding(encode_name)
    path = Path(pvf_path).expanduser()
    if not path.exists() or not path.is_file():
        raise FileNotFoundError("PVF file not found: {0}".format(pvf_path))
    if path.suffix.lower() != ".pvf":
        raise ValueError("PVF path must point to a .pvf file")

    md5 = _hash_file(path)
    pvf = pvf_reader.TinyPVF(
        pvfHeader=pvf_reader.PVFHeader(str(path)),
        encode=encode_name or "big5",
    )
    raw = _read_pvf_data(pvf_reader, pvf, logs)
    if not raw:
        raise RuntimeError("PVF read returned no data")

    stackable = raw.get("stackable", {})
    equipment = raw.get("equipment", {})
    data = {
        "job": raw.get("jobDict", {}),
        "avatar_hidden": raw.get("avatarHidden", [[], []]),
        "exp_table": raw.get("expTable", []),
    }
    items = []
    items.extend(_item_rows(md5, stackable, raw.get("stackable_detail", {}), "stackable"))
    items.extend(_item_rows(md5, equipment, raw.get("equipment_detail", {}), "equipment"))
    return {
        "md5": md5,
        "path": str(path),
        "file_size": path.stat().st_size,
        "encode": encode_name or "big5",
        "stackable_count": len(stackable),
        "equipment_count": len(equipment),
        "exp_level_count": len(data["exp_table"]),
        "items": items,
        "data": data,
    }


def refresh_pvf_cache(pvf_path, encode_name, created_by, log_func=None):
    if not _PVF_REFRESH_LOCK.acquire(blocking=False):
        raise RuntimeError("PVF refresh is already running")
    try:
        return _refresh_pvf_cache_unlocked(pvf_path, encode_name, created_by, log_func)
    finally:
        _PVF_REFRESH_LOCK.release()


def _refresh_pvf_cache_unlocked(pvf_path, encode_name, created_by, log_func=None):
    ensure_pvf_tables()
    pvf_reader = _load_pvf_reader()
    logs = []

    def capture(*args, **kwargs):
        message = " ".join(str(arg) for arg in args)
        logs.append(message)
        if log_func:
            log_func(message)

    pvf_reader.logFunc.append(capture)
    try:
        cache = _build_pvf_cache(pvf_path, encode_name, logs)
    finally:
        try:
            pvf_reader.logFunc.remove(capture)
        except ValueError:
            pass

    with mysql_cursor(settings.tool_db_name, commit=True) as cursor:
        cursor.execute("SELECT GET_LOCK(%s, 30) AS acquired", ("launcher:pvf_refresh",))
        lock_row = cursor.fetchone()
        if not lock_row or int(lock_row["acquired"] or 0) != 1:
            raise RuntimeError("PVF cache writer is busy, try again later")
        cursor.execute("DELETE FROM pvf_items WHERE pvf_md5=%s", (cache["md5"],))
        cursor.execute("DELETE FROM pvf_data WHERE pvf_md5=%s", (cache["md5"],))
        item_sql = """
            INSERT INTO pvf_items
                (pvf_md5, item_id, item_type, item_name, detail_json)
            VALUES(%s, %s, %s, %s, %s)
        """
        for rows in _chunked(cache["items"], 500):
            cursor.executemany(item_sql, rows)

        data_rows = [
            (cache["md5"], key, _json_dumps(value))
            for key, value in cache["data"].items()
        ]
        cursor.executemany(
            """
            INSERT INTO pvf_data(pvf_md5, data_key, data_json)
            VALUES(%s, %s, %s)
            """,
            data_rows,
        )
        cursor.execute("UPDATE pvf_meta SET active=0")
        cursor.execute(
            """
            INSERT INTO pvf_meta(
                md5, pvf_path, encode_name, file_size, stackable_count,
                equipment_count, exp_level_count, active, created_by
            )
            VALUES(%s, %s, %s, %s, %s, %s, %s, 1, %s)
            ON DUPLICATE KEY UPDATE
                pvf_path=VALUES(pvf_path),
                encode_name=VALUES(encode_name),
                file_size=VALUES(file_size),
                stackable_count=VALUES(stackable_count),
                equipment_count=VALUES(equipment_count),
                exp_level_count=VALUES(exp_level_count),
                active=1,
                created_by=VALUES(created_by)
            """,
            (
                cache["md5"],
                cache["path"],
                cache["encode"],
                cache["file_size"],
                cache["stackable_count"],
                cache["equipment_count"],
                cache["exp_level_count"],
                created_by,
            ),
        )
    result = {}
    for key in (
        "md5",
        "path",
        "file_size",
        "encode",
        "stackable_count",
        "equipment_count",
        "exp_level_count",
    ):
        result[key] = cache[key]
    result["logs"] = logs[-50:]
    return result


def get_active_pvf():
    ensure_pvf_tables()
    return fetch_one(
        """
        SELECT md5, pvf_path, encode_name, file_size, stackable_count,
               equipment_count, exp_level_count, active, created_by,
               created_at, updated_at
        FROM pvf_meta
        WHERE active=1
        ORDER BY updated_at DESC
        LIMIT 1
        """
    )


def _active_md5():
    active = get_active_pvf()
    if not active:
        return None
    return active["md5"]


def get_pvf_data(data_key, default=None):
    md5 = _active_md5()
    if not md5:
        return default
    row = fetch_one(
        "SELECT data_json FROM pvf_data WHERE pvf_md5=%s AND data_key=%s",
        (md5, data_key),
    )
    if not row:
        return default
    return json.loads(row["data_json"])


def get_exp_for_level(level):
    exp_table = get_pvf_data("exp_table", [])
    if not exp_table:
        raise RuntimeError("PVF exp table is not loaded")
    exp_table_with_lv0 = [0, 0] + exp_table
    if level >= len(exp_table_with_lv0):
        raise ValueError("Level is outside PVF exp table range")
    return int(exp_table_with_lv0[level]) + 1


def search_items(keyword="", item_id=None, item_type="", limit=50, page=1):
    md5 = _active_md5()
    if not md5:
        return {"items": [], "page": 1, "limit": limit, "total": 0}
    if limit < 1:
        limit = 1
    if limit > 100:
        limit = 100
    page = max(1, int(page or 1))
    offset = (page - 1) * limit
    clauses = ["pvf_md5=%s"]
    args = [md5]
    if item_id:
        clauses.append("item_id=%s")
        args.append(item_id)
    if item_type:
        clauses.append("item_type=%s")
        args.append(item_type)
    keyword = (keyword or "").strip()
    if keyword:
        fuzzy = "%{0}%".format(keyword)
        clauses.append("(item_name LIKE %s OR CAST(item_id AS CHAR)=%s)")
        args.extend([fuzzy, keyword])
    where_sql = " AND ".join(clauses)
    count_row = fetch_one(
        "SELECT COUNT(*) AS total FROM pvf_items WHERE {0}".format(where_sql),
        tuple(args),
    )
    query_args = list(args)
    query_args.extend([limit, offset])
    rows = fetch_all(
        """
        SELECT item_id, item_type, item_name, detail_json
        FROM pvf_items
        WHERE {0}
        ORDER BY CHAR_LENGTH(item_name), item_id
        LIMIT %s OFFSET %s
        """.format(where_sql),
        tuple(query_args),
    )
    return {
        "items": [
            {
                "item_id": int(row["item_id"]),
                "item_type": row["item_type"],
                "item_name": row["item_name"],
                "stack_limit": (
                    _stack_limit_from_detail(json.loads(row["detail_json"] or "{}"))
                    if row["item_type"] == "stackable"
                    else None
                ),
            }
            for row in rows
        ],
        "page": page,
        "limit": limit,
        "total": int(count_row["total"] or 0) if count_row else 0,
    }


def get_item_detail(item_id):
    md5 = _active_md5()
    if not md5:
        return {}
    row = fetch_one(
        "SELECT detail_json FROM pvf_items WHERE pvf_md5=%s AND item_id=%s",
        (md5, item_id),
    )
    if not row or not row["detail_json"]:
        return {}
    return json.loads(row["detail_json"])
