package lib

import (
	"database/sql"
	"fmt"
	"strings"
	"gylib/common"
	"gylib/common/datatype"
)

type Mysqlcon struct {
	Tablename   string
	Sql_where   string
	Sql_order   string
	Sql_fields  string
	Sql_limit   string
	Db_perfix   string
	Join_arr    map[string]string
	LastSqltext string
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
}

/*
设置数据表
*/
func (this *Mysqlcon) Tbname(name string) Querybuilder {
	//data := common.Getini("conf/app.ini","database",map[string]string{"db_perfix":""})
	this.Tablename = this.Db_perfix + name
	return this
}

func (this *Mysqlcon) Where(where string) Querybuilder {
	if this.Sql_where == "" {
		this.Sql_where = where
	} else {
		this.Sql_where += " and (" + where + ")"
	}
	return this
}

func (this *Mysqlcon) Order(orderstr string) Querybuilder {
	this.Sql_order = orderstr
	return this
}

func (this *Mysqlcon) Limit(limitstr string) Querybuilder {
	this.Sql_limit = limitstr
	return this
}

func (this *Mysqlcon) MapContains(src map[string]interface{}, key string) bool {
	if _, ok := src[key]; ok {
		return true
	}
	return false
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

func (this *Mysqlcon) Begin_tran(sqlstr []string) (int) {
	tx, err := mysqldb.Begin()
	if (err != nil) {
		return 0
	}
	for _, sqltext := range sqlstr {
		_, err = tx.Exec(sqltext)
		if (err != nil) {
			tx.Rollback()
			return 0
		}
	}
	tx.Commit()
	return 1
}

func (this *Mysqlcon) Get_key_eq_value(id string) (string) {
	tbname:=this.Tablename
	result := ""
	fd_list,ok:=G_dbtables[tbname]
	if (ok) {
		for _, v := range fd_list.([]map[string]string) {
			record := v
			if (record["key"] == "PRI") {
				result = record["field"] + "=" + this.checkstr(record["type"], id)
				break
			}
		}
		return result
	} else {
		this.Update_redis(tbname)

		rows, err := mysqldb.Query("SHOW full COLUMNS FROM " + tbname)
		//fmt.Println(rows)
		if(err!=nil){
			fmt.Println("SHOW full COLUMNS FROM " + tbname)
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
					record[strings.ToLower(columns[i])] = string(col.([]byte))
				}
			}
			if (record["key"] == "PRI") {
				result = record["field"] + "=" + this.checkstr(record["type"], id)
				break;
			}
		}
		return result
	}
}

func (this *Mysqlcon) Get_key_in_value(id string) (string) {
	result := ""
	fd_list,ok := G_dbtables[this.Tablename]
	if (ok) {
		for _, v := range fd_list.([]map[string]string) {
			record := v
			if (record["key"] == "PRI") {
				result = record["field"] + " in (" + this.set_in_where(record["type"], id) + ")"
				break
			}
		}
		return result
	} else {
		this.Update_redis(this.Tablename)
		rows, _ := mysqldb.Query("SHOW full COLUMNS FROM " + this.Tablename)
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
					record[strings.ToLower(columns[i])] = string(col.([]byte))
				}
			}
			if (record["key"] == "PRI") {
				result = record["field"] + " in (" + this.set_in_where(record["type"], id) + ")"
				break;
			}
		}
		return result
	}
}

