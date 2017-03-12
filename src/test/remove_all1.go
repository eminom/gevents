package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	sess, err := mgo.Dial(`localhost`)
	if nil != err {
		panic(err)
	}
	defer sess.Close()

	c := sess.DB("gevents").C("ips")
	c.RemoveAll(bson.M{})

	fmt.Println("Removing all done.")
}
