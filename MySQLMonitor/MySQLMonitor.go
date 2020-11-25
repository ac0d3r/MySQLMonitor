package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	getExecLogSQL        = `SELECT event_time, user_host, argument FROM mysql.general_log WHERE command_type ="Query" OR command_type ="Execute" ORDER BY event_time DESC LIMIT ?`
)

var (
	sigs = make(chan os.Signal, 1)

	db         *sql.DB
	limitNum   int = 1
	lastone    time.Time
	getExecLog = getExecLogSQL[:len(getExecLogSQL)-1]

	flagHelp   = flag.Bool("h", false, "Shows usage options.")
	flagHost   = flag.String("host", "localhost", "Bind mysql host.")
	flagPort   = flag.Uint("port", 3306, "Bind mysql port.")
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

func formatDatetime(timestr string) (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05.000000", timestr)
	if err != nil {
		return t, err
	}
	return t, nil
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
	rows, err := db.Query(getExecLogSQL, limitNum)
	if err != nil {
		log.Fatalf("exec %s failed: %q", getExecLogSQL, err)
	}
	for rows.Next() {
		var elog execLog
		err := rows.Scan(&elog.EventTime, &elog.UserHost, &elog.Argument)
		if err != nil {
			log.Fatalf("printExecLog rows.Scan failed: %q", err)
		}
		eventTime, err := formatDatetime(elog.EventTime)
		if err != nil {
			log.Fatalf("printExecLog time.format %s error: %q", elog.EventTime, err)
			continue
		}
		// 初始化 lastone
		if limitNum == 1 {
			lastone = eventTime
			limitNum = 5
		} else if eventTime.After(lastone) {
			if !strings.Contains(elog.Argument, getExecLog) {
				hasnew = true
				lastone = eventTime
				fmt.Printf("[%s] - exec: %s\n", eventTime.Format("01-02 15:04:05"), elog.Argument)
			}
		}
	}
	return hasnew
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

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		fmt.Println("Bye!")
		db.Close()
	}()

	_, err := db.Exec(setLogSQL)
	if err != nil {
		log.Fatalf("exec %s failed: %q", setLogSQL, err)
	}
	checkLogOutPut()

	for {
		select {
		case <-sigs:
			goto BREAK
		default:
			hasnew := printExecLog()
			if !hasnew {
				time.Sleep(time.Millisecond * 150)
			}
		}
		continue
	BREAK:
		break
	}
}
