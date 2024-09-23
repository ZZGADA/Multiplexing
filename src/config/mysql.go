package config

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"time"
)

type Config struct {
	Database MySqlConfig `yaml:"db" mapstructure:"db"`
}

type MySqlConfig struct {
	host     string `yaml:"host" mapstructure:"host"`
	username string `yaml:"username" mapstructure:"username"`
	password string `yaml:"password" mapstructure:"password"`
	port     string `yaml:"port" mapstructure:"port"`
	dataBase string `yaml:"database" mapstructure:"database"`
}

const configName = "application"
const suffix = "yaml"
const path = "./conf"

var mysqlClient *gorm.DB

func init() {
	const mysqlConnectStr string = "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"

	// 配置读取yaml 文件
	viper.SetConfigName(configName) // 配置文件名称(无扩展名)
	viper.SetConfigType(suffix)     // 或viper.SetConfigType("YAML")
	viper.AddConfigPath(path)       // 配置文件路径
	if err := viper.ReadInConfig(); err != nil {
		// 处理读取配置文件的错误
		// 小写
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	fmt.Println(viper.AllKeys())
	fmt.Println(viper.GetString("db.host"))

	MySqlConfigIns := &MySqlConfig{host: viper.GetString("db.host"),
		port:     viper.GetString("db.port"),
		dataBase: viper.GetString("db.database"),
		username: viper.GetString("db.username"),
		password: viper.GetString("db.password")}

	log.Printf("配置读取成功，%#v", MySqlConfigIns)

	dsn := fmt.Sprintf(mysqlConnectStr,
		MySqlConfigIns.username,
		MySqlConfigIns.password,
		MySqlConfigIns.host,
		MySqlConfigIns.port,
		MySqlConfigIns.dataBase)

	client, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",   // 表前缀
			SingularTable: true, // 禁用表名复数
		}})
	if err != nil {
		panic(err)
	}

	sqlDB, _ := client.DB()
	// SetMaxIdleConnections 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConnections 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(10)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(10 * time.Second)

	mysqlClient = client
}

func GetDatabaseClient() *gorm.DB {
	return mysqlClient
}
