package storage

import (
	"fmt"
	"gopkg.in/ini.v1"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
)

// Db is the SQLite3 db handler ype
type Db struct {
	Path          string
	Disabled      bool
	SkipMigration bool
	UseMySQL      bool
}

// NewDb sets up a new DB
func NewDb() *Db {
	return &Db{}
}

// Get gets a db handle
func (db *Db) Get() (*gorm.DB, error) {

	if db.Disabled {
		return nil, nil
	}

	if db.UseMySQL {

		cfg, err := ini.Load(db.Path)
		if err != nil {
			fmt.Printf("Fail to read file: %v", err)
			os.Exit(1)
		}

		conn, err := gorm.Open(mysql.New(mysql.Config{
			DSN:                       fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.Section("mysql").Key("user").String(), cfg.Section("mysql").Key("pass").String(), cfg.Section("mysql").Key("host").String(), cfg.Section("mysql").Key("port").String(), cfg.Section("mysql").Key("db").String()), // data source name
			DefaultStringSize:         256,                                                                                                                                                                                                                                                                                              // default size for string fields
			DisableDatetimePrecision:  true,                                                                                                                                                                                                                                                                                             // disable datetime precision, which not supported before MySQL 5.6
			DontSupportRenameIndex:    true,                                                                                                                                                                                                                                                                                             // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
			DontSupportRenameColumn:   true,                                                                                                                                                                                                                                                                                             // `change` when rename column, rename column not supported before MySQL 8, MariaDB
			SkipInitializeWithVersion: false,                                                                                                                                                                                                                                                                                            // auto configure based on currently MySQL version
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})
		if err != nil {
			return nil, err
		}
		if !db.SkipMigration {
			conn.AutoMigrate(&URL{}, &Header{}, &TLS{}, &TLSCertificate{}, &TLSCertificateDNSName{})
		}

		return conn, nil
	} else {
		conn, err := gorm.Open(sqlite.Open(db.Path+"?cache=shared"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})
		if err != nil {
			return nil, err
		}
		if !db.SkipMigration {
			conn.AutoMigrate(&URL{}, &Header{}, &TLS{}, &TLSCertificate{}, &TLSCertificateDNSName{})
		}

		return conn, nil
	}

}

// OrderPerception orders by perception hash if enabled
func OrderPerception(enabled bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if enabled {
			return db.Order("perception_hash desc")
		}
		return db
	}
}
