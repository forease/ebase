package ebase

import (
	"errors"
	"fmt"
	redis "github.com/alphazero/Go-Redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

var Dbh *Models

type Models struct {
	Orm    *xorm.Engine
	Redis  redis.Client
	Driver string
}

func NewModels() (*Models, error) {
	orm, err := NewXorm()
	if err != nil {
		return nil, err
	}
	redis, err := NewRedis()
	if err != nil {
		return nil, err
	}

	Dbh = new(Models)
	Dbh.Orm = orm
	Dbh.Redis = redis

	return Dbh, nil
}

func (dbh *Models) Close() error {

	dbh.Orm.Close()
	dbh.Redis.Quit()

	return nil
}

func NewXorm() (orm *xorm.Engine, err error) {
	Log.Info("db initializing...")
	var dsn string
	dbDriver, _ := Config.String("database.driver", "postgres")
	dbProto, _ := Config.String("database.proto", "tcp")
	dbHost, _ := Config.String("database.host", "localhost")
	dbUser, _ := Config.String("database.user", "")
	dbPassword, _ := Config.String("database.password", "")
	dbName, _ := Config.String("database.name", "")
	dbSsl, _ := Config.String("database.ssl", "")
	dbPath, _ := Config.String("database.path", "")
	dbPort, _ := Config.Int("database.port", 5432)
	dbDebug, _ := Config.Bool("database.debug", false)

	switch dbDriver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=utf8", dbUser, dbPassword,
			dbProto, dbHost, dbPort, dbName)
	case "postgres":
		dsn = fmt.Sprintf("dbname=%s host=%s user=%s password=%s port=%d sslmode=%s",
			dbName, dbHost, dbUser, dbPassword,
			dbPort, dbSsl)
	case "sqlite3":
		dsn = dbPath + dbName
	default:
		return nil, errors.New("Unspport Database Driver")
	}

	orm, err = xorm.NewEngine(dbDriver, dsn)
	if err != nil {
		Log.Panic("NewEngine", err)
	}

	orm.TZLocation = time.Local
	orm.ShowSQL = dbDebug
	//orm.Logger = xorm.NewSimpleLogger(Log.Loger)
	return orm, nil
}

func NewRedis() (redis.Client, error) {
	redisHost, _ := Config.String("redis.host", "localhost")
	redisAuth, _ := Config.String("redis.auth", "")
	//redisKeyFix, _ := Config.String("redis.key_prefix", "")
	redisPort, _ := Config.Int("redis.port", 6379)
	redisDb, _ := Config.Int("redis.db", 0)

	spec := redis.DefaultSpec().Db(redisDb)

	if redisHost != "" {
		spec.Host(redisHost)
	}
	if redisAuth != "" {
		spec.Password(redisAuth)
	}
	if redisPort > 0 && redisPort < 65535 {
		spec.Port(redisPort)
	}

	rd, err := redis.NewSynchClientWithSpec(spec)
	if err != nil {
		return nil, err
	}

	return rd, nil
}
