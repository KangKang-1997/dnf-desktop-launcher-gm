import base64
import subprocess

from ..core.config import settings


def _openssl_private_encrypt(data):
    key_path = settings.login_private_key_path
    if not key_path.is_absolute():
        key_path = key_path.resolve()
    if not key_path.exists():
        raise FileNotFoundError("private key not found: {0}".format(key_path))
    proc = subprocess.run(
        [
            "openssl",
            "rsautl",
            "-sign",
            "-inkey",
            str(key_path),
        ],
        input=data,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    if proc.returncode != 0:
        err = proc.stderr.decode(errors="replace").strip()
        raise RuntimeError("openssl rsautl failed: {0}".format(err))
    return base64.b64encode(proc.stdout).decode()


def create_dnf_login_token(uid):
    raw_hex = (
        "%08x010101010101010101010101010101010101010101010101010101010101010155914510010403030101"
        % uid
    )
    return _openssl_private_encrypt(bytes.fromhex(raw_hex))
