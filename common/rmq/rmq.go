package rmq

import (
	"github.com/streadway/amqp"
	"errors"
	"bytes"
	"strings"
	"fmt"
	"time"
)

type Rabitmq struct {
	Conn *amqp.Connection
	Channel *amqp.Channel
	Topics string
	Nodes string
	HasMQ bool
}


type Reader interface {
	Read(msg *string) (err error)
}

func (this *Rabitmq) Struct_init(){
	err := this.SetupRMQ("amqp://admin:123456@b.bug.com:5672/fourth") // amqp://用户名:密码@地址:端口号/host
	if err != nil {
		fmt.Println("err01 : ", err.Error())
	}
	err = this.Ping()
	if err != nil {
		fmt.Println("err02 : ", err.Error())
	}
	fmt.Println("receive message")
	err = this.Receive("first", "second", func (msg *string) {
		fmt.Printf("receve msg is :%s\n", *msg)
	})
	if err != nil {
		fmt.Println("err04 : ", err.Error())
	}
	fmt.Println("1 - end")
	fmt.Println("send message")
	for i := 0; i < 10; i++ {
		err = this.Publish("first", "当前时间：" + time.Now().String())
		if err != nil {
			fmt.Println("err03 : ", err.Error())
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("2 - end")
	this.Close()
}


// 初始化 参数格式：amqp://用户名:密码@地址:端口号/host
func (this *Rabitmq) SetupRMQ(rmqAddr string) (err error) {
	if this.Channel == nil {
		this.Conn, err = amqp.Dial(rmqAddr)
		if err != nil {
			return err
		}

		this.Channel, err = this.Conn.Channel()
		if err != nil {
			return err
		}

		this.HasMQ = true
	}
	return nil
}

// 是否已经初始化
func (this *Rabitmq) ISHasMQ() bool {
	return this.HasMQ
}

// 测试连接是否正常
func (this *Rabitmq)Ping() (err error) {

	if !this.HasMQ || this.Channel == nil {
		return errors.New("RabbitMQ is not initialize")
	}

	err = this.Channel.ExchangeDeclare("ping.ping", "topic", false, true, false, true, nil)
	if err != nil {
		return err
	}

	msgContent := "ping.ping"

	err = this.Channel.Publish("ping.ping", "ping.ping", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msgContent),
	})

	if err != nil {
		return err
	}

	err = this.Channel.ExchangeDelete("ping.ping", false, false)

	return err
}

// 发布消息
func (this *Rabitmq)Publish(topic, msg string) (err error) {

	if this.Topics == "" || !strings.Contains(this.Topics, topic) {
		err = this.Channel.ExchangeDeclare(topic, "topic", true, false, false, true, nil)
		if err != nil {
			return err
		}
		this.Topics += "  " + topic + "  "
	}

	err = this.Channel.Publish(this.Topics, this.Topics, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
	})

	return nil
}

// 监听接收到的消息
func (this *Rabitmq)Receive(topic, node string, reader func (msg *string)) (err error) {
	if this.Topics == "" || !strings.Contains(this.Topics, topic) {
		err = this.Channel.ExchangeDeclare(topic, "topic", true, false,false, true, nil)
		if err != nil {
			return err
		}
		this.Topics += "  " + topic + "  "
	}
	if this.Nodes == "" || !strings.Contains(this.Nodes, node) {
		_, err = this.Channel.QueueDeclare(node, true, false,false, true, nil)
		if err != nil {
			return err
		}
		err = this.Channel.QueueBind(node, this.Topics, this.Topics, true, nil)
		if err != nil {
			return err
		}
		this.Nodes += "  " + node + "  "
	}

	msgs, err := this.Channel.Consume(this.Nodes, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		//fmt.Println(*msgs)
		for d := range msgs {
			s := bytesToString(&(d.Body))
			reader(s)
		}
	}()

	return nil
}

// 关闭连接
func (this *Rabitmq) Close() {
	this.Channel.Close()
	this.Conn.Close()
	this.HasMQ = false
}

func bytesToString(b *[]byte) *string {
	s := bytes.NewBuffer(*b)
	r := s.String()
	return &r
}
