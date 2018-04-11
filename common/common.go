package common

import (
	"github.com/c4pt0r/ini"
	"strings"
	"net/smtp"
	"math/rand"
	"io"
	"fmt"
	"time"
	"sort"
	"crypto/md5"
	"net/http"
	"os"
	"bytes"
	"os/exec"
	"crypto/sha1"
	"reflect"
	"strconv"
	"github.com/satori/go.uuid"
	"encoding/base64"
	"io/ioutil"
	"path"
	"errors"
	"regexp"
	"bufio"
	"runtime"
)

func Getini(config_file, action string, post_data map[string]string) (map[string]string) {
	Conf := ini.NewConf(config_file)
	data := make(map[string]string)
	tmp_str := make(map[string]*string)
	for key, val := range post_data {
		tmp_str[key] = Conf.String(action, key, val)

	}
	Conf.Parse()
	for key, val := range tmp_str {
		data[key] = *val
	}
	return (data)
}


func Send_mail_public(title, mailto, mailbody string) (int) {
	data := Getini("conf/app.ini", "smtp", map[string]string{"smtp_user": "", "smtp_pass": "", "smtp_host": ""})
	to := mailto
	subject := title
	body := mailbody
	err := SendToMail(data["smtp_user"], data["smtp_pass"], data["smtp_host"], to, subject, body, "text")
	if err != nil {
		fmt.Println("Send mail error!")
		fmt.Println(err)
		return (0)
	} else {
		fmt.Println("Send mail success!")
		return (1)
	}
}

func SendToMail(user, password, host, to, subject, body, mailtype string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var content_type string
	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + user + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, send_to, msg)
	return err
}

func Signature_MD5(appid, appkey string, params map[string]string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sign := appid
	sort.Strings(keys)
	h := md5.New()
	//h := sha1.New()
	for _, k := range keys {
		sign += params[k]
	}
	sign += appkey
	//fmt.Println(sign)
	io.WriteString(h, sign)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Signature_sha1(appid, appkey string, params map[string]string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sign := appid
	sort.Strings(keys)
	h := sha1.New()
	//h := sha1.New()
	for _, k := range keys {
		sign += params[k]
	}
	sign += appkey
	io.WriteString(h, sign)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func RandomStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func Down_http_file(data map[string]string) {
	res, err := http.Get(data["down_url"])
	if err != nil {
		return
	}
	defer res.Body.Close()
	f, err := os.Create(data["save_path"])
	if err != nil {
		return
	}
	defer f.Close()
	io.Copy(f, res.Body)
}

func Shell_linux_exec(cmdstr string) {
	if (cmdstr == "") {
		return
	}
	in := bytes.NewBuffer(nil)
	cmd := exec.Command("/bin/bash")
	cmd.Stdin = in
	in.WriteString(cmdstr)
	err := cmd.Run()
	if (err != nil) {

	}
}

func Shell_win_exec(cmdstr string) {
	if (cmdstr == "") {
		return
	}
	in := bytes.NewBuffer(nil)
	cmd := exec.Command("cmd")
	cmd.Stdin = in
	var out bytes.Buffer
	cmd.Stdout = &out
	in.WriteString(cmdstr)
	err := cmd.Run()
	if (err != nil) {

	} else {
		//fmt.Println(out.String())
	}
}

//结构转map
func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		fd_name := Tolow_map_name(t.Field(i).Name)
		data[fd_name] = v.Field(i).Interface()
	}
	return data
}

//驼峰写法转下划线写法
func Tolow_map_name(name string) (string) {
	result := ""
	for k, v := range name {
		if (v >= 'A' && v <= 'Z') {
			if (k > 0) {
				result += "_" + strings.ToLower(string(v))
			} else {
				result += strings.ToLower(string(v))
			}
		} else {
			result += strings.ToLower(string(v))
		}

	}
	return result
}

