package mymaindb

import (
	"database/sql"
	"gylib/common"
	"fmt"
	"strings"
	"gylib/common/datatype"
	"math/rand"
	"strconv"
	"time"
)

type Mysqlcon struct {
	SqlTx       *sql.Tx
	Tablename   string
	Sql_where   string
	Sql_order   string
	Sql_fields  string
	Sql_limit   string
	Db_perfix   string
	Db_name     string
	Query_data  []map[string]interface{}
	Join_arr    map[string]string
	LastSqltext string
}

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

func Get_mysql_dict(tbname string) {
	db := NewMainDB()
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

func NewMainDB() (*Mysqlcon) {
	this := new(Mysqlcon)
	this.SqlTx = nil
	this.Dbinit()
	return this
}

func (this *Mysqlcon) Merge_And_where(where_str, new_str string) (string) {
	result := where_str
	if (where_str != "") {
		result += " and " + new_str
	} else {
		result = new_str
	}
	return result
}

func (this *Mysqlcon) Merge_OR_where(where_str, new_str string) (string) {
	result := where_str
	if (where_str != "") {
		result += " or " + new_str
	} else {
		result = new_str
	}
	return result
}

func (this *Mysqlcon) BeginStart() (bool) {
	tx, err := mysqldb.Begin()
	if (err != nil) {
		return false
	}
	this.SqlTx = tx
	return true
}

/**
初始化结构
*/
func (this *Mysqlcon) Dbinit() {
	this.Tablename = ""
	this.Sql_limit = ""
	this.Sql_order = ""
	this.Sql_fields = ""
	this.Sql_where = ""
	this.Db_perfix = Db_perfix
	this.Join_arr = make(map[string]string)
	this.Query_data = make([]map[string]interface{}, 0)
}

/*
设置数据表
*/
func (this *Mysqlcon) Tbname(name string) (*Mysqlcon) {
	//data := common.Getini("conf/app.ini","database",map[string]string{"db_perfix":""})
	this.Dbinit()
	this.Tablename = this.Db_perfix + name
	return this
}

func (this *Mysqlcon) Where(where interface{}) (*Mysqlcon) {
	//kk:= reflect.TypeOf(where)
	//fmt.Println(kk)
	switch where.(type) {
	case string:
		if this.Sql_where == "" {
			this.Sql_where = where.(string)
		} else {
			this.Sql_where += " and (" + where.(string) + ")"
		}
	default:
		tmp_arr := where.(map[string]interface{})
		if (len(tmp_arr) > 0) {
			this.Query_data = append(this.Query_data, tmp_arr)
		}
		//fmt.Println("query_data", this.Query_data)
	}

	return this
}

func (this *Mysqlcon) Order(orderstr string) (*Mysqlcon) {
	this.Sql_order = orderstr
	return this
}

func (this *Mysqlcon) Limit(limitstr string) (*Mysqlcon) {
	this.Sql_limit = limitstr
	return this
}

func (this *Mysqlcon) MapContains(src map[string]interface{}, key string) bool {
	if _, ok := src[key]; ok {
		return true
	}
	return false
}

func (this *Mysqlcon) Get_read_dbcon() (*sql.DB) {
	read_ct := len(Slavedb)
	//fmt.Println("read_ct",read_ct)
	if (read_ct == 0) {
		return mysqldb
	} else {
		result := rand.Intn(read_ct)
		//fmt.Println("readcon",result)
		return Slavedb[result]

	}
}

func (this *Mysqlcon) MapContains_str(src map[string]string, key string) bool {
	if _, ok := src[key]; ok {
		return true
	}
	return false
}

/*启动事务，返回事物指针*/
func (this *Mysqlcon) Start_tran() (*sql.Tx) {
	tx, err := mysqldb.Begin()
	if (err != nil) {
		return nil
	}
	return tx
}

func (this *Mysqlcon) Check_data_fields(fieldname string) (bool) {
	flag := false
	fd_list, ok := G_dbtables[this.Tablename]
	if (ok) {
		for _, v := range fd_list.([]map[string]string) {
			record := v
			if (record["key"] == "PRI" && record["extra"] == "auto_increment") {
				continue
			}
			if (record["field"] == fieldname) {
				flag = true
				break
			}

		}
		return flag
	} else {
		this.Update_redis(this.Tablename)
		rows, _ := mysqldb.Query("SHOW full COLUMNS FROM " + this.Tablename)
		if (rows == nil) {
			return false;
		}
		defer rows.Close()
		columns, _ := rows.Columns()
		scanArgs := make([]interface{}, len(columns))
		values := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		for rows.Next() {
			//将行数据保存到record字典
			record := make(map[string]string)
			_ = rows.Scan(scanArgs...)
			for i, col := range values {
				if col != nil {
					record[strings.ToLower(columns[i])] = this.Type2str(col)
				}
			}
			if (record["key"] == "PRI" && record["extra"] == "auto_increment") {
				continue
			}

			if (record["field"] == fieldname) {
				flag = true
				break
			}

		}
		return flag
	}
}

func (this *Mysqlcon) Type2str(val interface{}) (string) {
	//fmt.Println(fmt.Sprintf("%T,%v",val,val))
	var result string = ""
	switch val.(type) {
	case []string:
		strArray := val.([]string)
		result = strings.Join(strArray, "")
	case []uint8:
		result = string(val.([]uint8))
	default:
		result = fmt.Sprintf("%v", val)
	}
	return result
}

func (this *Mysqlcon) Insert(postdata map[string]interface{}) (sql.Result, error) {
	var sqltext string
	sqltext = "insert into " + this.Tablename + " ("
	values := " values ("
	i := 0
	param_data := make([]interface{}, 0)
	for k, v := range postdata {
		if (this.Check_data_fields(k) == false) {
			continue
		}
		if (i > 0) {
			sqltext += ","
			values += ","
		}
		i++
		sqltext += "`" + k + "`"
		values += " ? "
		param_data = append(param_data, v)
	}
	sqltext += ") " + values + ")"
	this.LastSqltext = sqltext
	//fmt.Println(i,sqltext)
	//fmt.Println(len(param_data),param_data)
	var result sql.Result
	var err error
	if (this.SqlTx != nil) {
		result, err = this.SqlTx.Exec(sqltext, param_data...)
	} else {
		result, err = mysqldb.Exec(sqltext, param_data...)
	}
	//fmt.Println(err)
	return result, err

}

func (this *Mysqlcon) Update(postdata map[string]interface{}) (sql.Result, error) {
	var sqltext string
	sqltext = fmt.Sprintf("update %v set ", this.Tablename)
	i := 0
	param_data := make([]interface{}, 0)
	for k, v := range postdata {
		if (this.Check_data_fields(k) == false) {
			continue
		}
		if (i > 0) {
			sqltext += ","

		}
		i++
		sqltext += "`" + k + "`" + "= ?"
		param_data = append(param_data, v)
	}
	sqlwhere, param := this.Build_where()
	for _, v := range param {
		param_data = append(param_data, v)
	}
	sqltext += sqlwhere
	this.LastSqltext = sqltext
	//fmt.Println(sqltext, param_data)
	var result sql.Result
	var err error
	if (this.SqlTx != nil) {
		result, err = this.SqlTx.Exec(sqltext, param_data...)
	} else {
		result, err = mysqldb.Exec(sqltext, param_data...)
	} //fmt.Println(err)
	return result, err
}

func (this *Mysqlcon) Delete() (sql.Result, error) {
	sqlwhere, param := this.Build_where()
	sqltext := fmt.Sprintf(" delete from %v %v", this.Tablename, sqlwhere)
	this.LastSqltext = sqltext
	var result sql.Result
	var err error
	if (this.SqlTx != nil) {
		result, err = this.SqlTx.Exec(sqltext, param...)
	} else {
		result, err = mysqldb.Exec(sqltext, param...)
	}
	return result, err
}

func (this *Mysqlcon) SetDec(fdname string, quantity int) (sql.Result, error) {
	sqltext := fmt.Sprintf("update %v set %v=%v-%v", this.Tablename, fdname, fdname, quantity)
	sqlwhere, param := this.Build_where()
	sqltext += sqlwhere
	this.LastSqltext = sqltext
	var result sql.Result
	var err error
	if (this.SqlTx != nil) {
		result, err = this.SqlTx.Exec(sqltext, param...)
	} else {
		result, err = mysqldb.Exec(sqltext, param...)
	}
	return result, err
}

func (this *Mysqlcon) SetInc(fdname string, quantity int) (sql.Result, error) {
	sqlwhere, param := this.Build_where()
	sqltext := fmt.Sprintf("update %v set %v=%v+%v  %v", this.Tablename, fdname, fdname, quantity, sqlwhere)
	this.LastSqltext = sqltext
	var result sql.Result
	var err error
	if (this.SqlTx != nil) {
		result, err = this.SqlTx.Exec(sqltext, param...)
	} else {
		result, err = mysqldb.Exec(sqltext, param...)
	}
	return result, err
}

func (this *Mysqlcon) Query(sqltext string, param []interface{}) []map[string]string {
	this.LastSqltext = sqltext
	var rows *sql.Rows
	var err error
	if (this.SqlTx != nil) {
		rows, err = this.SqlTx.Query(sqltext, param...)
	} else {
		sqldbcon := this.Get_read_dbcon()
		rows, err = sqldbcon.Query(sqltext, param...)
	}
	if (err != nil) {
		return nil
	}
	defer rows.Close()
	//字典类型
	//构造scanArgs、values两个数组，scanArgs的每个值指向values相应值的地址
	columns, _ := rows.Columns()
	result := make([]map[string]string, 0)
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		//将行数据保存到record字典
		record := make(map[string]string)
		_ = rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = datatype.Type2str(col)
			} else {
				record[columns[i]] = ""
			}
		}
		result = append(result, record)
	}
	//fmt.Print(result)
	if (len(result) == 0) {
		return nil
	}
	return result
}

