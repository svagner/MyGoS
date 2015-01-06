package events

import (
	"errors"
)

type clientChan struct {
	c  chan string
	ip string
}

type chanList []clientChan

func (self chanList) Remove(clchan clientChan) (res chanList) {
	for _, cn := range self {
		if cn != clchan {
			res = append(res, cn)
		}
	}
	return res
}

type Event struct {
	genEvent    func(string, string)
	channel     chan string
	subscribers chanList
}

var Events = make(map[string]*Event)

func (self *Event) Notifier() {
	for {
		select {
		case data := <-self.channel:
			for _, cc := range self.subscribers {
				cc.c <- data
			}
		}
	}
}

func (self *Event) AddUser(out chan string, ip string) {
	self.subscribers = append(self.subscribers, clientChan{c: out, ip: ip})
}

func Init() {
	Events["connectlist"] = &Event{ConnectionListSubscribe, make(chan string), make(chanList, 0)}
	go Events["connectlist"].Notifier()
	Events["replicationGroups"] = &Event{ConnectionListSubscribe, make(chan string), make(chanList, 0)}
	go Events["replicationGroups"].Notifier()
	// Events about updata/add/delete MySQL hosts
	Events["MySQLHost"] = &Event{ConnectionListSubscribe, make(chan string), make(chanList, 0)}
	go Events["MySQLHost"].Notifier()
	// Events about mysql data statistics etc.
	Events["MySQLData"] = &Event{MySQLDataSubscribe, make(chan string), make(chanList, 0)}
	go Events["MySQLData"].Notifier()
	Events["replicationSteps"] = &Event{replicationStepsSubscribe, make(chan string), make(chanList, 0)}
	go Events["replicationSteps"].Notifier()
}

func Unsubscribe(event string, out chan string, ip string) error {
	Events[event].subscribers = Events[event].subscribers.Remove(clientChan{c: out, ip: ip})
	return nil
}

func Subscribe(event string, out chan string, ip string) error {
	if _, ok := Events[event]; !ok {
		return errors.New("Channel wasn't found")
	}
	Events[event].AddUser(out, ip)
	if Events[event].genEvent != nil {
		Events[event].genEvent(event, ip)
	}
	return nil
}
