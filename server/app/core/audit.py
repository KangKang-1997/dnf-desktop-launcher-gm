from .db import execute


def write_audit_log(request, uid, action, detail=""):
    execute(
        "INSERT INTO operation_logs(uid, action, detail, ip) VALUES(%s, %s, %s, %s)",
        (
            int(uid),
            action,
            detail,
            request.client.host if request.client else "",
        ),
    )
