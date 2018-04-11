package main

import (
	_ "gylib/common"
	"fmt"
	"gylib/dbfun"
	"gylib/common"
	"time"
	"path"
)

type Nav struct {
	Id          int    `xorm:"not null pk autoincr INT(11)"`
	ParentId    int    `xorm:"default 0 INT(11)"`
	NavName     string `xorm:"VARCHAR(200)"`
	NavCode     string `xorm:"VARCHAR(200)"`
	NavModule   string `xorm:"VARCHAR(200)"`
	NavImage    string `xorm:"VARCHAR(300)"`
	IsDisplay   int    `xorm:"default 0 INT(11)"`
	OrderNumber int    `xorm:"INT(11)"`
	IsTel       int    `xorm:"default 0 INT(11)"`
}

func main() {
	//xorm 自定义查询 结果体map转换测试
	db := lib.NewQuerybuilder()
	where := make(map[string]interface{})
	where["I_name"] = "王"
	w_str := db.Tbname("login").Get_where_data(where)
	postdata := make(map[string]interface{})
	postdata["name"] = "王''不'''清白"
	db.Tbname("login").Where(w_str).Update(postdata)
	//fmt.Println(db.GetLastSql())
	db.Tbname("login").Join("gy_sqwdz", "inner join", ""+
		"yzk_gy_sqwdz.id=yzk_login.w_id", "").Join(""+
		"qxz", "inner join", "yzk_login.qxz=yzk_qxz.name", "yzk_qxz.memo,yzk_qxz.level"+
		"").Where("yzk_login.name like '李%'").Select()
	//fmt.Println(w_str)
	//for _,val :=range rows{
	//	//fmt.Println(val)
	//}
	//fmt.Println(rows)

	values := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	resultChan := make(chan int, 4)
	go sum(values[:len(values)/4], resultChan)
	go sum(values[len(values)/2:], resultChan)
	go sum(values[len(values)/3:], resultChan)
	go sum(values[len(values)/6:], resultChan)
	sum1, sum2, sum3 ,sum4:= <-resultChan, <-resultChan, <-resultChan,<-resultChan
	fmt.Println("Result:", sum1, sum2, sum3,sum4)
	

}

func test() {
	nav := new(Nav)
	db := lib.NewQuerybuilder()
	rows := db.Tbname("nav").Find()
	common.DataToStruct(rows, nav)
	fmt.Println(nav)
	fmt.Println(common.Struct2Map(*nav))
	//时间转换测试
	timestamp := time.Now().Unix()
	fmt.Println(common.Int2Time_str(timestamp))
	fmt.Println(common.String2Time("2006-01-02 15:04:05"))
	fmt.Println(path.Ext("guyi.txt"))
}



func sum(values []int, resultChan chan int) {
	sum := 0
	fmt.Println(values)
	for _, value := range values {
		sum += value
	}
	// 将计算结果发送到channel中
	resultChan <- sum

}