func (this *Mysqlcon) get_insert_sql(postdata map[string]interface{}) (result string, val string) {

	fd_list,ok := G_dbtables[this.Tablename]
	if (ok) {
		for _, v := range fd_list.([]map[string]string) {
			record := v
			if (record["key"] == "PRI" && record["extra"] == "auto_increment") {
				continue
			}

			if this.MapContains(postdata, record["field"]) == false {
				continue
			}
			val_str := this.Type2str(postdata[record["field"]])
			if result == "" {
				result = "`" + record["field"] + "`"
				val = this.checkstr(record["type"], val_str)

			} else {
				result += ",`" + record["field"] + "`"
				val += "," + this.checkstr(record["type"], val_str)
			}

		}
		return result, val
	} else {
		this.Update_redis(this.Tablename)
		rows, _ := mysqldb.Query("SHOW full COLUMNS FROM " + this.Tablename)
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
					record[strings.ToLower(columns[i])] = string(col.([]byte))
				}
			}
			if (record["key"] == "PRI" && record["extra"] == "auto_increment") {
				continue
			}

			if this.MapContains(postdata, record["field"]) == false {
				continue
			}
			val_str := this.Type2str(postdata[record["field"]])
			if result == "" {
				result = "`" + record["field"] + "`"
				val = this.checkstr(record["type"], val_str)

			} else {
				result += ",`" + record["field"] + "`"
				val += "," + this.checkstr(record["type"], val_str)
			}

		}
		return result, val
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

func (this *Mysqlcon) Get_fields_sql(fd_name, val_name string) (result string) {

	fd_list,ok :=G_dbtables[this.Tablename]
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

func (this *Mysqlcon) get_update_sql(postdata map[string]interface{}) (result string) {

	fd_list,ok := G_dbtables[this.Tablename]
	if (ok) {
		for _, v := range fd_list.([]map[string]string) {
			record := v
			if (record["key"] == "PRI" && record["extra"] == "auto_increment") {
				continue
			}
			if this.MapContains(postdata, record["field"]) == false {
				continue
			}
			val_str := this.Type2str(postdata[record["field"]])
			if result == "" {
				result = "`" + record["field"] + "`=" + this.checkstr(record["type"], val_str)
			} else {
				result += ",`" + record["field"]+ "`=" + this.checkstr(record["type"], val_str)
			}
		}
		return result
	} else {
		this.Update_redis(this.Tablename)
		rows, _ := mysqldb.Query("SHOW full COLUMNS FROM " + this.Tablename)
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
					record[strings.ToLower(columns[i])] = string(col.([]byte))
				}
			}
			if (record["key"] == "PRI" && record["extra"] == "auto_increment") {
				continue
			}
			if this.MapContains(postdata, record["field"]) == false {
				continue
			}
			val_str := this.Type2str(postdata[record["field"]])

			//if (val_str == "") {
			//	continue
			//}
			if result == "" {
				result = "`" + record["field"] + "`=" + this.checkstr(record["type"], val_str)
			} else {
				result += ",`" + record["field"] + "`=" + this.checkstr(record["type"], val_str)
			}
		}
		return result
	}
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

func (this *Mysqlcon) set_in_where(fdtype string, fdvalue string) (string) {
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
		arr := strings.Split(fdvalue, ",")
		result := ""
		for _, v := range arr {
			if (result == "") {
				result = "'" + strings.Replace(v, "'", "\\'", -1) + "'"
			} else {
				result += ",'" + strings.Replace(v, "'", "\\'", -1) + "'"
			}
		}
		return result
	}

}

func (this *Mysqlcon) Insert(postdata map[string]interface{}) (sql.Result, error) {
	var sqltext string
	fields, value := this.get_insert_sql(postdata)
	sqltext = fmt.Sprintf("insert into %v (%v) values (%v) ", this.Tablename, fields, value)
	//fmt.Println(sqltext)
	this.LastSqltext = sqltext
	result, err := mysqldb.Exec(sqltext)
	return result, err

}

func (this *Mysqlcon) Delete() (sql.Result, error) {
	sqltext := fmt.Sprintf(" delete from %v where %v", this.Tablename, this.Sql_where)
	this.LastSqltext = sqltext
	result, err := mysqldb.Exec(sqltext)
	return result, err
}

