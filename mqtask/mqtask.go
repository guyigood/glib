package mqtask

import (
	"gylib/common/rbmq"
	"fmt"
	"gylib/common/datatype"
	"sync"
	"gylib/common"
	"github.com/silenceper/pool"
	"time"
)


type WebRbMQ struct {
	Slock sync.Mutex
	Conn  *rbmq.Rbmq
	Data  map[string]interface{}
}
type WebMqPool struct {
	MqPool pool.Pool
	Mq_data map[string]string
	Mq_name string
	WebRb *WebRbMQ
}

func GetNewWebMqPool(mq_name string)(*WebMqPool) {
	this:=new(WebMqPool)
	if(mq_name==""){
		this.Mq_name="amqp"
	}else{
		this.Mq_name=mq_name
	}
	this.Mq_data=make(map[string]string)
	this.Mq_data = common.Getini("conf/app.ini", this.Mq_name, map[string]string{"queuename": "crmback", "exname": "", "routekey": "", "max_pool": "100", "min_pool": "5"})
	factory := func() (interface{}, error) { return this.PoolWebMq() }
	close := func(v interface{}) error { return v.(*WebRbMQ).Close() }
	min_pool := datatype.Str2Int(this.Mq_data["min_pool"])
	max_pool := datatype.Str2Int(this.Mq_data["max_pool"])
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
	this.MqPool, err = pool.NewChannelPool(poolConfig)
	if err != nil {
		fmt.Println("err=", err)
	}
	return this
}

func (this *WebMqPool) GetPoolWebMq()(*WebRbMQ,error){
	that, err := this.MqPool.Get()
	if (err != nil) {
		fmt.Println(err)
		return nil,err
	}
	wbmq := that.(*WebRbMQ)
	return wbmq,nil
}



func  (this *WebMqPool) PoolWebMq() (*WebRbMQ, error) {
	//mq_data = common.Getini("conf/app.ini", "amqp", map[string]string{"queuename": "crmback", "exname": "", "routekey": "", "max_pool": "100", "min_pool": "5"})
	that := new(WebRbMQ)
	that.Conn = rbmq.NewRabbitMqConn()
	_, err := that.Conn.MqConnect()
	if (err != nil) {
		fmt.Println("mqtask_pool_error", err)
		return nil,err
	}
	that.Conn.SetExchage(this.Mq_data["exname"]).SetRouteKey(this.Mq_data["routekey"]).SetQueueName(this.Mq_data["queuename"]).SetupMQ()
	that.Slock.Lock()
	that.Data = make(map[string]interface{})
	that.Slock.Unlock()
	return that,nil
}



/*func GetNewGylib_WebRbMQ(qname string) (*WebRbMQ) {
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

}*/

func (this *WebRbMQ) SetData(data map[string]interface{}) (*WebRbMQ) {
	this.Slock.Lock()
	this.Data = data
	this.Slock.Unlock()
	return this
}

func (this *WebRbMQ) Publish(mqPool *WebMqPool) (bool) {
	flag:=this.Conn.Publish(string(datatype.Map2Json(this.Data)))
	mqPool.MqPool.Put(this)
	return flag
}

func (this *WebRbMQ) Close() (error) {
	return this.Conn.Conn.Close()
}



