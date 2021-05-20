package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	flagHelp   = flag.Bool("h", false, "Shows usage options.")
	flagHost   = flag.String("host", "localhost", "Bind mysql host.")
	flagPort   = flag.Uint("port", 3306, "Bind mysql port.")
	flagUser   = flag.String("user", "", "Select mysql username.")
	flagPasswd = flag.String("passwd", "", "Input mysql password.")

	flagStart = flag.Bool("start", false, "start mysql monitor.")
	flagClean = flag.Bool("clean", false, "clean mysql general_log.")
)

var (
	db *sql.DB
)

func banner() {
	t := `
	███╗   ███╗██╗   ██╗███████╗ ██████╗ ██╗     
	████╗ ████║╚██╗ ██╔╝██╔════╝██╔═══██╗██║     
	██╔████╔██║ ╚████╔╝ ███████╗██║   ██║██║     
	██║╚██╔╝██║  ╚██╔╝  ╚════██║██║▄▄ ██║██║     
	██║ ╚═╝ ██║   ██║   ███████║╚██████╔╝███████╗
	╚═╝     ╚═╝   ╚═╝   ╚══════╝ ╚══▀▀═╝ ╚══════╝
███╗   ███╗ ██████╗ ███╗   ██╗██╗████████╗ ██████╗ ██████╗ 
████╗ ████║██╔═══██╗████╗  ██║██║╚══██╔══╝██╔═══██╗██╔══██╗
██╔████╔██║██║   ██║██╔██╗ ██║██║   ██║   ██║   ██║██████╔╝
██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║   ██║   ██║██╔══██╗
██║ ╚═╝ ██║╚██████╔╝██║ ╚████║██║   ██║   ╚██████╔╝██║  ██║
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝`
	fmt.Println(t)
}

func main() {
	flag.Parse()
	if *flagHelp || *flagUser == "" {
		fmt.Println("Usage: MySQLMonitor [options]")
		flag.PrintDefaults()
		return
	}
	banner()
	if err := initDB(); err != nil {
		log.Fatalf("initDB error: %s", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database connection error: %s \n", err)
		}
		fmt.Println("\nBye hacker :)")
	}()

	if *flagClean {
		fmt.Println("clean mysql `general_log`...")
		if err := cleanLog(); err != nil {
			log.Fatalf("cleanLog error: %s", err)
		}
		return
	}

	if *flagStart {
		fmt.Println("start mysql monitor ...")
		if err := setMySQLLogOutput(); err != nil {
			log.Fatalf("setMySQLLogOutput error: %s", err)
		}
		watchdog()
	}
}

func initDB() error {
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true", *flagUser, *flagPasswd, *flagHost, *flagPort))
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	return nil
}

type mysqlVariable struct {
	Name  string `sql:"Variable_name"`
	Value string `sql:"Value"`
}

func setMySQLLogOutput() error {
	if _, err := db.Exec("SET GLOBAL general_log='ON'"); err != nil {
		return err
	}

	variable := mysqlVariable{}
	row := db.QueryRow("SHOW VARIABLES LIKE 'log_output'")
	if err := row.Scan(&variable.Name, &variable.Value); err != nil {
		return err
	}
	if variable.Name != "log_output" {
		return fmt.Errorf("exec: `SHOW VARIABLES LIKE 'log_output'` not found log_output")
	}
	if variable.Value != "TABLE" {
		if _, err := db.Exec("SET GLOBAL log_output = ? ", "TABLE"); err != nil {
			return err
		}
	}
	return nil
}

func cleanLog() error {
	if _, err := db.Exec("RENAME TABLE mysql.`general_log` TO mysql.`general_log_temp`"); err != nil {
		return err
	}
	if _, err := db.Exec("DELETE FROM mysql.`general_log_temp`"); err != nil {
		return err
	}
	if _, err := db.Exec("RENAME TABLE mysql.`general_log_temp` TO mysql.`general_log`"); err != nil {
		return err
	}
	return nil
}

type execLog struct {
	EventTime time.Time `sql:"event_time"`
	UserHost  string    `sql:"user_host"`
	Argument  string    `sql:"argument"`
}

func watchdog() {
	var (
		hasnew  bool
		limit   int = 1
		lastone time.Time

		execLogSQL = `SELECT event_time, user_host, argument FROM mysql.general_log WHERE command_type ="Query" OR command_type ="Execute" ORDER BY event_time DESC LIMIT ?`
		sqlPrefix  = execLogSQL[:len(execLogSQL)-1]
	)

	defer func() {
		if _, err := db.Exec("SET GLOBAL general_log='OFF'"); err != nil {
			log.Printf("set `general_log`='OFF' error: %s \n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

LOOP:
	for {
		select {
		case <-quit:
			break LOOP
		default:
			// print execlog with table
			hasnew = false
			rows, err := db.Query(execLogSQL, limit)
			if err != nil {
				log.Printf("watchdog Query error: %s", err)
				goto LOOP
			}
			for rows.Next() {
				var elog execLog
				err = rows.Scan(&elog.EventTime, &elog.UserHost, &elog.Argument)
				if err != nil {
					log.Printf("printExecLog rows.Scan error: %s", err)
					continue
				}
				// 初始化 lastone
				if limit == 1 {
					limit = 5
					lastone = elog.EventTime
				} else if elog.EventTime.After(lastone) && !strings.Contains(elog.Argument, sqlPrefix) {
					hasnew = true
					lastone = elog.EventTime
					fmt.Printf("[%s] - exec: %s\n", elog.EventTime.Format("2006-01-02 15:04:05"), elog.Argument)
				}
			}
			if !hasnew {
				time.Sleep(time.Millisecond * 150)
			}
		}
	}
}
