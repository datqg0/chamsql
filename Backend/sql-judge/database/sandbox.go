package database

import (
	"database/sql"
	"fmt"
	"regexp"
)

func CreateSandboxDB(name string) (*sql.DB, error) {
	// Validate database name to prevent SQL injection
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(name) {
		return nil, fmt.Errorf("invalid database name")
	}

	root, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/")
	if err != nil {
		return nil, err
	}
	defer root.Close()

	if _, err := root.Exec("CREATE DATABASE " + name); err != nil {
		return nil, err
	}

	return sql.Open("mysql", fmt.Sprintf("root:123456@tcp(127.0.0.1:3306)/%s", name))
}

func DropSandboxDB(name string) error {
	// Validate database name to prevent SQL injection
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(name) {
		return fmt.Errorf("invalid database name")
	}

	root, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/")
	if err != nil {
		return err
	}
	defer root.Close()

	_, err = root.Exec("DROP DATABASE " + name)
	return err
}
