package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	flagHelp   = flag.Bool("help", false, "Shows usage options.")
	flagHost   = flag.String("host", "localhost", "Bind mysql host.")
	flagPort   = flag.Uint("port", 3306, "Bind mysql port.")
	flagUser   = flag.String("u", "", "Select mysql username.")
	flagPasswd = flag.String("p", "", "Input mysql password.")
)

var (
	logfile string
	db      *sql.DB
	cstZone = time.FixedZone("CST", 8*3600)
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
	banner()

	if runtime.GOOS != "windows" && !isRoot() {
		log.Fatalln("run as a user with root! Thx:)")
	}

	flag.Parse()
	if *flagHelp || *flagUser == "" {
		fmt.Println("Usage: MySQLMonitor [options]")
		flag.PrintDefaults()
		return
	}

	if err := initDB(); err != nil {
		log.Fatalf("initDB error: %s", err)
	}

	defer func() {
		if err := cleanGenerakLog(); err != nil {
			log.Printf("cleanGenerakLog error: %s \n", err)
		}
		if err := db.Close(); err != nil {
			log.Printf("close database connection error: %s \n", err)
		}
		fmt.Println("\nBye hacker :)")
	}()

	fmt.Println("start mysql monitor ...")
	if err := setMySQLLogOutput(); err != nil {
		log.Fatalf("setMySQLLogOutput error: %s", err)
	}

	watchdog()
}

func initDB() error {
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local", *flagUser, *flagPasswd, *flagHost, *flagPort))
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
	variable := mysqlVariable{}
	row := db.QueryRow("SHOW VARIABLES LIKE 'general_log_file'")
	if err := row.Scan(&variable.Name, &variable.Value); err != nil {
		return err
	}
	if variable.Name == "general_log_file" {
		logfile = variable.Value
	}

	if _, err := db.Exec("SET GLOBAL log_output = 'FILE'"); err != nil {
		return err
	}
	if _, err := db.Exec("SET GLOBAL general_log='ON'"); err != nil {
		return err
	}
	return nil
}

func cleanGenerakLog() error {
	if _, err := db.Exec("SET GLOBAL general_log='OFF'"); err != nil {
		return err
	}
	if logfile != "" {
		return os.Truncate(logfile, 0)
	}
	return nil
}

func watchdog() {
	var f *os.File

	if logfile == "" {
		log.Fatalln("general_log_file was empty :(")
	}
	f, err := os.Open(logfile)
	if err != nil {
		log.Fatalf("Open '%s' error: %s", logfile, err)
	}
	defer f.Close()
	// 指向文件尾部
	_, err = f.Seek(0, 2)
	if err != nil {
		log.Fatalf("'%s' File.Seek(0,2) error: %s", logfile, err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
LOOP:
	for {
		select {
		case <-quit:
			break LOOP
		default:
			handle := func(line string) {
				if strings.Contains(line, "Execute") || strings.Contains(line, "Query") {
					lines := strings.Split(line, "\t")
					t, err := str2Time(lines[0], "2006-01-02T15:04:05Z")
					if err == nil {
						fmt.Printf("[%s] %s\n", t.In(cstZone).Format("15:04:05"), lines[2])
					} else {
						fmt.Printf("[%s] %s\n", lines[0], lines[2])
					}
				}
			}
			if err := linePrinter(f, handle); err != nil {
				log.Printf("linePrinter error: %s \n", err)
				break LOOP
			}
			time.Sleep(time.Millisecond * 550)
		}
	}
}

func linePrinter(r io.Reader, call func(string)) error {
	buf := make([]byte, 32*1024)
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		if c == 0 {
			return nil
		}
		lines := bytes.Split(buf[:c], lineSep)
		for i := range lines {
			call(string(lines[i]))
		}
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		}
	}
}

// util part

func str2Time(timestr string, format string) (time.Time, error) {
	var (
		t   time.Time
		err error
	)
	t, err = time.Parse(format, timestr)
	if err != nil {
		return t, err
	}
	return t, nil
}

func isRoot() bool {
	return os.Geteuid() == 0
}
