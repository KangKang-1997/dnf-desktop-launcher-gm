package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"dnf-launcher-go/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

type Store struct {
	settings config.Settings
	tool     *sql.DB
	game     *sql.DB
	mu       sync.Mutex
	named    map[string]*sql.DB
}

func New(settings config.Settings) (*Store, error) {
	tool, err := open(settings, settings.ToolDBName)
	if err != nil {
		return nil, err
	}
	game, err := open(settings, settings.GameDBName)
	if err != nil {
		tool.Close()
		return nil, err
	}
	return &Store{settings: settings, tool: tool, game: game, named: map[string]*sql.DB{}}, nil
}

func open(settings config.Settings, database string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local&timeout=5s",
		settings.DBUser,
		settings.DBPassword,
		settings.ServerHost,
		settings.DBPort,
		database,
		settings.DBCharset,
	)
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)
	return conn, nil
}

func (s *Store) Ping() error {
	if err := s.tool.Ping(); err != nil {
		return fmt.Errorf("ping %s: %w", s.settings.ToolDBName, err)
	}
	if err := s.game.Ping(); err != nil {
		return fmt.Errorf("ping %s: %w", s.settings.GameDBName, err)
	}
	requiredDatabases := []string{"taiwan_cain", "taiwan_cain_2nd", "taiwan_billing"}
	for _, database := range requiredDatabases {
		conn, err := s.GameNamed(database)
		if err != nil {
			return err
		}
		if err := conn.Ping(); err != nil {
			return fmt.Errorf("ping %s: %w", database, err)
		}
	}
	return nil
}

func (s *Store) Close() error {
	if s.tool != nil {
		_ = s.tool.Close()
	}
	if s.game != nil {
		_ = s.game.Close()
	}
	for _, conn := range s.named {
		_ = conn.Close()
	}
	return nil
}

func (s *Store) Tool() *sql.DB {
	return s.tool
}

func (s *Store) Game() *sql.DB {
	return s.game
}

func (s *Store) GameNamed(database string) (*sql.DB, error) {
	if database == s.settings.GameDBName {
		return s.game, nil
	}
	if database == s.settings.ToolDBName {
		return s.tool, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if conn, ok := s.named[database]; ok {
		return conn, nil
	}
	conn, err := open(s.settings, database)
	if err != nil {
		return nil, fmt.Errorf("open database %s: %w", database, err)
	}
	s.named[database] = conn
	return conn, nil
}

func (s *Store) MustGameNamed(database string) *sql.DB {
	conn, err := s.GameNamed(database)
	if err != nil {
		panic(err)
	}
	return conn
}