func(this *Mysqlcon) SetDec(fdname string,quantity int)(sql.Result,error){
	sqltext := fmt.Sprintf("update %v set %v=%v-%v where %v", this.Tablename, fdname,fdname,quantity,this.Sql_where)
	this.LastSqltext = sqltext
	result, err := mysqldb.Exec(sqltext)
	return result, err
}

func(this *Mysqlcon) SetInc(fdname string,quantity int)(sql.Result,error){
	sqltext := fmt.Sprintf("update %v set %v=%v+%v where %v", this.Tablename, fdname,fdname,quantity,this.Sql_where)
	this.LastSqltext = sqltext
	result, err := mysqldb.Exec(sqltext)
	return result, err
}


func (this *Mysqlcon) Update(postdata map[string]interface{}) (sql.Result, error) {
	sqltext := this.get_update_sql(postdata)
	sqltext = fmt.Sprintf("update %v set %v where %v", this.Tablename, sqltext, this.Sql_where)
	this.LastSqltext = sqltext
	//fmt.Println(sqltext)
	//fmt.Println(postdata)
	result, err := mysqldb.Exec(sqltext)
	return result, err
}

func (this *Mysqlcon) Get_Insert(postdata map[string]interface{}) (string) {
	var sqltext string
	fields, value := this.get_insert_sql(postdata)
	sqltext = fmt.Sprintf("insert into %v (%v) values (%v) ", this.Tablename, fields, value)
	//fmt.Println(sqltext)
	this.LastSqltext = sqltext
	return sqltext

}

func (this *Mysqlcon) Get_Update(postdata map[string]interface{}) (string) {
	sqltext := this.get_update_sql(postdata)
	sqltext = fmt.Sprintf("update %v set %v where %v", this.Tablename, sqltext, this.Sql_where)
	this.LastSqltext = sqltext
	return sqltext
}

