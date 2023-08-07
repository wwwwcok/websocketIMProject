package tool

import (
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"websocketIMProject/model"
	"xorm.io/xorm"
)

var XormMysqlEngine *xorm.Engine
var Rediscli *redis.Client

//初始化xorm引擎
func InitXormMysqlEngine() error {

	dbEngine, err := xorm.NewEngine("mysql", "root:root@tcp(127.0.0.1:3306)/websocketIMProject?charset=utf8")
	if err != nil {
		return err
	}
	dbEngine.ShowSQL(true)
	err = dbEngine.Sync2(&model.User{}, &model.Contact{}, &model.Community{})
	if err != nil {
		return err
	}
	XormMysqlEngine = dbEngine
	//插入初始化表数据
	return err

}

func InitRedisEngine() error {
	client := redis.NewClient(&redis.Options{
		Addr:     "192.168.0.90:6379",
		Password: "",
		DB:       0,
	})

	err := client.Ping().Err()

	Rediscli = client

	return err
}