func (this *Mysqlcon) Excute(sqltext string, param []interface{}) (sql.Result, error) {
	this.LastSqltext = sqltext
	var result sql.Result
	var err error
	if (this.SqlTx != nil) {
		result, err = this.SqlTx.Exec(sqltext, param...)
	} else {
		result, err = mysqldb.Exec(sqltext, param...)
	}
	return result, err
}

//
//func (this *Mysqlcon) delete(fields ...string) (*Mysqlcon) {
//	this.Tokens = append(this.Tokens, "SELECT", strings.Join(fields,","))
//	return qb
//}
func (this *Mysqlcon) Join(tbname string, jointype string, where string, fileds string) (*Mysqlcon) {
	if (this.Join_arr["tbname"] == "") {
		this.Join_arr["tbname"] = this.Tablename + " " + jointype + " " + Db_perfix + tbname + " on " + where
		if (fileds != "") {
			this.Join_arr["fields"] = this.Tablename + ".*," + fileds
		} else {
			this.Join_arr["fields"] = this.Tablename + ".*"
		}
	} else {
		this.Join_arr["tbname"] += " " + jointype + " " + Db_perfix + tbname + " on " + where
		if (fileds != "") {
			this.Join_arr["fields"] += "," + fileds
		}
	}

	return this
}

