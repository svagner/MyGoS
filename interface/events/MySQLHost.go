package events

import (
	"../convert"
	"../databases"
	"encoding/json"
)

func MySQLHost(data string, co chan string, ip string) error {
	var userData databases.Db
	if err := json.Unmarshal([]byte(data), &userData); err != nil {
		Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
		co <- convert.ConvertToJSON_HTML(Data)
		return err
	}
	db, err := databases.AddMySQLHost(userData)
	if err != nil {
		Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
		co <- convert.ConvertToJSON_HTML(Data)
		return err
	}

	Data := ResCmd{Channel: "MySQLHost", Command: "add", Data: db}
	res, err := json.Marshal(Data)
	if err != nil {
		Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
		co <- convert.ConvertToJSON_HTML(Data)
		return err
	}
	Events["MySQLHost"].channel <- string(res)
	return nil
}

func MySQLHostDelete(data string, co chan string, ip string) error {
	if err := databases.DeleteMySQLHost(data); err != nil {
		Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
		co <- convert.ConvertToJSON_HTML(Data)
		return err
	}
	res := ResCmd{Channel: "MySQLHost", Command: "delete", Data: data}
	Events["MySQLHost"].channel <- convert.ConvertToJSON_HTML(res)
	return nil
}
