package xormlib

import (
	"fmt"
	"strconv"
	"github.com/go-xorm/xorm"
	_ "github.com/go-sql-driver/mysql"
	"gylib/common"
	"time"
)

var Db_Engine *xorm.Engine
var Db_perfix string;
var Db_name string;
func init() {
	var err error
	data := common.Getini("conf/app.ini", "database", map[string]string{"db_user": "root", "db_password": "",
		"db_host": "127.0.0.1", "db_port": "3306", "db_name": "", "db_perfix":"","db_maxpool": "200", "db_minpool": "100"})
	Db_perfix=data["db_perfix"];
	Db_name=data["db_name"]
	con := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", data["db_user"],
		data["db_password"], data["db_host"],
		data["db_port"], data["db_name"])
	Db_Engine, err = xorm.NewEngine("mysql", con)
	if (err != nil) {
		fmt.Println(err)
	}

	maxpool, _ := strconv.Atoi(data["db_maxpool"])
	minpool, _ := strconv.Atoi(data["db_minpool"])
	Db_Engine.SetMaxOpenConns(maxpool)
	Db_Engine.SetMaxIdleConns(minpool)
	Db_Engine.TZLocation, _ = time.LoadLocation("Asia/Shanghai")
	//Db_Engine.Sync2(new(model.YzkLogin))
}
