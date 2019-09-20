package rbmq

import (
	"github.com/streadway/amqp"
	"gylib/common"
	"fmt"
)

type Rbmq struct {
	Conn      *amqp.Connection
	Channel   *amqp.Channel
	Consumer  string
	Hasmq     bool
	Exchage   string
	Key       string
	QueueName string
}

func NewRabbitMqConn() (*Rbmq) {
	this := new(Rbmq)
	this.Consumer = ""
	this.Hasmq = false
	return this
}

func (this *Rbmq) MqConnect() (bool, error) {
	// amqp://用户名:密码@地址:端口号/host
	data := common.Getini("conf/app.ini", "amqp", map[string]string{"user": "root", "password": "",
		"host_ip": "127.0.0.1", "port": "5672", "host_name": ""})
	var err error
	constr := fmt.Sprintf("amqp://%v:%v@%v:%v/%v", data["user"], data["password"], data["host_ip"], data["port"], data["host_name"])
	//fmt.Println(constr)
	this.Conn, err = amqp.Dial(constr)
	if (err != nil) {
		fmt.Println("conerror:", constr)
		this.Hasmq = false
		return false, err
	} else {
		this.Channel, err = this.Conn.Channel()
		this.Hasmq = true
		return true, nil
	}

}

func (this *Rbmq) SetupMQ() (*Rbmq) {
	err := this.Channel.ExchangeDeclarePassive(this.Exchage, "direct", true, false, false, false, nil)
	if err != nil {
		err := this.Channel.ExchangeDeclare(this.Exchage, "direct", true, false, false, false, nil)
		if (err != nil) {
			fmt.Println("setex", err, this.Exchage)
		}
	}
	_, err = this.Channel.QueueDeclarePassive(this.QueueName, true, false, false, false, nil)
	if (err != nil) {
		_, err = this.Channel.QueueDeclare(this.QueueName, true, false, false, false, nil)
		if (err != nil) {
			fmt.Println("queue", err, this.QueueName)
		}
	}
	err = this.Channel.QueueBind(this.QueueName, this.Key, this.Exchage, false, nil)
	if (err != nil) {
		fmt.Println("queuebind", err)
	}
	/*err=this.Channel.ExchangeBind(this.Exchage,this.Key,this.QueueName,false,nil)
	if(err!=nil){
		fmt.Println("bind",err)

	}*/
	return this

}

func (this *Rbmq) SetExchage(exname string) (*Rbmq) {
	this.Exchage = exname
	return this
}

func (this *Rbmq) SetConsumer(Conname string) (*Rbmq) {
	this.Consumer = Conname
	return this
}

func (this *Rbmq) SetRouteKey(Keyname string) (*Rbmq) {
	this.Key = Keyname
	return this
}

func (this *Rbmq) SetQueueName(Keyname string) (*Rbmq) {
	this.QueueName = Keyname
	return this
}

func (this *Rbmq) Publish(msg string) (bool) {
	if (!this.Hasmq) {
		this.MqConnect()
	}
	var err error
	//ch:=this.Channel
	//defer ch.Close()
	err = this.Channel.Publish(
		this.Exchage,
		this.Key, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
	if (err != nil) {
		fmt.Println("publish", err)
		return false
	}
	return true
}

func (this *Rbmq) GetRbmq() (amqp.Delivery) {
	msg, _, err1 := this.Channel.Get(this.QueueName, true)
	if (err1 == nil) {
		return msg
	} else {
		return amqp.Delivery{}
	}
}

func (this *Rbmq) Receive() (<-chan amqp.Delivery) {
	if this.Channel == nil {
		this.MqConnect()
	}
	err := this.Channel.Qos(1, 0, true)
	if err != nil {
		return nil
	}
	msgs, err := this.Channel.Consume(this.QueueName, this.Consumer, false, false, false, false, nil)
	if (err != nil) {
		fmt.Printf("获取消费通道异常:%s \n", err)
		return nil
	}
	return msgs
	/*result := make([]string, 0)
	/*for msg := range msgs {
		// 处理数据
	    s :=string(msg.Body)
	    if(s!="") {
			result = append(result, s)
		}
	}*/

	/*forever := make(chan bool)
	go func() {
		for d := range msgs {
			s := fmt.Sprintf("%s",d.Body)
			//fmt.Println(s)
			result = append(result, s)
		}
		forever <- true
	}()

	<-forever
	defer close(forever)*/
	//return result
}
