package webclient

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"bytes"
	"net/url"
	"gylib/common/datatype"
	"crypto/tls"
)

func Web_Form_POST(url_add string,data url.Values)string{
	//s_data:=url.Values{}
	//for k,v:=range data{
	//	s_data.Set(k,datatype.Type2str(v))
	//}
	res, err := http.PostForm(url_add, data)
	//设置http中header参数，可以再此添加cookie等值
	//res.Header.Add("User-Agent", "***")
	//res.Header.Add("http.socket.timeou", 5000)

	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	return string(body)
}


func HttpGet(url_add string) string{
	resp, err := http.Get(url_add)
	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	return string(body)
}

func Http_post_json(url_add string,data string)string{
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println("json err:", err)
	}

	body := bytes.NewBuffer([]byte(b))
	res,err := http.Post(url_add, "application/json;charset=utf-8", body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	result, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {

		return ""
	}
	return string(result)
}

func Https_post(url_add string,data map[string]interface{})(string){
	s_data:=url.Values{}
	for k,v:=range data{
		s_data.Set(k,datatype.Type2str(v))
	}
	var resp *http.Response
	var err error
	var result []byte
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	//启用cookie
	//client.Jar, _ = cookiejar.New(nil)
	resp, err = client.PostForm(url_add, s_data)
	if result, err = ioutil.ReadAll(resp.Body); err == nil {
		fmt.Printf("%s\n", data)
		return ""
	}
	defer resp.Body.Close()
    return string(result)

}


func DoBytesPost(url string, data []byte) (string, error) {

	body := bytes.NewReader(data)
	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		//log.Println("http.NewRequest,[err=%s][url=%s]", err, url)
		return "", err
	}
	request.Header.Set("Connection", "Keep-Alive")
	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		//log.Println("http.Do failed,[err=%s][url=%s]", err, url)
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Println("http.Do failed,[err=%s][url=%s]", err, url)
		return "", err
	}
	return string(b), err
}