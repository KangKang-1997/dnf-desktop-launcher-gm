from fastapi import HTTPException

from ..core.db import game_fetch_one
from .tokens import create_dnf_login_token
from ..core.permissions import get_client_pvf_md5


def create_direct_launch(uid):
    account = game_fetch_one("SELECT accountname FROM accounts WHERE uid=%s", (uid,))
    try:
        dnf_token = create_dnf_login_token(uid)
    except Exception as exc:
        raise HTTPException(
            status_code=500,
            detail="Failed to create DNF login token: {0}".format(exc),
        )
    return {
        "uid": uid,
        "account_name": account["accountname"] if account else "",
        "dnf_token": dnf_token,
        "client_pvf_md5": get_client_pvf_md5(),
    }
