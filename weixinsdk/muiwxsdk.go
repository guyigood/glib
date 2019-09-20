package weixinsdk

import (
	"strings"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"gylib/common/datatype"
	"gylib/common/webclient"
)

type MuiWxsdk struct {
	Access_token  string
	Appid, Appkey string
	Body          string
	Client        *webclient.Http_Client
}

func NewMuiWxsdk() (*MuiWxsdk) {
	this := new(MuiWxsdk)
	this.Client = webclient.NewHttpClient()
	return this
}

func (this *MuiWxsdk) MuiGet_access_token() (string) {
	url := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + this.Appid + "&secret=" + this.Appkey
	access_token := ""

	res, err := this.Client.Client.Get(url)
	if (err != nil) {
		fmt.Println(url, res.Body)
		return ""
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(string(body), err) // handle error
		return ""
	}
	http_result := make(map[string]interface{})
	//fmt.Println("get_access_token",string(body))
	json.Unmarshal([]byte(body), &http_result)
	v, ok := http_result["access_token"].(string)
	if (ok) {
		access_token = v
	} else {
		return ""
	}
	return access_token
}

func (this *MuiWxsdk) MuiSend_wx_template(memo string) (int) {

	url := "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=" + this.Access_token
	//data := fmt.Sprintf("{\"touser\":\"%s\",\"msgtype\":\"text\",\"text\":{\"content\":\"%s\"}}", userid, memo)
	resp, err := this.Client.Client.Post(url, "application/json", strings.NewReader(memo))
	if err != nil {
		fmt.Println(err, url, resp.Body)
		return 0
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		return 0
	}
	this.Body = string(body)
	aStr := make(map[string]interface{})
	json.Unmarshal([]byte(body), &aStr)
	v, ok := aStr["errcode"]
	if (ok) {
		if (datatype.Type2str(v) == "0") {
			//fmt.Println("body-ok", string(body), aStr)
			return 1
		} else {
			fmt.Println(url, "body", string(body))
		}
	}
	return 0

}

func (this *MuiWxsdk) MuiGet_Jsapi_ticket() (string) {
	access_token := ""

	url := "https://api.weixin.qq.com/cgi-bin/ticket/getticket?type=jsapi&access_token=" + this.Access_token
	res, err := this.Client.Client.Get(url)
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
	json.Unmarshal([]byte(body), &http_result)
	v, ok := http_result["ticket"]
	if (ok) {
		access_token = datatype.Type2str(v)
	} else {
		return ""
	}
	return datatype.Type2str(access_token)
}