//map转结构体
func DataToStruct(data map[string]string, out interface{}) {
	ss := reflect.ValueOf(out).Elem()
	for i := 0; i < ss.NumField(); i++ {
		name := ss.Type().Field(i).Name
		val, ok := data[Tolow_map_name(name)]
		if (ok == false) {
			continue
		}
		switch ss.Field(i).Kind() {
		case reflect.String:
			ss.FieldByName(name).SetString(val)
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.Atoi(val)
			if err != nil {
				continue
			}
			ss.FieldByName(name).SetInt(int64(i))
		case reflect.Uint16, reflect.Uint32, reflect.Uint64:
			i, err := strconv.Atoi(val)
			if err != nil {
				continue
			}
			ss.FieldByName(name).SetUint(uint64(i))
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				continue
			}
			ss.FieldByName(name).SetFloat(f)
		case reflect.Struct:
			fmt.Println(ss.Field(i), ss.Field(i).NumField())
			//f,err:=time.Parse("2006-01-02 15:04:05", val)
			//ss.FieldByName(name).Set(f)
		default:
			fmt.Println("unknown type:%+v", ss.Field(i).Kind())
		}
	}
	return
}

func Has_map_index(name string, data map[string]interface{}) bool {
	_, ok := data[name]
	return ok

}

func Type2str(val interface{}) (string) {
	//fmt.Println(fmt.Sprintf("%T,%v",val,val))
	var result string = ""
	switch val.(type) {
	case []string:
		strArray := val.([]string)
		result = strings.Join(strArray, "")
	case []uint8:
		result = string(val.([]uint8))
	default:
		result = fmt.Sprintf("%v", val)
	}
	return result
}

func Byte2str(postdata []map[string][]byte) []map[string]interface{} {
	data := make([]map[string]interface{}, 0)
	for _, val := range postdata {
		temp := make(map[string]interface{})
		for k, v := range val {
			temp[strings.ToLower(k)] = string(v[:])
		}
		data = append(data, temp)
	}
	return data
}

func Map2str(postdata map[string]interface{}) (map[string]string) {
	data := make(map[string]string)
	for key, val := range postdata {
		data[key] = Type2str(val)
	}
	return data
}

func Get_UUID() (string) {
	uuid, _ := uuid.NewV4()
	return uuid.String()

}

//下划线转驼峰写法转写法
func ToUP_map_name(name string) (string) {
	result := ""
	flag := false
	for _, v := range name {
		if (v == '_') {
			flag = true
		} else {
			if (flag) {
				result += strings.ToUpper(string(v))
				flag = false
			} else {
				result += strings.ToLower(string(v))
			}
		}

	}
	return strings.Title(result)
}

func String2Time(date string) (int64) {
	toBeCharge := date
	timeLayout := "2006-01-02 15:04:05"                             //转化所需模板
	loc, _ := time.LoadLocation("Local")                            //重要：获取时区
	theTime, _ := time.ParseInLocation(timeLayout, toBeCharge, loc) //使用模板在对应时区转化为time.time类型
	sr := theTime.Unix()                                            //转化为时间戳 类型是int64
	return sr
}

func Int2Time_str(date int64) (string) {
	//格式化为字符串,tm为Time类型
	tm := time.Unix(date, 0)
	return tm.Format("2006-01-02 15:04:05")

}

func Int2Date_str(date int64) (string) {
	//格式化为字符串,tm为Time类型
	tm := time.Unix(date, 0)
	return tm.Format("2006-01-02")

}

func MapString2interface(data map[string]string)map[string]interface{}{
	result:=make(map[string]interface{})
	for key,val:=range data{
		result[key]=val
    }
    return result
}

