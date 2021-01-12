# sec-tools
## MySQLMonitor
> 监控 MySQL 并实时显示程序执行过的语句

**Usage:**
```bash
$ ./bin/MySQLMonitor 
Usage: MySQLMonitor [options]

  -h    Shows usage options.
  -host string
        Bind mysql host. (default "localhost")
  -port uint
        Bind mysql port. (default 3306)
  -user string
        Select mysql username.
  -passwd string
        Input mysql password.
```

## http.server
> 创建简单的 HTTP 服务用于访问指定目录下的文件(类似于：`python3 -m http.server`)

**Usage:**
```bash
$ ./bin/http.server

Usage: http.server [options]

  -h    Shows usage options.
  -host string
        listen host (default "0.0.0.0")
  -port uint
        listen port (default 8080)
  -dir string
        listen directory (default "./")
```