package main

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/lunny/xorm"
)

func GetEngine() (*xorm.Engine, error) {
	engine, err := xorm.NewEngine("sqlite3", "./ouro.db")
	if err != nil {
		return nil, err
	}
	return engine, err
}