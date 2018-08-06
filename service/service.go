package service

import (
	"fmt"
	path "path"

	"crawlers/logrus"
	"database/sql"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

type (
	config struct {
		// 数据库相关参数
		DbUrl     string `toml:"db_url"`      // 数据库URL
		DbPort    string `toml:"db_port"`     // 数据库端口
		DbName    string `toml:"db_name"`     // 数据库名称
		DbUser    string `toml:"db_user"`     // 数据库用户名
		DbPasswd  string `toml:"db_passwd"`   // 数据库密码
		DbMaxConn int    `toml:"db_max_conn"` //数据库最大连接数
		DbMaxIdle int    `toml:"db_max_idle"` //最大空闲链接

		// REDIS设置
		RedisUrl      string `toml:"redis_url"`
		RedisPoolSize int    `toml:"redis_pool_size"`
		RedisDb       int    `toml:"redis_db"`
		RedisPass     string `toml:"redis_pass"`
		//日志配置
		LogLevel string `toml:"log_level"`
		LogDest  string `toml:"log_dest"`
		LogDir   string `toml:"log_dir"`

		Action string
	}

	CrawlersContext struct {
		RedisClient *redis.Client
		MysqlClient *sql.DB
		Logrus      *logrus.Logger
		XormSession *xorm.Engine
		Config      *config
		Clear       chan bool
	}
)

func (this *CrawlersContext) Start() {

	this.Logrus = logrus.NewLogger(this.Config.LogLevel, this.Config.LogDest, path.Join(this.Config.LogDir, this.Config.Action))
	this.LoadRedis()
	this.LoadMysql()
}

func (this *CrawlersContext) Stop() {

	if this.RedisClient != nil {
		this.Logrus.Info("正在关闭Redis程序.....")
		this.RedisClient.Close()
	}

	if this.MysqlClient != nil {
		this.Logrus.Info("正在关闭Mysql程序.....")
		this.MysqlClient.Close()
	}

	if this.XormSession != nil {
		this.Logrus.Info("正在关闭XORM程序.....")
		this.XormSession.Close()
	}

}

func (this *CrawlersContext) LoadRedis() {
	this.Logrus.Info("正在加载Redis--------->")

	this.RedisClient = redis.NewClient(&redis.Options{
		Addr:     this.Config.RedisUrl,
		Password: this.Config.RedisPass,
		DB:       this.Config.RedisDb,
	})

	if pong, err := this.RedisClient.Ping().Result(); err != nil {
		this.Logrus.Fatal("PING redis失败: " + pong + err.Error())
		return
	}

	this.Logrus.Debug(this.RedisClient)

	this.Logrus.Info("Redis加载成功<---------")
}

func (this *CrawlersContext) LoadMysql() {
	this.Logrus.Info("正在加载MySQL--------->")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", this.Config.DbUser, this.Config.DbPasswd, this.Config.DbUrl, this.Config.DbPort, this.Config.DbName)
	this.Logrus.Info(dsn)

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		this.Logrus.Fatal("初始化mysql失败: " + err.Error())
		return
	}

	if e := db.Ping(); e != nil {
		fmt.Println("PING数据库失败: " + dsn)
		return
	}

	db.SetMaxIdleConns(this.Config.DbMaxIdle)
	db.SetMaxOpenConns(this.Config.DbMaxConn)
	this.MysqlClient = db

	this.Logrus.Info("MySQL加载成功<---------")

	// 1、初始化数据库连接

	//dsn := params.DbUser + ":" + params.DbPasswd + "@tcp" + "(" + params.DbUrl + ":" + params.DbPort + ")" + "/" + params.Dbe + "?charset=utf8"

	this.XormSession, err = xorm.NewEngine("mysql", dsn)

	if err != nil {
		this.Logrus.Fatal("初始化XORM失败: " + err.Error())
		return
	}
}
