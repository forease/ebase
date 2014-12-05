package ebase

import (
	"encoding/json"
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

type Redis struct {
	redis.Client
	RedisPrefix string
}

type Models struct {
	Orm          *xorm.Engine
	OrmCache     bool
	OrmCacheTime int64
	Redis        *Redis
	RedisEnable  bool
	RedisPrefix  string
	Driver       string
}

func NewModels() (*Models, error) {
	orm, err := NewXorm()
	if err != nil {
		return nil, err
	}

	Dbh = new(Models)
	Dbh.Orm = orm

	redisEnable, _ := Config.Bool("redis.enable", false)
	if redisEnable {
		redis, err := NewRedis()
		if err != nil {
			return nil, err
		}
		dbCache, _ := Config.Bool("database.cache", false)
		dbCacheTime, _ := Config.Int("database.cachetime", 300)
		Dbh.OrmCache = dbCache
		Dbh.OrmCacheTime = int64(dbCacheTime)
		Dbh.RedisEnable = true

		Dbh.Redis = redis

	}

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

func NewRedis() (*Redis, error) {
	//redisEnable, _ := Config.Bool("redis.enable", false)
	redisHost, _ := Config.String("redis.host", "localhost")
	redisAuth, _ := Config.String("redis.auth", "")
	redisPort, _ := Config.Int("redis.port", 6379)
	redisDb, _ := Config.Int("redis.db", 0)
	redisKeyFix, _ := Config.String("redis.keyprefix", "ebase")

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
	ret := &Redis{Client: rd, RedisPrefix: redisKeyFix}
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