func (this *Mysqlcon) set_sql(flag int) (string) {
	sqltext := ""
	if (flag == 0) {
		if (this.MapContains_str(this.Join_arr, "tbname")) {
			if (this.Join_arr["fields"] != "") {
				sqltext = "select " + this.Join_arr["fields"] + " from " + this.Join_arr["tbname"]
			} else {
				sqltext = "select " + this.Tablename + ".* from " + this.Tablename
			}
		} else {
			sqltext = "select  * from " + this.Tablename
		}
	} else {
		if (this.MapContains_str(this.Join_arr, "tbname")) {
			sqltext = "select count(" + this.Tablename + ".*) as ct " + " from " + this.Join_arr["tbname"]
		} else {
			sqltext = "select count(*) as ct from " + this.Tablename
		}
	}
	return sqltext
}

func (this *Mysqlcon) Build_where() (string, []interface{}) {
	is_where := false
	sqltext := ""
	if this.Sql_where != "" {
		sqltext += " where " + this.Sql_where
		is_where = true
	}
	param_data := make([]interface{}, 0)
	if (len(this.Query_data) > 0) {
		if (is_where) {
			sqltext += " and "

		} else {
			sqltext += " where "
		}
		i := 0
		for _, v := range this.Query_data {
			for key, val := range v {
				//if (this.Check_data_fields(key) == false) {
				//	continue
				//}
				if (i > 0) {
					sqltext += " and "
				}
				i++
				switch val.(type) {
				//data["name"]=" %v like ?"
				//data["name"]=" %v>=(?)"
				//data["name"]="locate(?,`"+this.Tablename+"`.`%v`)>0"
				case map[string]interface{}:
					param_data = append(param_data, val.(map[string]interface{})["value"])
					sqltext += datatype.Type2str(val.(map[string]interface{})["name"])
				default:
					param_data = append(param_data, val)
					sqltext += key + "=(?) "

				}
			}
		}

	}
	return sqltext, param_data
}

