import json
from pathlib import Path


DEFAULT_CONFIG = {
    "listen_host": "0.0.0.0",
    "listen_port": 8000,
    "server_host": "127.0.0.1",
    "db_port": 3306,
    "db_user": "数据库用户名",
    "db_password": "数据库密码",
    "game_db_name": "d_taiwan",
    "tool_db_name": "dnf_launcher",
    "db_charset": "utf8",
    "session_secret": "发布时改成随机长字符串",
    "session_ttl_seconds": 86400,
    "cors_origins": ["*"],
    "login_private_key_path": "/opt/server/privatekey.pem",
}


def default_config_path() -> Path:
    return Path.cwd() / "config.json"


def load_config():
    cfg = DEFAULT_CONFIG.copy()
    config_path = default_config_path()
    if config_path.exists():
        loaded = json.loads(config_path.read_text(encoding="utf-8-sig"))
        if not isinstance(loaded, dict):
            raise ValueError(f"Config must be a JSON object: {config_path}")
        cfg.update(loaded)
    return cfg


def _to_origin_list(value):
    if value == "*":
        return ["*"]
    if isinstance(value, str):
        return [item.strip() for item in value.split(",") if item.strip()]
    return [str(item) for item in value]


class Settings:
    def __init__(
        self,
        listen_host,
        listen_port,
        server_host,
        db_port,
        db_user,
        db_password,
        game_db_name,
        tool_db_name,
        db_charset,
        session_secret,
        session_ttl_seconds,
        cors_origins,
        login_private_key_path,
    ):
        self.listen_host = listen_host
        self.listen_port = listen_port
        self.server_host = server_host
        self.db_port = db_port
        self.db_user = db_user
        self.db_password = db_password
        self.game_db_name = game_db_name
        self.tool_db_name = tool_db_name
        self.db_charset = db_charset
        self.session_secret = session_secret
        self.session_ttl_seconds = session_ttl_seconds
        self.cors_origins = cors_origins
        self.login_private_key_path = login_private_key_path

    @property
    def db_host(self):
        return self.server_host

    @property
    def cors_origin_list(self):
        return self.cors_origins


def make_settings():
    cfg = load_config()
    return Settings(
        listen_host=str(cfg["listen_host"]),
        listen_port=int(cfg["listen_port"]),
        server_host=str(cfg["server_host"]),
        db_port=int(cfg["db_port"]),
        db_user=str(cfg["db_user"]),
        db_password=str(cfg["db_password"]),
        game_db_name=str(cfg["game_db_name"]),
        tool_db_name=str(cfg["tool_db_name"]),
        db_charset=str(cfg["db_charset"]),
        session_secret=str(cfg["session_secret"]),
        session_ttl_seconds=int(cfg["session_ttl_seconds"]),
        cors_origins=_to_origin_list(cfg["cors_origins"]),
        login_private_key_path=Path(cfg["login_private_key_path"]),
    )


settings = make_settings()
