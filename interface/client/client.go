package client

import (
	"../events"
	"github.com/gorilla/websocket"
	"log"
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
		client.output <- "pong: " + self.Data
	case "subscribe":
		if client.events.Find(self.Data) {
			client.output <- "Client already subscribed to events '" + self.Data + "'"
			return
		}
		if err := events.Subscribe(self.Data, client.output, client.ws.RemoteAddr().String()); err != nil {
			log.Println(err.Error())
		} else {
			client.events = append(client.events, self.Data)
		}
	case "unsubscribe":
		if err := events.Unsubscribe(self.Data, client.output, client.ws.RemoteAddr().String()); err != nil {
			log.Println(err.Error())
		} else {
			client.events = client.events.Remove(self.Data)
		}
	default:
		client.output <- "Command wasn't found"
	}
}
