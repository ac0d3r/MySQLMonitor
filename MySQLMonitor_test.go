package main

import (
	"fmt"
	"log"
	"testing"
)

func TestMain(m *testing.M) {
	*flagUser = "root"
	*flagHost = "localhost"
	*flagPort = 3306
	*flagPasswd = "12345678"
	if err := initDB(); err != nil {
		log.Fatalf("initDB error: %s", err)
	}
	m.Run()
}

func Test_catMySQLVersion(t *testing.T) {
	var version string
	row := db.QueryRow("SELECT version();")
	if err := row.Scan(&version); err != nil {
		t.Fatal(err)
	}
	fmt.Println(version)
}
