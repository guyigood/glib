package weixinsdk

import (
	"strings"
	"net/http"
	"gylib/common/redispack"
	"fmt"
	"io/ioutil"
	"encoding/json"

)

type wxsdk struct {
	Access_token string
}

func get_access_token(appid,appkey string) (string) {
	redis_pool := redispack.Get_redis_pool()
	redis := redis_pool.Get()
	access_token, err := redis.Do("GET", "access_token")
	if (err != nil) {
		url := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" +appid + "&secret=" + appkey
		res, err := http.Get(url)
		if (err != nil) {
			return ""
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			// handle error
			return ""
		}
		http_result := make(map[string]interface{})
		json.Unmarshal([]byte(body), http_result)
		v,ok:=http_result["access_token"].(string)
		if(ok) {
			access_token =v
		}else {
			return ""
		}
	}
	redis.Do("SETEX", "access_token", 7000, access_token)
	return access_token.(string)
}

func (this *wxsdk)send_wx_template(memo string) (int) {

	url := "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=" + this.Access_token
	//data := fmt.Sprintf("{\"touser\":\"%s\",\"msgtype\":\"text\",\"text\":{\"content\":\"%s\"}}", userid, memo)
	resp, err := http.Post(url, "application/json", strings.NewReader(memo))
	if err != nil {
		fmt.Println(err)
		return 0
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		return 0
	}
	aStr := make(map[string]interface{})
	json.Unmarshal([]byte(body), aStr)
	v,ok:=aStr["errcode"].(string)
	if(ok){
	  if(v=="0"){
	  	return 1
	  }
	}
		return 0


}

