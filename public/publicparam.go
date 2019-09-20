package public

import "gylib/common"

var App_name string

func init()  {
	app_data := common.Getini("conf/app.ini", "server", map[string]string{"appname": ""})
	App_name = app_data["appname"]
}

