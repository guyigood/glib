package mqmuitask


import (
	"gylib/common/rbmq"
	"fmt"
	"gylib/common/datatype"
	"sync"
	"gylib/common"
	"github.com/silenceper/pool"
	"time"
)

//var Rbcon *WebRbMQ
var mq_data map[string]string
//var Pmq *pool.ObjectPool
//var  ArbmqPool pool.Pool

type WebRbMQ struct {
	Slock sync.Mutex
	Conn  *rbmq.Rbmq
	Data  map[string]interface{}
}
var MqPool pool.Pool

func init() {
	mq_data = common.Getini("conf/app.ini", "amqp", map[string]string{"queuename": "webback", "exname": "", "routekey": "", "max_pool": "100", "min_pool": "5"})
	factory := func() (interface{}, error) { return PoolWebMq() }
	close := func(v interface{}) error { return v.(*WebRbMQ).Close() }
	min_pool := datatype.Str2Int(mq_data["min_pool"])
	max_pool := datatype.Str2Int(mq_data["max_pool"])
	//time_out:=datatype.Str2Int(mq_data["timeout"])

	poolConfig := &pool.Config{
		InitialCap: min_pool,
		MaxCap:     max_pool,
		Factory:    factory,
		Close:      close,
		//链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
		IdleTimeout: 180* time.Second,
	}
	var err error
	MqPool, err = pool.NewChannelPool(poolConfig)
	if err != nil {
		fmt.Println("err=", err)
	}
}

func GetPoolWebMq()(*WebRbMQ, error){
	this, err := MqPool.Get()
	if (err != nil) {
		fmt.Println(err)
		return nil,err
	}
	wbmq := this.(*WebRbMQ)
	return wbmq,nil
}

func PoolWebMq() (*WebRbMQ, error) {
	//mq_data = common.Getini("conf/app.ini", "amqp", map[string]string{"queuename": "crmback", "exname": "", "routekey": "", "max_pool": "100", "min_pool": "5"})
	this := new(WebRbMQ)
	this.Conn = rbmq.NewRabbitMqConn()
	_, err := this.Conn.MqConnect()
	if (err != nil) {
		fmt.Println("mqtask_pool_error", err)
		return nil,err
	}
	fmt.Println(mq_data)
	this.Conn.SetExchage(mq_data["exname"]).SetRouteKey(mq_data["routekey"]).SetQueueName(mq_data["queuename"]).SetupMQ()
	this.Slock.Lock()
	this.Data = make(map[string]interface{})
	this.Slock.Unlock()
	return this,nil
}



func GetNewGylib_WebRbMQ(qname string) (*WebRbMQ) {
	this := new(WebRbMQ)
	this.Conn = rbmq.NewRabbitMqConn()
	_, err := this.Conn.MqConnect()
	if (err != nil) {
		fmt.Println("mqtask_error", err)
		return nil
	}
	//fmt.Println(mq_data)
	this.Conn.SetExchage(mq_data["exname"]).SetRouteKey(mq_data["routekey"]).SetQueueName(qname).SetupMQ()
	this.Slock.Lock()
	this.Data = make(map[string]interface{})
	this.Slock.Unlock()
	return this

}

func (this *WebRbMQ) SetData(data map[string]interface{}) (*WebRbMQ) {
	this.Slock.Lock()
	this.Data = data
	this.Slock.Unlock()
	return this
}

func (this *WebRbMQ)  Publish() (bool) {
	flag:=this.Conn.Publish(string(datatype.Map2Json(this.Data)))
	MqPool.Put(this)
	return flag
}

func (this *WebRbMQ) Close() (error) {
	return this.Conn.Conn.Close()
}
