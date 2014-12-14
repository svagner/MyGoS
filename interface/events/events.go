package events

import (
	"errors"
	"log"
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
	handle      func()
	subscribers chanList
}

var Events = make(map[string]*Event)

func Init() {
	Events["connectlist"] = &Event{}
}

func Unsubscribe(event string, out chan string, ip string) error {
	Events[event].subscribers = Events[event].subscribers.Remove(clientChan{c: out, ip: ip})
	return nil
}

func Subscribe(event string, out chan string, ip string) error {
	if _, ok := Events[event]; !ok {
		return errors.New("Channel wasn't found")
	}
	Events[event].subscribers = append(Events[event].subscribers, clientChan{c: out, ip: ip})
	for _, cc := range Events[event].subscribers {
		log.Println("Chan " + event + ": client " + cc.ip)
		cc.c <- "New client [" + ip + "] subscribe to event " + event
	}
	return nil
}
