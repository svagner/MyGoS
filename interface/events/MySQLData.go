package events

import (
	"encoding/json"
	"time"

	"github.com/svagner/MyGoS/interface/convert"
	"github.com/svagner/MyGoS/tools/mysql"
)

const (
	SLEEP_TIME = 2
)

func MySQLDataSubscribe(event string, ip string) {
	if len(Events[event].subscribers) == 1 {
		go MySQLDataLoop(&Events[event].subscribers)
	}
}

func MySQLDataLoop(clients *chanList) {
	tick := time.NewTicker(SLEEP_TIME * time.Second)
	for {
		<-tick.C
		if len(*clients) == 0 {
			return
		} else {
			data, err := mysql.GetMySQLInfo()
			if err != nil {
				Data := ResCmd{Channel: "Error", Command: "new", Data: "MySQL Error:" + err.Error()}
				for _, co := range *clients {
					co.c <- convert.ConvertToJSON_HTML(Data)
				}
			} else {
				Data := ResCmd{Channel: "MySQLData", Command: "update", Data: data}
				res, err := json.Marshal(Data)
				if err != nil {
					Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
					for _, co := range *clients {
						co.c <- convert.ConvertToJSON_HTML(Data)
					}
				}
				Events["MySQLData"].channel <- string(res)
			}
		}
	}
}
