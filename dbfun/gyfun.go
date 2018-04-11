package lib

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gylib/common"
	"strconv"
)

var mysqldb *sql.DB

type Db_conn struct {
	db_host     string
	db_port     string
	db_name     string
	db_password string

}

var Db_perfix string
//var DataTable map[string]string

func init() {
	//DataTable=make(map[string]string)
	data := common.Getini("conf/app.ini", "database", map[string]string{"db_user": "root", "db_password": "",
		"db_host": "127.0.0.1", "db_port": "3306", "db_name": "", "db_maxpool": "200", "db_minpool": "100","db_perfix": ""})
	con := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", data["db_user"],
		data["db_password"], data["db_host"],
		data["db_port"], data["db_name"])
	mysqldb, _ = sql.Open("mysql", con)
	maxpool, _ := strconv.Atoi(data["db_maxpool"])
	minpool, _ := strconv.Atoi(data["db_minpool"])
	mysqldb.SetMaxOpenConns(maxpool)
	mysqldb.SetMaxIdleConns(minpool)
	mysqldb.Ping()
	Db_perfix=data["db_perfix"]
}


//type Postdata map[string]interface{}

type Querybuilder interface {
	Start_tran()(*sql.Tx)
	Find() map[string]string
	Select() []map[string]string
	Tbname(name string) Querybuilder
	Where(wherestr string) Querybuilder
	Order(orderstr string) Querybuilder
	Limit(limitstr string) Querybuilder
	Insert(postdata map[string]interface{}) (sql.Result, error)
	Delete() (sql.Result, error)
	Update(postdata map[string]interface{}) (sql.Result, error)
	Query(string) []map[string]string
	Excute(string) (sql.Result, error)
	GetLastSql() string
	Dbinit()
	Count() int64
	Get_where_data(map[string]interface{}) string
	Get_new_add() map[string]string
	Begin_tran([]string) (int)
	Get_Update(map[string]interface{}) (string)
	Get_Insert(map[string]interface{}) (string)
	Join(tbname string, jointype string,where string, fileds string) Querybuilder
	//Get_new_str()map[string]string
	//Table_json()
	Type2str(val interface{}) (string)
	MapContains(src map[string]interface{}, key string) bool
}

//func NewQuerybuilder(dirver string) (qb Querybuilder,err error) {
//	if(dirver=="mysql") {
//		qb = new(Mysqlcon)
//	}else {
//		err = errors.New("unknown driver for query builder")
//	}
//	return
//}

func NewQuerybuilder() (qb Querybuilder) {
	qb = new(Mysqlcon)
	qb.Dbinit()
	return
}

func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s)) //使用zhifeiya名字做散列值，设定后不要变
	return hex.EncodeToString(h.Sum(nil))
}

//func Build_menu(qxz_memo string) map[int]map[string]interface{} {
//	db := NewQuerybuilder()
//	db.Dbinit()
//
//	rows := db.Tbname("nav").Where("is_display=1 and is_tel=0 and parent_id=0").Order("order_number").Getall()
//	result := make(map[int]map[string]interface{}, len(rows))
//	i := 0
//	for _, val := range rows {
//
//		if get_menu_flag(qxz_memo, val["nav_code"]) {
//			result[i] = make(map[string]interface{})
//			result[i]["first"] = val
//			db.Dbinit()
//			secdb := db.Tbname("nav").Where("is_display=1 and is_tel=0 and parent_id=" + val["id"]).Order("order_number").Getall()
//			j := 0
//			thr_arr := make(map[int]map[string]interface{})
//			four_arr := make(map[int]map[string]interface{})
//			for _, sec_val := range secdb {
//				if get_menu_flag(qxz_memo, sec_val["nav_code"]) {
//					db.Dbinit()
//					thr_arr[j] = make(map[string]interface{})
//					for sec_key, sec_value := range sec_val {
//						thr_arr[j][sec_key] = sec_value
//					}
//					k := 0
//					three_db := db.Tbname("nav").Where("is_display=1 and is_tel=0 and parent_id=" + sec_val["id"]).Order("order_number").Getall()
//					for _, thr_val := range three_db {
//						four_arr[k] = make(map[string]interface{})
//						four_arr[k]["thr_data"] = thr_val
//						k++
//					}
//					if len(four_arr) > 0 {
//						thr_arr[j]["is_menu"] = true
//						thr_arr[j]["thr_data"] = four_arr
//					} else {
//						thr_arr[j]["is_menu"] = false
//					}
//
//					j++
//				}
//			}
//			if len(thr_arr) > 0 {
//				result[i]["sec_data"] = thr_arr
//			}
//			i++
//		}
//	}
//
//	return result
//
//}
//
//func Get_All_menu() map[int]map[string]interface{} {
//	db := NewQuerybuilder()
//	db.Dbinit()
//	rows := db.Tbname("nav").Where("parent_id=0").Order("order_number").Getall()
//	result := make(map[int]map[string]interface{})
//	i := 0
//	for _, val := range rows {
//		result[i] = make(map[string]interface{})
//		result[i]["first"] = val
//		db.Dbinit()
//		secdb := db.Tbname("nav").Where("parent_id=" + val["id"]).Order("order_number").Getall()
//		j := 0
//		thr_arr := make(map[int]map[string]interface{})
//		four_arr := make(map[int]map[string]interface{})
//		for _, sec_val := range secdb {
//
//			db.Dbinit()
//			thr_arr[j] = make(map[string]interface{})
//			for sec_key, sec_value := range sec_val {
//				thr_arr[j][sec_key] = sec_value
//			}
//			k := 0
//			three_db := db.Tbname("nav").Where("parent_id=" + sec_val["id"]).Order("order_number").Getall()
//			for _, thr_val := range three_db {
//				four_arr[k] = make(map[string]interface{})
//				four_arr[k]["thr_data"] = thr_val
//				k++
//			}
//			if len(four_arr) > 0 {
//				thr_arr[j]["is_menu"] = true
//				thr_arr[j]["thr_data"] = four_arr
//			} else {
//				thr_arr[j]["is_menu"] = false
//			}
//
//			j++
//		}
//
//		if len(thr_arr) > 0 {
//			result[i]["sec_data"] = thr_arr
//		}
//		i++
//
//	}
//
//	return result
//
//}
//
//func get_menu_flag(qxz_memo string, code string) bool {
//	db := strings.Split(qxz_memo, ",")
//	var flag bool
//	flag = false
//	for _, val := range db {
//		if val == code {
//			flag = true
//			break
//		}
//	}
//	return flag
//}
