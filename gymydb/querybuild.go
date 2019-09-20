package gymydb

import (
	"fmt"
	"database/sql"
	"strconv"
	"gylib/common"
	"strings"
	"crypto/md5"
	"encoding/hex"
	"time"
)

var mysqldb *sql.DB
var Slavedb []*sql.DB

type Db_conn struct {
	Db_host     string
	Db_port     string
	Db_name     string
	Db_password string
	Db_perfix   string
}

var Db_perfix string
var Db_Struct Db_conn
var Is_db_init bool = false
var G_dbtables map[string]interface{}
var G_fd_list map[string]interface{}
var G_tb_dict map[string]interface{}
var G_fd_dict map[string]interface{}

func init() {
	G_dbtables = make(map[string]interface{})
	G_fd_list = make(map[string]interface{})
	G_tb_dict = make(map[string]interface{})
	G_fd_dict = make(map[string]interface{})
	//DataTable=make(map[string]string)
	data := common.Getini("conf/app.ini", "database", map[string]string{"db_user": "root", "db_password": "",
		"db_host": "127.0.0.1", "db_port": "3306", "db_name": "", "db_maxpool": "20", "db_minpool": "5", "db_perfix": "", "slavedb": ""})
	con := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", data["db_user"],
		data["db_password"], data["db_host"],
		data["db_port"], data["db_name"])
	mysqldb, _ = sql.Open("mysql", con)
	maxpool, _ := strconv.Atoi(data["db_maxpool"])
	minpool, _ := strconv.Atoi(data["db_minpool"])
	mysqldb.SetMaxOpenConns(maxpool)
	mysqldb.SetMaxIdleConns(minpool)
	mysqldb.SetConnMaxLifetime(time.Minute * 5)
	mysqldb.Ping()

	Slavedb = make([]*sql.DB, 0)
	iplist := strings.Split(data["slavedb"], ",")
	for _, v := range iplist {
		con1 := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", data["db_user"],
			data["db_password"], v,
			data["db_port"], data["db_name"])
		sqldb1, _ := sql.Open("mysql", con1)
		maxpool, _ := strconv.Atoi(data["db_maxpool"])
		minpool, _ := strconv.Atoi(data["db_minpool"])
		sqldb1.SetMaxOpenConns(maxpool)
		sqldb1.SetMaxIdleConns(minpool)
		sqldb1.SetConnMaxLifetime(time.Minute * 5)
		sqldb1.Ping()
		Slavedb = append(Slavedb, sqldb1)
	}
	Db_Struct.Db_perfix = data["db_perfix"]
	Db_Struct.Db_name = data["db_name"]
	Db_Struct.Db_host = data["db_host"]
	Db_Struct.Db_port = data["db_port"]
	Db_Struct.Db_password = data["db_password"]
	Db_perfix = data["db_perfix"]
	Init_redis_table_struct()
}

func Init_redis_table_struct() {
	qb := new(Mysqlcon)
	if (Is_db_init == false) {
		Is_db_init = true
		data := qb.Query("show TABLES", nil)
		for _, v := range data {
			qb.Dbinit()
			tbname := v["Tables_in_"+Db_Struct.Db_name]
			list := qb.Query("SHOW full COLUMNS FROM "+tbname, nil)
			if (list != nil) {
				data_list := make([]map[string]string, 0)
				for _, val := range list {
					col := make(map[string]string)
					for key, _ := range val {
						col[common.Tolow_map_name(key)] = val[key]
					}

					data_list = append(data_list, col)
				}
				G_dbtables[tbname] = data_list
				tbname = strings.Replace(tbname, Db_perfix, "", -1)
				Get_mysql_dict(tbname)
			}
		}
	}
}

func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s)) //使用zhifeiya名字做散列值，设定后不要变
	return hex.EncodeToString(h.Sum(nil))
}

func Get_mysql_dict(tbname string) {
	db := NewGymysqldb()
	data := db.Tbname("db_tb_dict").Where(fmt.Sprintf("name='%v'", Db_perfix+tbname)).Find();
	if (data == nil) {
		return
	}
	db.Dbinit()
	fd_data := db.Tbname("db_fd_dict").Where(fmt.Sprintf("t_id=%v", data["id"])).Select()
	list_data := db.Tbname("db_fd_dict").Where(fmt.Sprintf("t_id=%v and list_tb_name<>'0'", data["id"])).Select()
	G_tb_dict[tbname] = data

	if (fd_data != nil) {
		G_fd_dict[tbname] = fd_data

	}
	if (list_data != nil) {
		G_fd_list[tbname] = list_data
	}
}

func NewGymysqldb() (*Mysqlcon) {
	this := new(Mysqlcon)
	this.SqlTx = nil
	this.Dbinit()
	return this
}
