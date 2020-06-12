package ebase

import (
	"encoding/json"
	"errors"
	"fmt"
	redis "github.com/alphazero/Go-Redis"
	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

var Dbh *Models

type (
	Redis struct {
		redis.Client
		RedisPrefix string
	}

	Models struct {
		Orm          *xorm.Engine
		OrmCache     bool
		OrmCacheTime int64
		Redis        *Redis
		RedisEnable  bool
		RedisPrefix  string
		Driver       string
	}

	ModelOption struct {
		Orm   OrmOption
		Redis RedisOption
	}

	OrmOption struct {
		Driver    string
		Proto     string
		Host      string
		User      string
		Password  string
		Name      string
		Ssl       string
		Path      string
		Log       string
		Port      int
		CacheTime int
		Cache     bool
		Debug     bool
	}

	RedisOption struct {
		Host   string
		Auth   string
		Prefix string
		Port   int
		Db     int
		Enable bool
	}

	PageOptions struct {
		OrderIdx  int
		SearchKey string
		OrderBy   string
	}
)

// 从默认配置加载数据库

func NewDefaultModels() (dbh *Models, err error) {
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
	dbCache, _ := Config.Bool("database.cache", false)
	dbCacheTime, _ := Config.Int("database.cachetime", 300)
	dbLogFile, _ := Config.String("database.log", "var/database.log")

	redisEnable, _ := Config.Bool("redis.enable", false)
	redisHost, _ := Config.String("redis.host", "localhost")
	redisAuth, _ := Config.String("redis.auth", "")
	redisPort, _ := Config.Int("redis.port", 6379)
	redisDb, _ := Config.Int("redis.db", 0)
	redisKeyFix, _ := Config.String("redis.keyprefix", "ebase")

	opt := new(ModelOption)
	opt.Orm.Driver = dbDriver
	opt.Orm.Host = dbHost
	opt.Orm.Proto = dbProto
	opt.Orm.User = dbUser
	opt.Orm.Password = dbPassword
	opt.Orm.Name = dbName
	opt.Orm.Ssl = dbSsl
	opt.Orm.Path = dbPath
	opt.Orm.Port = dbPort
	opt.Orm.Debug = dbDebug
	opt.Orm.Cache = dbCache
	opt.Orm.CacheTime = dbCacheTime
	opt.Orm.Log = dbLogFile

	opt.Redis.Host = redisHost
	opt.Redis.Enable = redisEnable
	opt.Redis.Auth = redisAuth
	opt.Redis.Port = redisPort
	opt.Redis.Db = redisDb
	opt.Redis.Prefix = redisKeyFix

	dbh, err = NewModels(opt)
	if err == nil {
		Dbh = dbh
	}

	return
}

// new model
// exmple
//	opt := new(ModelOption)
//	opt.Orm.Driver = dbDriver
//	opt.Orm.Host = dbHost
//	opt.Orm.Proto = dbProto
//	opt.Orm.User = dbUser
//	opt.Orm.Password = dbPassword
//	opt.Orm.Name = dbName
//	opt.Orm.Ssl = dbSsl
//	opt.Orm.Path = dbPath
//	opt.Orm.Port = dbPort
//	opt.Orm.Debug = dbDebug
//	opt.Orm.Cache = dbCache
//	opt.Orm.CacheTime = dbCacheTime
//
//	opt.Redis.Host = redisHost
//	opt.Redis.Enable = redisEnable
//	opt.Redis.Auth = redisAuth
//	opt.Redis.Port = redisPort
//	opt.Redis.Db = redisDb
//	opt.Redis.Prefix = redisKeyFix
//
//	dbh, err = NewModels(opt)

func NewModels(opt *ModelOption) (*Models, error) {
	orm, err := NewXorm(&opt.Orm)
	if err != nil {
		return nil, err
	}

	dbh := new(Models)
	dbh.Orm = orm

	if opt.Redis.Enable {
		redis, err := NewRedis(&opt.Redis)
		if err != nil {
			orm.Close()
			return nil, err
		}
		dbh.OrmCache = opt.Orm.Cache
		dbh.OrmCacheTime = int64(opt.Orm.CacheTime)
		dbh.RedisEnable = true

		dbh.Redis = redis
	}

	return dbh, nil
}

func (dbh *Models) Close() error {

	dbh.Orm.Close()
	if dbh.RedisEnable {
		dbh.Redis.Quit()
	}

	return nil
}

func NewXorm(opt *OrmOption) (orm *xorm.Engine, err error) {
	Log.Trace("db initializing...")
	var dsn string

	switch opt.Driver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=utf8", opt.User, opt.Password,
			opt.Proto, opt.Host, opt.Port, opt.Name)
	case "postgres":
		dsn = fmt.Sprintf("dbname=%s host=%s user=%s password=%s port=%d sslmode=%s",
			opt.Name, opt.Host, opt.User, opt.Password, opt.Port, opt.Ssl)
	case "sqlite3":
		dsn = opt.Path + opt.Name
	default:
		return nil, errors.New("Unspport Database Driver")
	}

	orm, err = xorm.NewEngine(opt.Driver, dsn)
	if err != nil {
		Log.Panic("NewEngine", err)
	}

	orm.TZLocation = time.Local
	//orm.ShowSQL = opt.Debug
	//orm.Logger = xorm.NewSimpleLogger(Log.Loger)
	if opt.Debug {
		f, err := os.Create(opt.Log)
		if err != nil {
			println(err.Error())
			return orm, err
		}
		orm.Logger().SetLevel(5)
		logger := xorm.NewSimpleLogger(f)
		logger.ShowSQL(true)
		orm.SetLogger(logger)
	}
	return orm, nil
}

