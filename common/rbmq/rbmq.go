package rbmq

import (
	"github.com/streadway/amqp"
	"gylib/common"

	"fmt"
	"bytes"
)

type Rbmq struct {
	Conn      *amqp.Connection
	Channel   *amqp.Channel
	Hasmq     bool
	Exchage   string
	Key       string
	QueueName string
}

func New_RabitMqConn() (*Rbmq) {
	this := new(Rbmq)
	this.Hasmq = false
	return this
}

func (this *Rbmq) MqConnect() (bool, error) {
	// amqp://用户名:密码@地址:端口号/host
	data := common.Getini("conf/app.ini", "amqp", map[string]string{"user": "root", "password": "",
		"host_ip": "127.0.0.1", "port": "5672", "host_name": "/"})
	var err error
	this.Conn, err = amqp.Dial(fmt.Sprintf("amqp://%v:%v@%v:%v%v", data["user"], data["password"], data["host_ip"], data["port"], data["host_name"]))
	if (err != nil) {
		this.Hasmq = false
		return false, err
	} else {
		this.Channel, err = this.Conn.Channel()
		this.Hasmq = true
		return true, nil
	}
}

func (this *Rbmq) Publish(msg string) (bool) {
	if (!this.Hasmq) {
		this.MqConnect()
	}
	ch, _ := this.Conn.Channel()
	defer ch.Close()
	err := ch.Publish(
		this.Exchage,
		this.Key, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
	if (err != nil) {
		fmt.Println(err)
		return false
	}
	return true
}

func (this *Rbmq) Receive() ([]string) {
	if this.Channel == nil {
		this.MqConnect()
	}
	msgs, err := this.Channel.Consume(this.QueueName, "", true, false, false, false, nil)
	if (err != nil) {
		return nil
	}
	result := make([]string, 0)
	forever := make(chan bool)
	go func() {
		//fmt.Println(*msgs)
		for d := range msgs {
			s := BytesToString(&(d.Body))
			result = append(result, *s)
		}
	}()
	<-forever
	defer close(forever)
	return result
}

func BytesToString(b *[]byte) *string {
	s := bytes.NewBuffer(*b)
	r := s.String()
	return &r
}
