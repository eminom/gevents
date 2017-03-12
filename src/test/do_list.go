package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
)

type ReqNow struct {
	ClientIP string
	Time     int
}

func main() {
	sess, err := mgo.Dial(`localhost`)
	if nil != err {
		fmt.Fprintf(os.Stderr, "cannot connect to db")
		os.Exit(-1)
	}
	defer sess.Close()
	c := sess.DB("gevents").C("ips")
	q := c.Find(bson.M{})
	var reqNow []ReqNow
	q.Iter().All(&reqNow)
	fmt.Println("***************")
	for i := 0; i < len(reqNow); i++ {
		fmt.Println(reqNow[i])
	}
}