func NewRedis(opt *RedisOption) (*Redis, error) {

	spec := redis.DefaultSpec().Db(opt.Db)

	if opt.Host != "" {
		spec.Host(opt.Host)
	}
	if opt.Auth != "" {
		spec.Password(opt.Auth)
	}
	if opt.Port > 0 && opt.Port < 65535 {
		spec.Port(opt.Port)
	}

	rd, err := redis.NewSynchClientWithSpec(spec)
	if err != nil {
		return nil, err
	}
	ret := &Redis{Client: rd, RedisPrefix: opt.Prefix}
	return ret, nil
}

func (self *Redis) RedisSetJson(expire int64, mp interface{}, keys ...interface{}) (err error) {
	encoder, err := json.Marshal(mp)
	if err != nil {
		Log.Errorf("Json Marshal err %s", err)
		return
	}

	key := self.GetRedisKey(keys...)

	err = self.Set(key, encoder)
	if err != nil {
		Log.Errorf("Redis set err %s", err)
		return
	}
	if expire > 0 {
		self.Expire(key, expire)
	}

	return nil
}

func (self *Redis) RedisGetJson(v interface{}, keys ...interface{}) (err error) {

	value, err := self.RedisGet(keys...)
	if err != nil {
		return
	} else if value == nil {
		return errors.New("key not exists")
	}

	err = json.Unmarshal(value, v)
	if err != nil {
		Log.Errorf("Json Unmarshal err %s", err)
		return
	}

	return
}

func (self *Redis) RedisSet(expire int64, value interface{}, keys ...interface{}) (err error) {
	var data []byte

	switch value.(type) {
	case int:
		data = []byte(GetIntStr(value.(int)))
	case int64:
		data = []byte(GetInt64Str(value.(int64)))
	case uint64:
		data = []byte(GetUint64Str(value.(uint64)))
	case string:
		data = []byte(value.(string))
	case []byte:
		data = value.([]byte)
	default:
		data, err = json.Marshal(value)
		if err != nil {
			Log.Errorf("Json Marshal err %s", err)
			return
		}
	}

	key := self.GetRedisKey(keys...)
	err = self.Set(key, data)
	if err != nil {
		Log.Errorf("Redis set err %s", err)
		return
	}

	if expire > 0 {
		self.Expire(key, expire)
	}

	return nil
}

func (self *Redis) RedisGet(keys ...interface{}) (value []byte, err error) {
	key := self.GetRedisKey(keys...)
	value, err = self.Get(key)
	if err == nil {
		return value, nil
	}

	return self.Get(key)
}

func (self *Redis) RedisExists(keys ...interface{}) (bool, error) {
	key := self.GetRedisKey(keys...)

	return self.Exists(key)
}

func (self *Redis) RedisIncr(keys ...interface{}) (int64, error) {

	key := self.GetRedisKey(keys...)

	return self.Incr(key)
}

func (self *Redis) RedisDelete(keys ...interface{}) (bool, error) {

	key := self.GetRedisKey(keys...)

	return self.Del(key)
}

func (self *Redis) RedisDeleteAll(keys ...interface{}) (s bool, err error) {

	key := self.GetRedisKey(keys...)
	var list []string
	list, err = self.Keys(key + "*")
	if err != nil {
		return
	}
	for _, k := range list {
		self.Del(k)
	}

	return true, nil
}

func (self *Redis) GetRedisKey(keys ...interface{}) (key string) {
	key = self.RedisPrefix
	for _, k := range keys {

		switch k.(type) {
		case int:
			key += ":" + GetIntStr(k.(int))
		case int64:
			key += ":" + GetInt64Str(k.(int64))
		case uint64:
			key += ":" + GetUint64Str(k.(uint64))
		case string:
			key += ":" + k.(string)
		default:
			key += ":" + fmt.Sprintf("%v", k)
		}
	}
	return
}
