def main() -> int:
    import uvicorn
    from app.core.config import settings
    from app.main import app

    uvicorn.run(
        app,
        host=settings.listen_host,
        port=settings.listen_port,
        reload=False,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