func (this *Mysqlcon) Query(sqltext string) []map[string]string {
	this.LastSqltext = sqltext
	rows, err := mysqldb.Query(sqltext)
	if(err!=nil){
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
				record[columns[i]] = string(col.([]byte))
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

func (this *Mysqlcon) Excute(sqltext string) (sql.Result, error) {
	this.LastSqltext = sqltext
	result, err := mysqldb.Exec(sqltext)
	return result, err
}

//
//func (this *Mysqlcon) delete(fields ...string) Querybuilder {
//	this.Tokens = append(this.Tokens, "SELECT", strings.Join(fields,","))
//	return qb
//}
func (this *Mysqlcon) Join(tbname string, jointype string, where string, fileds string) Querybuilder {
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

func (this *Mysqlcon) Find() map[string]string {
	sqltext := this.set_sql(0)
	if this.Sql_where != "" {
		sqltext += " where " + this.Sql_where
	}
	if this.Sql_order != "" {
		sqltext += " order by " + this.Sql_order
	}
	this.LastSqltext = sqltext + " limit 1"
	rows, err := mysqldb.Query(sqltext + " limit 1")
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
				record[columns[i]] = string(col.([]byte))

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
	if this.Sql_where != "" {
		sqltext += " where " + this.Sql_where
	}
	rows := mysqldb.QueryRow(sqltext)
	var record int64
	rows.Scan(&record)

	return record
}

func (this *Mysqlcon) Sum(fd string) (float64) {
	var result float64
	sqltext := this.set_sql(1)
	sqltext = strings.Replace(sqltext, "count(*)", "sum("+fd+")", -1)
	if this.Sql_where != "" {
		sqltext += " where " + this.Sql_where
	}
	rows := mysqldb.QueryRow(sqltext)
	rows.Scan(&result)
	return result
}

func (this *Mysqlcon) Select() []map[string]string {
	sqltext := this.set_sql(0)
	if this.Sql_where != "" {
		sqltext += " where " + this.Sql_where
	}
	if this.Sql_order != "" {
		sqltext += " order by " + this.Sql_order
	}
	if this.Sql_limit != "" {
		sqltext += " limit " + this.Sql_limit
	}
	this.LastSqltext = sqltext
	//fmt.Println(sqltext)
	rows, err := mysqldb.Query(sqltext)
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
				record[columns[i]] = string(col.([]byte))
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

func (this *Mysqlcon) Get_where_data(postdata map[string]interface{}) string {
	var result string
	for key, val := range postdata {
		//strArray := val.([]string)
		//val_str := strings.Join(strArray, "")
		val_str := strings.TrimSpace(this.Type2str(val))
		if val_str != "" {
			if strings.Contains(key, "S_") {
				key1 := strings.Replace(key, "S_", "", -1)
				postdatakey := make(map[string]interface{})
				postdatakey[key1] = val_str
				if result == "" {
					result = this.get_update_sql(postdatakey)
				} else {

					result += " and " + this.get_update_sql(postdatakey)
				}
				//fmt.Println(postdatakey, key1)
			}

			if strings.Contains(key, "I_") {
				val_str = strings.Replace(val_str, "'", "\\'", -1)
				key1 := strings.Replace(key, "I_", "", -1)
				//if result == "" {
				//	result = this.Tablename+"."+key1 + " like '%" + val_str + "%'"
				//} else {
				//	result += " and " + this.Tablename+"."+key1 + " like '%" + val_str + "%'"
				//}
				if result == "" {
					result = "locate('" + val_str + "'," + this.Tablename + "." + key1 + ")>0"
				} else {
					result += " and locate('" + val_str + "'," + this.Tablename + "." + key1 + ")>0"
				}
			}
		}
	}
	return (result)
}

func (this *Mysqlcon) Get_new_add() map[string]string {

	fd_list,ok := G_dbtables[this.Tablename]
	if (ok) {
		//fmt.Println(fd_list)
		result := make(map[string]string)
		for _, v := range fd_list.([]map[string]string) {
			fd_name:=v["field"]
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
	list := this.Query("SHOW full COLUMNS FROM "+tbname)
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
}

func (this *Mysqlcon) Get_select_data(d_data map[string]string,masterdb string) (map[string]string) {
	data,ok := G_fd_list[masterdb]
	if (ok) {
		for _, v := range data.([]map[string]string) {
			listname := strings.Replace(v["list_tb_name"], this.Db_perfix, "", -1)
			tbname := strings.Replace(v["list_tb_name"], this.Db_perfix, "", -1)
			listname = strings.Replace(listname, "_", "", -1)
			this.Dbinit()
			where := v["list_where"]
			list_val := v["list_val"]
			list_display := datatype.Type2str(v["list_display"])
			if (where != "") {
				where += " and " + this.Tbname(tbname).Get_fields_sql(list_val, d_data[v["name"]])
			} else {
				where = this.Tbname(tbname).Get_fields_sql(list_val, d_data[v["name"]])
			}
			this.Dbinit()
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

//func (this *Mysqlcon)Table_json(){
//	rows:=this.Query("SHOW full COLUMNS FROM " + this.Tablename)
//	json_str,_:=json.Marshal(rows)
//	DataTable[this.Tablename]=string(json_str)
//}
//
//func (this *Mysqlcon) Get_new_str()map[string]string{
//	table_str,ok:=DataTable[this.Tablename]
//	if(!ok){
//		this.Table_json()
//		table_str=DataTable[this.Tablename]
//	}
//	rows:=make([]map[string]string,0)
//	json.Unmarshal([]byte(table_str),&rows)
//	record := make(map[string]string)
//	result := make(map[string]string)
//	for _,val:=range rows{
//		for i, col := range val {
//			if(strings.Trim(col,"")==""){
//				continue
//			}
//			if i!="" && col != "" {
//				//fmt.Print("%h",col)
//				record[strings.ToLower(i)] = col
//				result[record["field"]] = "0"
//			}
//		}
//	}
//	return result
//}