func (this *Mysqlcon) Find() map[string]string {
	sqltext := this.set_sql(0)
	param_data := make([]interface{}, 0)
	tmpstr := ""
	tmpstr, param_data = this.Build_where()
	sqltext += tmpstr
	if this.Sql_order != "" {
		sqltext += " order by " + this.Sql_order
	}
	this.LastSqltext = sqltext + " limit 1"
	var rows *sql.Rows
	var err error
	if (this.SqlTx != nil) {
		rows, err = this.SqlTx.Query(sqltext+" limit 1", param_data...)
	} else {
		sqldbcon := this.Get_read_dbcon()
		rows, err = sqldbcon.Query(sqltext+" limit 1", param_data...)
	}
	//fmt.Println("rows",rows,err)
	if (err != nil) {
		return nil
	}
	if (rows == nil) {
		return nil
	}

	defer rows.Close()
	//字典类型
	//构造scanArgs、values两个数组，scanArgs的每个值指向values相应值的地址
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	record := make(map[string]string)
	for rows.Next() {
		//将行数据保存到record字典
		_ = rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = datatype.Type2str(col)

			} else {
				record[columns[i]] = ""
			}
		}

	}
	if (len(record) == 0) {
		return nil
	}
	return record
}

func (this *Mysqlcon) Count() int64 {
	sqltext := this.set_sql(1)
	sqlwhere, param := this.Build_where()
	sqltext += sqlwhere
	this.LastSqltext = sqltext
	sqldbcon := this.Get_read_dbcon()
	rows := sqldbcon.QueryRow(sqltext, param...)
	var record int64
	rows.Scan(&record)

	return record
}

func (this *Mysqlcon) Sum(fd string) (float64) {
	var result float64
	sqltext := this.set_sql(1)
	sqltext = strings.Replace(sqltext, "count(*)", "sum("+fd+")", -1)
	sqlwhere, param := this.Build_where()
	sqltext += sqlwhere
	this.LastSqltext = sqltext
	sqldbcon := this.Get_read_dbcon()
	rows := sqldbcon.QueryRow(sqltext, param...)
	rows.Scan(&result)
	return result
}

func (this *Mysqlcon) Select() []map[string]string {
	sqltext := this.set_sql(0)
	sqlwhere, param := this.Build_where()
	sqltext += sqlwhere
	if this.Sql_order != "" {
		sqltext += " order by " + this.Sql_order
	}
	if this.Sql_limit != "" {
		sqltext += " limit " + this.Sql_limit
	}
	this.LastSqltext = sqltext
	//fmt.Println(sqltext)
	sqldbcon := this.Get_read_dbcon()
	rows, err := sqldbcon.Query(sqltext, param...)
	if (err != nil) {
		return nil
	}
	if (rows == nil) {
		return nil
	}

	defer rows.Close()

	//字典类型
	//构造scanArgs、values两个数组，scanArgs的每个值指向values相应值的地址
	columns, _ := rows.Columns()
	result := make([]map[string]string, 0)
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	j := 0
	for rows.Next() {
		//将行数据保存到record字典
		record := make(map[string]string)
		_ = rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = this.Type2str(col)
				//record[columns[i]] = col.([]byte)
			} else {
				record[columns[i]] = ""
			}
		}
		result = append(result, record)
		//result[j] = record
		j++

	}
	if (len(result) == 0) {
		return nil
	}
	return result
}

func (this *Mysqlcon) GetLastSql() string {
	return this.LastSqltext
}

