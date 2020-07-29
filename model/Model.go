package model

import (
	_ "github.com/go-sql-driver/mysql"
)

type Model struct {
	AddDatetime    string `gorm:"column:add_datetime;-" json:"add_datetime"`
	UpdateDatetime string `gorm:"column:upt_datetime;-" json:"update_datetime"`
}

var (
	prefix string
)

func init() {
	prefix = "fs_"
}