/*Base64图片编码写入文件*/
func Base64ToImg(datasource string, dirpath string) (string) {
	if (dirpath == "") {
		dirpath = "./static/uploads/"
	}
	dirpath += Int2Date_str(time.Now().Unix()) + "/"
	if (!PathExists(dirpath)) {
		os.Mkdir(dirpath, os.ModePerm)
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	dirpath += strconv.Itoa(r.Intn(99999))
	dirpath += strconv.FormatInt(time.Now().Unix(), 10) + ".jpg"
	f, _ := base64.StdEncoding.DecodeString(datasource) //成图片文件并把文件写入到buffer
	err2 := ioutil.WriteFile(dirpath, f, os.ModePerm)   //buffer输出到jpg文件中（不做处理，直接写到文件）
	if (err2 != nil) {
		return ""
	}
	return dirpath[1:]
}

func PathExists(dirpath string) (bool) {
	_, err := os.Stat(dirpath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func Get_Upload_filename(filestr, dirstr string) (string) {
	dirpath := dirstr
	if (dirpath == "") {
		dirpath = "./static/uploads/"
	}
	dirpath += Int2Date_str(time.Now().Unix()) + "/"
	if (!PathExists(dirpath)) {
		os.Mkdir(dirpath, os.ModePerm)
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	dirpath += strconv.Itoa(r.Intn(99999))
	extname := path.Ext(filestr)
	if (extname == "") {
		extname = ".jpg"
	}
	dirpath += strconv.FormatInt(time.Now().Unix(), 10) + extname
	return dirpath
}

func Upload_File(r *http.Request, uploadfile string) ([]string) {
	fhs := r.MultipartForm.File[uploadfile]
	result := make([]string, len(fhs))
	for i := 0; i < len(fhs); i++ {
		file, err := fhs[i].Open()
		if err != nil {
			continue
		}
		filename := Get_Upload_filename(fhs[i].Filename, "")
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		defer f.Close()
		io.Copy(f, file)
		result = append(result, filename[1:])
	}
	return result
}


func Int64toint(val int64) int {
	result := strconv.FormatInt(val,10)
	data, err1 := strconv.Atoi(result)
	if (err1 != nil) {
		return -1
	}
	return data
}

func Eval(code string, imports ...string) (result string, err error) {
	var (
		dirSeparator string = "/"
		tempDir      string
	)

	if runtime.GOOS == "windows" {
		dirSeparator = "\\"
	}
	tempDir = os.TempDir() + dirSeparator + "goeval"
	os.Mkdir(tempDir, os.ModePerm)

	tmpfile, err := os.Create(tempDir + dirSeparator + "temp.go")
	if err != nil {
		return "", err
	}
	w := bufio.NewWriter(tmpfile)
	w.WriteString("package main\r\n")
	w.WriteString("\r\n")
	if len(imports) > 0 {
		tmpArgs := []string{"get"}
		tmpArgs = append(tmpArgs, imports...)
		goget := exec.Command("go", tmpArgs...)
		_, err = goget.Output()
		if err != nil {
			return "", err
		}
		w.WriteString("import (\r\n")
		for _, v := range imports {
			w.WriteString("\t" + `"` + v + `"` + "\r\n")
		}
		w.WriteString(")\r\n")
		w.WriteString("\r\n")
	}
	w.WriteString("func main() {\r\n")
	w.WriteString("\t" + code + "\r\n")
	w.WriteString("}\r\n")
	w.Flush()
	tmpfile.Close()
	cmd := exec.Command("go", "run", tmpfile.Name())
	res, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(res) + err.Error())
	}
	os.Remove(tmpfile.Name())
	return string(res), nil
}

func ConvertNumToCny(num float64) string {
	strnum := strconv.FormatFloat(num*100, 'f', 0, 64)
	sliceUnit := []string{"仟", "佰", "拾", "亿", "仟", "佰", "拾", "万", "仟", "佰", "拾", "元", "角", "分"}
	// log.Println(sliceUnit[:len(sliceUnit)-2])
	s := sliceUnit[len(sliceUnit)-len(strnum) : len(sliceUnit)]
	upperDigitUnit := map[string]string{"0": "零", "1": "壹", "2": "贰", "3": "叁", "4": "肆", "5": "伍", "6": "陆", "7": "柒", "8": "捌", "9": "玖"}
	str := ""
	for k, v := range strnum[:] {
		str = str + upperDigitUnit[string(v)] + s[k]
	}
	reg, err := regexp.Compile(`零角零分$`)
	str = reg.ReplaceAllString(str, "整")

	reg, err = regexp.Compile(`零角`)
	str = reg.ReplaceAllString(str, "零")

	reg, err = regexp.Compile(`零分$`)
	str = reg.ReplaceAllString(str, "整")

	reg, err = regexp.Compile(`零[仟佰拾]`)
	str = reg.ReplaceAllString(str, "零")

	reg, err = regexp.Compile(`零{2,}`)
	str = reg.ReplaceAllString(str, "零")

	reg, err = regexp.Compile(`零亿`)
	str = reg.ReplaceAllString(str, "亿")

	reg, err = regexp.Compile(`零万`)
	str = reg.ReplaceAllString(str, "万")

	reg, err = regexp.Compile(`零*元`)
	str = reg.ReplaceAllString(str, "元")

	reg, err = regexp.Compile(`亿零{0, 3}万`)
	str = reg.ReplaceAllString(str, "^元")

	reg, err = regexp.Compile(`零元`)
	str = reg.ReplaceAllString(str, "零")
	if err != nil {
		fmt.Print(err)
	}
	return str
}