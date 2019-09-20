package apiserver

import (
	"gylib/common/webclient"
	"gylib/common/datatype"
	"gylib/common/rediscomm"
	"strings"
	"sync"
)

type AppCenter struct {
	Token     string
	User      string
	Pass      string
	Url       string
	Token_url string
	Login_url string
	Lock      sync.Mutex
	Client    *webclient.Http_Client
	Data      map[string]interface{}
}

func NewAppcenter(code, pass, url string) (*AppCenter) {
	this := new(AppCenter)
	this.Lock.Lock()
	this.Data = make(map[string]interface{})
	this.Lock.Unlock()
	this.Url = url
	this.Pass = pass
	this.Token_url = "/api/get_token"
	this.Login_url = "/api/login"
	this.User = code
	this.Client=webclient.NewHttpClient()
	this.Get_Token()
	return this
}

func NewNoLoginAppcenter(code, pass, url string) (*AppCenter) {
	this := new(AppCenter)
	this.Lock.Lock()
	this.Data = make(map[string]interface{})
	this.Lock.Unlock()
	this.Url = url
	this.Pass = pass
	this.Token_url = "/api/get_token"
	this.Login_url = "/api/login"
	this.User = code
	this.Client=webclient.NewHttpClient()
	//this.Get_Token()
	return this
}

func (this *AppCenter) Get_Access_token() (string) {
	redis := rediscomm.NewRedisComm()
	if (redis.SetKey("api_server_token").HasKey()) {
		data := redis.SetKey("api_server_token").Get_value()
		if (data != nil) {
			this.Token = datatype.Type2str(data.(map[string]interface{})["token"])
			//redis.SetKey("api_server_token").SetData(map[string]string{"token": this.Token}).SetExec("SETEX").SetTime(3000).Set_value()
			return this.Token
		}
	}
	return ""
}
func (this *AppCenter) Del_Access_token() {
	redis := rediscomm.NewRedisComm()
	redis.SetKey("api_server_token").DelKey()
}

func (this *AppCenter) Get_Login_Token() {
	redis := rediscomm.NewRedisComm()
	if (redis.SetKey("api_server_token").HasKey()) {
		data := redis.SetKey("api_server_token").Get_value()
		if (data != nil) {
			this.Token = datatype.Type2str(data.(map[string]interface{})["token"])
			//redis.SetKey("api_server_token").SetData(map[string]string{"token": this.Token}).SetExec("SETEX").SetTime(3000).Set_value()
			return
		}
	}
	s_data := make(map[string]interface{})
	s_data["code"] = this.User
	s_data["pass"] = this.Pass
	data := this.Get_client(this.Token_url)
	if (data != nil) {
		s_data["access_token"] = datatype.Type2str(data["data"])
	}

	result := this.Client.Https_post(this.Url+this.Login_url, s_data)
	list := datatype.String2Json(result)
	if (list != nil) {
		this.Token = datatype.Type2str(data["data"])
		redis.SetKey("api_server_token").SetData(map[string]string{"token": this.Token}).SetExec("SETEX").SetTime(3000).Set_value()
	}
	/*list := datatype.String2Json(result)
	if (list != nil) {
		this.Token = "apidb_" + datatype.Type2str(list["data"])
		redis.SetKey("api_server_token").SetData(map[string]string{"token": this.Token}).SetExec("SETEX").SetTime(3000).Set_value()
	} else {
		this.Token = ""
	}*/
}

func (this *AppCenter) Get_Token() {
	redis := rediscomm.NewRedisComm()
	if (redis.SetKey("api_server_token").HasKey()) {
		data := redis.SetKey("api_server_token").Get_value()
		if (data != nil) {
			this.Token = datatype.Type2str(data.(map[string]interface{})["token"])
			//redis.SetKey("api_server_token").SetData(map[string]string{"token": this.Token}).SetExec("SETEX").SetTime(3000).Set_value()
			return
		}
	}
	s_data := make(map[string]interface{})
	s_data["code"] = this.User
	s_data["pass"] = this.Pass
	data := this.Get_client(this.Token_url)
	if (data != nil) {
		s_data["access_token"] = datatype.Type2str(data["data"])
	}

	result := this.Client.Https_post(this.Url+this.Login_url, s_data)
	list := datatype.String2Json(result)
	if (list != nil) {
		this.Token = datatype.Type2str(data["data"])
		redis.SetKey("api_server_token").SetData(map[string]string{"token": this.Token}).SetExec("SETEX").SetTime(3000).Set_value()
	}
	/*list := datatype.String2Json(result)
	if (list != nil) {
		this.Token = "apidb_" + datatype.Type2str(list["data"])
		redis.SetKey("api_server_token").SetData(map[string]string{"token": this.Token}).SetExec("SETEX").SetTime(3000).Set_value()
	} else {
		this.Token = ""
	}*/
}

func (this *AppCenter) SetData(data map[string]interface{}) {
	this.Lock.Lock()
	this.Data = data
	this.Lock.Unlock()
}

func (this *AppCenter) Post_route(url string) (map[string]interface{}) {
	/*if (this.Token == "") {
		this.Get_Token()
	}*/
	this.Get_Login_Token()
	this.Lock.Lock()
	this.Data["access_token"] = this.Token
	this.Lock.Unlock()
	data := this.Client.Https_post(this.Url+url, this.Data)
	//fmt.Println("post",data)
	if (data == "") {
		return nil
	} else {
		return datatype.String2Json(data)
	}
}

func (this *AppCenter) Post_route_data(url string) ([]map[string]interface{}) {
	this.Get_Login_Token()
	this.Lock.Lock()
	this.Data["access_token"] = this.Token
	this.Lock.Unlock()
	data := this.Client.Https_post(this.Url+url, this.Data)
	//fmt.Println("post_route_data",this.Url+url,data,this.Data)
	if (data == "") {
		return nil
	} else {
		web_data := datatype.String2Json(data)
		result_data := make([]map[string]interface{}, 0)
		if (datatype.Type2str(web_data["status"]) != "100") {
			return nil
		}
		list, ok := web_data["data"].([]interface{})
		if (!ok) {
			return nil
		}
		for _, v := range list {
			tmp, ok := v.(map[string]interface{})
			if (ok) {
				result_data = append(result_data, tmp)
			}
			//fmt.Println(tmp)
		}
		return result_data
	}
}

func (this *AppCenter) Get_route(params string) (map[string]interface{}) {
	this.Get_Login_Token()

	if (strings.Contains(params, "?")) {
		params += "&access_token=" + this.Token
	} else {
		params = "?access_token=" + this.Token
	}
	//fmt.Println(this.Url+params)
	data := this.Client.HttpGet(this.Url + params)
	//fmt.Println(this.Url + params)
	if (data == "") {
		return nil
	} else {
		return datatype.String2Json(data)
	}
}

func (this *AppCenter) Get_client(url string) (map[string]interface{}) {
	data := this.Client.HttpGet(this.Url + url)
	if (data == "") {
		return nil
	} else {
		return datatype.String2Json(data)
	}
}
