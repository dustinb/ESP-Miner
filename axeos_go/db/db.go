package db

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Bitaxe struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	MacAddr   string `gorm:"primaryKey" json:"macAddr"`
	IP        string
	Hostname  string
}

type Measurement struct {
	gorm.Model
	IP        string
	Hostname  string
	MacAddr   string  `json:"macAddr"`
	Temp      float64 `json:"temp"`
	VRTemp    float64 `json:"vrTemp"`
	Hashrate  float64 `json:"hashRate"`
	Frequency int     `json:"frequency"`
}

type Config struct {
	gorm.Model
	IP             string
	Hostname       string
	Frequency      int
	CoreVoltage    int
	AsicCount      int
	SmallCoreCount int
	Passed         bool
	Measurements   []Measurement
}

var Database *gorm.DB

func InitDB() {
	var err error
	Database, err = gorm.Open(sqlite.Open("axego.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	Database.AutoMigrate(&Bitaxe{})
	Database.AutoMigrate(&Config{})
	Database.AutoMigrate(&Measurement{})
}
