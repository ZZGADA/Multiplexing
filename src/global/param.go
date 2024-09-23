package global

import (
	"Multiplexing_/src/config"
	"gorm.io/gorm"
)

var MySqlClient *gorm.DB

func init() {
	MySqlClient = config.GetDatabaseClient()
}
