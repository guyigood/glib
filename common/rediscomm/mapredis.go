package rediscomm

import (
	"gylib/common/redispack"
	"encoding/json"
	"gylib/common/datatype"
)

type RedisComm struct {
	Key         string
	Field       string
	Common_exec string
	Timeout     int
	Re_prefix   string
	Data        interface{}
}

func NewRedisComm() (*RedisComm) {
	this := new(RedisComm)
	this.Re_prefix = redispack.Redis_data["redis_perfix"]
	this.Timeout = 3600
	this.Common_exec = "SET"
	return this
}

func (this *RedisComm) Flushdb() {
	client := redispack.Get_redis_pool().Get()
	//defer client.Close()
	client.Do("FLUSHDB")
}

func (this *RedisComm) HasKey() (bool) {
	client := redispack.Get_redis_pool().Get()
	//defer client.Close()
	hasok, err := client.Do("EXISTS", this.Key)
	if (err != nil) {
		return false
	}
	if (datatype.Type2int(hasok) == 0) {
		return false
	} else {
		return true
	}

}

func (this *RedisComm) DelKey() (bool) {
	client := redispack.Get_redis_pool().Get()
	//defer client.Close()
	_, err := client.Do("DEL", this.Key)
	if (err != nil) {
		return false
	}

	return true
}

func (this *RedisComm) GetRawValue() (interface{}) {
	client := redispack.Get_redis_pool().Get()
	//defer client.Close()
	raw, err := client.Do("GET", this.Key)
	if (err != nil) {
		return nil
	}
	if (raw == nil) {
		return nil
	}
	return raw
}

func (this *RedisComm) SetRawValue() {
	client := redispack.Get_redis_pool().Get()
	if (this.Common_exec == "SETEX") {
		client.Do("SETEX", this.Key, this.Timeout, this.Data)
	} else {
		client.Do("SET", this.Key, this.Data)
	}

}

func(this *RedisComm)Getkey()(string){
	return this.Key
}

func (this *RedisComm) SetKey(key string) (*RedisComm) {
	if (this.Re_prefix != "") {
		strlen := len(this.Re_prefix)
		if(len(key)<strlen){
			this.Key = this.Re_prefix + key
			return this
		}
		if (key[:strlen] == this.Re_prefix) {
			this.Key = key
		} else {
			this.Key = this.Re_prefix + key
		}
	} else {
		this.Key = key
	}
	//fmt.Println(this.Key,key)
	return this
}

func (this *RedisComm) SetFiled(key string) (*RedisComm) {
	this.Field = key
	return this
}

func (this *RedisComm) SetExec(key string) (*RedisComm) {
	this.Common_exec = key
	return this
}

func (this *RedisComm) SetTime(timect int) (*RedisComm) {
	this.Timeout = timect
	return this
}

func (this *RedisComm) SetData(data interface{}) (*RedisComm) {
	this.Data = data
	return this
}

func (this *RedisComm) Get_value() (interface{}) {
	client := redispack.Get_redis_pool().Get()
	//defer client.Close()
	raw, err := client.Do("GET", this.Key)
	if (err != nil) {
		return nil
	}
	if (raw == nil) {
		return nil
	}
	//var data interface{}
	err = json.Unmarshal(raw.([]byte), &this.Data)
	if (err != nil) {
		return nil
	}
	return this.Data
}

func (this *RedisComm) Set_value() {
	client := redispack.Get_redis_pool().Get()
	//defer client.Close()
	raw, _ := json.Marshal(&this.Data)
	if (this.Common_exec == "SETEX") {
		client.Do("SETEX", this.Key, this.Timeout, raw)
	} else {
		client.Do("SET", this.Key, raw)
	}
}

func (this *RedisComm) SetList() {
	client := redispack.Get_redis_pool().Get()
	//defer client.Close()
	raw, _ := json.Marshal(&this.Data)
	client.Do("LPUSH", this.Key, raw)
}

func (this *RedisComm) GetList() (interface{}) {
	client := redispack.Get_redis_pool().Get()
	//defer client.Close()
	raw, err := client.Do("RPOP", this.Key)
	if (err != nil) {
		return nil
	}
	if (raw == nil) {
		return nil
	}
	//var data interface{}
	err = json.Unmarshal(raw.([]byte), &this.Data)
	if (err != nil) {
		return nil
	}
	return this.Data
}

func (this *RedisComm) Hset_map() {
	client := redispack.Get_redis_pool().Get()
	//defer client.Close()
	raw, _ := json.Marshal(&this.Data)
	client.Do("HSET", this.Key, this.Field, raw)
}

func (this *RedisComm) Hget_map() (interface{}) {
	client := redispack.Get_redis_pool().Get()
	//defer client.Close()
	raw, err := client.Do("HGET", this.Key, this.Field)
	if (err != nil) {
		return nil
	}
	if (raw == nil) {
		return nil
	}
	//var data interface{}
	//fmt.Println("raw=",string(raw.([]byte)))
	err = json.Unmarshal(raw.([]byte), &this.Data)
	if (err != nil) {
		//fmt.Println(this.Key,this.Field,err)
		return nil
	}
	return this.Data
}

func (this *RedisComm) Push(channel_name, message string) (int) { //发布者
	client := redispack.Get_redis_pool().Get()
	//defer client.Close()
	raw, _ := client.Do("PUBLISH", channel_name, message)
	return datatype.Type2int(raw)
}
