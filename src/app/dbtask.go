package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

type DbTask struct {
	close chan bool
	inMsg <-chan interface{}
}

func NewDbTask(inMsg <-chan interface{}) *DbTask {
	return &DbTask{
		close: make(chan bool),
		inMsg: inMsg,
	}
}

type AppendMsg struct {
	IpString string
}

type ReqUpdate struct {
	ClientIP string
	Time     int
}

func (d *DbTask) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go dbServiceGo(d.close, d.inMsg, wg)
}

func (d *DbTask) Shutdown() {
	d.close <- true
}

func dbServiceGo(shutdown <-chan bool, inMsg <-chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	var err error
	var sess *mgo.Session
	var doIncFor mgo.Change
	var myCollect *mgo.Collection
	sess, err = mgo.Dial(`localhost`)
	if nil == err {
		defer sess.Close()
		myCollect = sess.DB("gevents").C("ips")
		doIncFor = mgo.Change{
			Upsert:    true,
			ReturnNew: true,
			Update:    bson.M{"$inc": bson.M{"time": 1}},
			//~Remove: false
		}
		fmt.Println(`db ready`)
	} else {
		fmt.Println(`db is unavailable`)
	}
A100:
	for {
		select {
		case <-shutdown:
			break A100
		case msg := <-inMsg:
			if nil == myCollect {
				break
			}
			if am, ok := msg.(*AppendMsg); ok {
				ipAddr := filterIP(am.IpString)
				if `` != ipAddr {
					break
				}
				var reqUpdate ReqUpdate
				myCollect.Find(bson.M{"clientip": ipAddr}).Apply(doIncFor, &reqUpdate)
				fmt.Printf("update to %v\n", reqUpdate)
			}
		}
	}
}