func (this *Mysqlcon) Get_new_add() map[string]string {
	fd_list, ok := G_dbtables[this.Tablename]
	if (ok) {
		//fmt.Println(fd_list)
		result := make(map[string]string)
		for _, v := range fd_list.([]map[string]string) {
			fd_name := v["field"]
			result[fd_name] = ""
		}
		return result
	} else {
		this.Update_redis(this.Tablename)
		rows, _ := mysqldb.Query("SHOW full COLUMNS FROM " + this.Tablename)
		defer rows.Close()
		columns, _ := rows.Columns()
		scanArgs := make([]interface{}, len(columns))
		values := make([]interface{}, len(columns))
		result := make(map[string]string)
		for i := range values {
			scanArgs[i] = &values[i]
		}
		for rows.Next() {
			//将行数据保存到record字典
			record := make(map[string]string)
			_ = rows.Scan(scanArgs...)
			for i, col := range values {
				if col != nil {
					record[strings.ToLower(columns[i])] = string(col.([]byte))
					result[record["field"]] = ""
				}
			}
		}

		return result
	}
}

func (this *Mysqlcon) Update_redis(tbname string) {

	list := this.Query("SHOW full COLUMNS FROM "+tbname, nil)
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
	}
	//this.Dbinit()
}

func (this *Mysqlcon) Get_fields_sql(fd_name, val_name string) (result string) {
	fd_list, ok := G_dbtables[this.Tablename]
	if (ok) {
		for _, v := range fd_list.([]map[string]string) {
			record := v
			if (fd_name == record["field"]) {
				result = "`" + record["field"] + "`=" + this.checkstr(record["type"], val_name)
				break
			}
		}
	}

	return result

}

func (this *Mysqlcon) checkstr(fdtype string, fdvalue string) (string) {
	if (fdvalue == "") {
		return "null"
	}
	if (strings.Contains(fdtype, "tinyint") ||
		strings.Contains(fdtype, "double") ||
		strings.Contains(fdtype, "float") ||
		strings.Contains(fdtype, "int") ||
		strings.Contains(fdtype, "decimal")) {
		return fdvalue
	} else {
		//result :=strings.Replace(fdvalue, "\\", "\\\\", -1)
		//result = "'" + strings.Replace(result, "'", "\\'", -1) + "'"
		result := "'" + strings.Replace(fdvalue, "'", "\\'", -1) + "'"
		return result
	}

}

func (this *Mysqlcon) Get_select_data(d_data map[string]string, masterdb string) (map[string]string) {
	data, ok := G_fd_list[masterdb]
	if (ok) {

		for _, v := range data.([]map[string]string) {
			listname := strings.Replace(v["list_tb_name"], this.Db_perfix, "", -1)
			tbname := strings.Replace(v["list_tb_name"], this.Db_perfix, "", -1)
			listname = strings.Replace(listname, "_", "", -1)
			where := v["list_where"]
			list_val := v["list_val"]
			list_display := datatype.Type2str(v["list_display"])
			if (where != "") {
				where += " and " + this.Tbname(tbname).Get_fields_sql(list_val, d_data[v["name"]])
			} else {
				where = this.Tbname(tbname).Get_fields_sql(list_val, d_data[v["name"]])
			}
			list_data := this.Tbname(tbname).Where(where).Find()
			//fmt.Println(v,this.GetLastSql())
			//fmt.Println(list_data)
			if (list_data != nil) {
				d_data[v["name"]+"_name"] = list_data[list_display]
			} else {
				d_data[v["name"]+"_name"] = ""
			}
		}
	}
	//fmt.Println(d_data)
	return d_data
}

func (this *Mysqlcon) Get_where_data(postdata map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, val := range postdata {
		val_str := strings.TrimSpace(this.Type2str(val))
		if val_str != "" {
			if strings.Contains(key, "S_") {
				key1 := strings.Replace(key, "S_", "", -1)
				result[key1] = val_str
			}

			if strings.Contains(key, "I_") {
				key1 := strings.Replace(key, "I_", "", -1)
				result[key1] = map[string]interface{}{"name": "locate(?,`" + this.Tablename + "`.`" + key1 + "`)>0", "value": val_str}
			}
		}
	}
	return (result)
}

func (this *Mysqlcon) Rollback() {
	if (this.SqlTx == nil) {
		return
	}
	this.SqlTx.Rollback()
	this.SqlTx = nil
}

func (this *Mysqlcon) Commit() {
	if (this.SqlTx == nil) {
		return
	}
	this.SqlTx.Commit()
	this.SqlTx = nil
}
