package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var DBHOST = "127.0.0.1"
var DBUSER = "root"
var DBPASSWORD = "root"
var DBNAME = "online-video-room"
var DBPORT = "3306"

var DB *gorm.DB

func GetInstance() *gorm.DB {
	if DB == nil {
		connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", DBUSER, DBPASSWORD, DBHOST, DBPORT, DBNAME)
		DB, _ = gorm.Open(mysql.Open(connectionString), &gorm.Config{})
		sqlDB, err := DB.DB()
		if err != nil {
			log.Panic(err.Error())
			return nil
		}
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxIdleTime(time.Minute * 20)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}
	return DB
}
