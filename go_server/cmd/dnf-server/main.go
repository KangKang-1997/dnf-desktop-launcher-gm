package main

import (
	"log"
	"os"
	"path/filepath"

	"dnf-launcher-go/internal/config"
	"dnf-launcher-go/internal/httpserver"
)

func main() {
	configPath, err := executableConfigPath()
	if err != nil {
		log.Fatalf("resolve config path: %v", err)
	}
	settings, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	server, err := httpserver.New(settings)
	if err != nil {
		log.Fatalf("initialize server: %v", err)
	}
	log.Printf("dnf launcher go server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func executableConfigPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(executable), "config.json"), nil
}
