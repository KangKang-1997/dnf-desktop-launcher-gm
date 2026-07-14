import base64
import hashlib
import hmac
import json
import time
from typing import Any, Dict, Optional

from fastapi import Header, HTTPException

from .config import settings
from .permissions import ALL_PERMISSIONS, get_account_permissions


def _b64url(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).decode().rstrip("=")


def _unb64url(data: str) -> bytes:
    return base64.urlsafe_b64decode(data + "=" * (-len(data) % 4))


def create_session_token(uid, account_name, user_type="game"):
    payload = {
        "uid": uid,
        "account_name": account_name,
        "user_type": user_type,
        "exp": int(time.time()) + settings.session_ttl_seconds,
    }
    body = _b64url(json.dumps(payload, separators=(",", ":")).encode())
    sig = hmac.new(settings.session_secret.encode(), body.encode(), hashlib.sha256)
    return f"{body}.{_b64url(sig.digest())}"


def parse_session_token(token):
    try:
        body, sig = token.split(".", 1)
        expected = hmac.new(
            settings.session_secret.encode(), body.encode(), hashlib.sha256
        ).digest()
        if not hmac.compare_digest(_unb64url(sig), expected):
            raise ValueError("bad signature")
        payload = json.loads(_unb64url(body))
    except Exception as exc:
        raise HTTPException(status_code=401, detail="Invalid session token") from exc
    if int(payload.get("exp", 0)) < int(time.time()):
        raise HTTPException(status_code=401, detail="Session expired")
    return payload


def get_current_user(authorization=Header(default=None)):
    if not authorization or not authorization.lower().startswith("bearer "):
        raise HTTPException(status_code=401, detail="Missing bearer token")
    payload = parse_session_token(authorization.split(" ", 1)[1].strip())
    uid = int(payload["uid"])
    user_type = payload.get("user_type", "game")
    if user_type == "admin":
        permissions = ALL_PERMISSIONS[:]
    else:
        permissions = get_account_permissions(uid)
    return {
        "uid": uid,
        "account_name": payload["account_name"],
        "user_type": user_type,
        "permissions": permissions,
        "can_launch": user_type == "game",
    }


def require_permission(user, permission):
    if permission not in user["permissions"]:
        raise HTTPException(status_code=403, detail=f"缺少权限：{permission}")
