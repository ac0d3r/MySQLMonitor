#  MySQLMonitor
> 监控 MySQL 并实时打印执行语句

- 监控 `general_log_file` 日志文件（效率高，结果准确，不会遗漏执行语句） 🚀
- 结束执行时会关闭 `general_log`且会清空日志文件 🙈


正确以及错误执行的SQL语句：
![image](https://user-images.githubusercontent.com/26270009/124460050-74076380-ddc1-11eb-8a09-136a3f8e4279.png)

**tips**: mysql8.0或其他高版本 `general_log` 不会记录执行错误的SQL语句到日志,需要在配置文件 `[mysqld]` 下添加配置 `log-raw=1`