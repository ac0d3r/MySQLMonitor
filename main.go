package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type logOutPutType string

type logOutPut struct {
	VariableName string `sql:"Variable_name"`
	Value        string `sql:"Value"`
}
type execLog struct {
	EventTime string `sql:"event_time"`
	UserHost  string `sql:"user_host"`
	Argument  string `sql:"argument"`
}

const (
	fileLogOutPut  logOutPutType = "FILE"
	tableLogOutPut logOutPutType = "TABLE"
	// sql
	setLogSQL            = "SET GLOBAL general_log='ON'"
	showLogOutPutTypeSQL = "SHOW VARIABLES LIKE 'log_output'"
	setLogOutPutTypeSQL  = "SET GLOBAL log_output=?"
	getExecLogSQL        = `SELECT event_time, user_host, argument FROM mysql.general_log WHERE command_type ="Query" OR command_type ="Execute" ORDER BY event_time DESC LIMIT 2`
)

var (
	db       *sql.DB
	flagHelp = flag.Bool("h", false, "Shows usage options.")
	flagHost = flag.String("host", "localhost", "Bind mysql host.")
	flagPort = flag.Uint("port", 3306, "Bind mysql port.")

	flagUser   = flag.String("user", "", "Select mysql username.")
	flagPasswd = flag.String("passwd", "", "Input mysql password.")
)

func banner() {
	fmt.Println(`Starting monitor MySQL Query log...`)
}

func initDB() {
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", *flagUser, *flagPasswd, *flagHost, *flagPort))
	if err != nil {
		log.Fatalf("connect to mysql failed: %q", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
}

func checkLogOutPut() {
	var logout logOutPut
	row := db.QueryRow(showLogOutPutTypeSQL)
	err := row.Scan(&logout.VariableName, &logout.Value)
	if err != nil {
		log.Fatalf("exec %s failed: %q", showLogOutPutTypeSQL, err)
	}
	if logout.Value != string(tableLogOutPut) {
		_, err := db.Exec(setLogOutPutTypeSQL, tableLogOutPut)
		if err != nil {
			log.Fatalf("exec %s failed: %q", setLogOutPutTypeSQL, err)
		}
	}
}

func printExecLog() bool {
	var hasnew bool
	rows, err := db.Query(getExecLogSQL)
	if err != nil {
		log.Fatalf("exec %s failed: %q", getExecLogSQL, err)
	}
	for rows.Next() {
		var elog execLog
		err := rows.Scan(&elog.EventTime, &elog.UserHost, &elog.Argument)
		if err != nil {
			log.Fatalf("printExecLog rows.Scan failed: %q", err)
		}
		if elog.Argument != getExecLogSQL {
			hasnew = true
			log.Printf("[exec] %s\n", elog.Argument)
		}
	}
	return hasnew
}

func fetchRows(rows *sql.Rows) {
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// Fetch rows
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		var value string
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			// handlerFunc(columns[i])
			fmt.Println(columns[i], value)
		}
	}
	if err = rows.Err(); err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
}

func main() {
	flag.Parse()
	if *flagHelp || *flagUser == "" {
		fmt.Printf("Usage: MySQLMonitor [options]\n\n")
		flag.PrintDefaults()
		return
	}
	banner()
	initDB()
	defer db.Close()

	_, err := db.Exec(setLogSQL)
	if err != nil {
		log.Fatalf("exec %s failed: %q", setLogSQL, err)
	}
	checkLogOutPut()
	for {
		hasnew := printExecLog()
		if !hasnew {
			time.Sleep(time.Millisecond * 100)
		}
	}
}
