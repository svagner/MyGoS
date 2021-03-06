package client

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/svagner/MyGoS/interface/convert"
	"github.com/svagner/MyGoS/interface/databases"
	"github.com/svagner/MyGoS/interface/events"
)

type eventsList []string

type Client struct {
	ip     string
	ua     string
	ws     *websocket.Conn
	output chan string
	events eventsList
}

type Command struct {
	Cmd  string
	Data string
}

func (self eventsList) Remove(data string) eventsList {
	var res eventsList
	for _, rec := range self {
		if rec != data {
			res = append(res, rec)
		}
	}
	return res
}

func (self eventsList) Find(data string) bool {
	for _, rec := range self {
		if rec == data {
			return true
		}
	}
	return false
}

func (self *Client) ReadCmd() {
	for {
		cmdData := &Command{}
		if err := self.ws.ReadJSON(cmdData); err != nil {
			log.Println(err.Error())
			break
		}
		go cmdData.Run(self)
	}
	self.Close()
}

func (self *Client) Close() {
	self.ws.Close()
	for _, event := range self.events {
		events.Unsubscribe(event, self.output, self.ws.RemoteAddr().String())
	}
}

func (self *Client) Receiver() {
	for {
		if err := self.ws.WriteJSON(<-self.output); err != nil {
			log.Println(err.Error())
			break
		}
	}
	self.Close()
}

func NewClient(ws *websocket.Conn, ip, ua string) {
	newClient := &Client{ip: ip, ua: ua, ws: ws, output: make(chan string)}
	go newClient.ReadCmd()
	go newClient.Receiver()
}

func (self *Command) Run(client *Client) {
	switch self.Cmd {
	case "test":
		data, _ := json.Marshal("pong: " + self.Data)
		client.output <- string(data)
	case "subscribe":
		if client.events.Find(self.Data) {
			data, _ := json.Marshal("Client already subscribed to events '" + self.Data + "'")
			Data := events.ResCmd{Channel: "Error", Command: "new", Data: "Subscribe to [" + self.Data + "] error: " + string(data)}
			client.output <- convert.ConvertToJSON_HTML(Data)
			return
		}
		if err := events.Subscribe(self.Data, client.output, client.ws.RemoteAddr().String()); err != nil {
			Data := events.ResCmd{Channel: "Error", Command: "new", Data: "Subscribe to [" + self.Data + "] error: " + err.Error()}
			client.output <- convert.ConvertToJSON_HTML(Data)
		} else {
			client.events = append(client.events, self.Data)
		}
	case "unsubscribe":
		if err := events.Unsubscribe(self.Data, client.output, client.ws.RemoteAddr().String()); err != nil {
			log.Println(err.Error())
		} else {
			client.events = client.events.Remove(self.Data)
		}
	case "replicationGroups":
		if err := events.ReplicationGroups(self.Data, client.output, client.ws.RemoteAddr().String()); err != nil {
			log.Println(err.Error())
		}
	case "replicationGroupsEdit":
		if err := events.ReplicationGroupsEdit(self.Data, client.output, client.ws.RemoteAddr().String()); err != nil {
			log.Println(err.Error())
		}
	case "replicationGroupsDelete":
		if err := events.ReplicationGroupsDelete(self.Data, client.output, client.ws.RemoteAddr().String()); err != nil {
			log.Println(err.Error())
		}
	case "getDatabasesData":
		res := struct {
			Command string
			Data    interface{}
		}{
			Command: self.Cmd,
			Data:    databases.GetDatabasesList(),
		}
		log.Println(res)
		client.output <- convert.ConvertToJSON_HTML(res)
	case "getHostData":
		if _, ok := databases.HostsList[self.Data]; !ok {
			Data := events.ResCmd{Channel: "Error", Command: "new", Data: "Host " + self.Data + " wasn't found"}
			client.output <- convert.ConvertToJSON_HTML(Data)
			return
		}
		res := struct {
			Command string
			Data    interface{}
		}{
			Command: self.Cmd,
			Data:    databases.HostsList[self.Data].GetDescription(),
		}
		client.output <- convert.ConvertToJSON_HTML(res)
	case "MySQLHost":
		if err := events.MySQLHost(self.Data, client.output, client.ws.RemoteAddr().String()); err != nil {
			log.Println(err.Error())
		}
	case "MySQLHostEdit":
		log.Println("Try to edit host")
	case "addReplicationStepScript":
		data := new(databases.MySQLReplicaScript)
		if err := json.Unmarshal([]byte(self.Data), &data); err != nil {
			Data := events.ResCmd{Channel: "Error", Command: "new", Data: "Command data [" + self.Data + "] isn't correct"}
			client.output <- convert.ConvertToJSON_HTML(Data)
		}
		log.Println(data)
		events.ReplicationStepAdd(*data)
	case "deleteReplicationStepScript":
		events.ReplicationStepDelete(self.Data)
	case "testReplicationStepScript":
		data, errRes := databases.RunReplicationStep(self.Data, true)
		type ScriptRes struct {
			Data  string
			Error string
		}
		var res []byte
		var err error
		if errRes != nil {
			res, err = json.Marshal(ScriptRes{Data: data, Error: errRes.Error()})
		} else {
			res, err = json.Marshal(ScriptRes{Data: data, Error: ""})
		}
		if err != nil {
			Data := events.ResCmd{Channel: "testReplicationStepScript", Command: "error", Data: "Convert to JSON was failed! Was run: " + self.Data + " .Ended with error: " + err.Error()}
			client.output <- convert.ConvertToJSON_HTML(Data)
		} else {
			if errRes != nil {
				Data := events.ResCmd{Channel: "testReplicationStepScript", Command: "error", Data: string(res)}
				client.output <- convert.ConvertToJSON_HTML(Data)
			} else {
				Data := events.ResCmd{Channel: "testReplicationStepScript", Command: "result", Data: string(res)}
				client.output <- convert.ConvertToJSON_HTML(Data)
			}
		}
	case "saveReplicationStepsSelected":
		var data []string
		if err := json.Unmarshal([]byte(self.Data), &data); err != nil {
			Data := events.ResCmd{Channel: "Error", Command: "new", Data: "Command data [" + self.Data + "] isn't correct"}
			client.output <- convert.ConvertToJSON_HTML(Data)
		}
		events.SaveReplicationStepsSelected(client.output, data)
	case "MySQLHostDelete":
		if err := events.MySQLHostDelete(self.Data, client.output, client.ws.RemoteAddr().String()); err != nil {
			log.Println(err.Error())
		}
	default:
		Data := events.ResCmd{Channel: "Error", Command: "new", Data: "Command [" + self.Cmd + "] wasn't found"}
		client.output <- convert.ConvertToJSON_HTML(Data)
	}
}
