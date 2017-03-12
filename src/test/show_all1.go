package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UpdateReq struct {
	ClieNtIp string
	Time     int
}

func main() {
	sess, err := mgo.Dial(`localhost`)
	defer sess.Close()
	if nil != err {
		panic(err)
	}
	c := sess.DB(`gevents`).C(`ips`)
	q := c.Find(bson.M{}) //find all documents
	n, _ := q.Count()
	fmt.Printf("%v in all\n", n)

	//~ 遍历:
	iter := q.Iter()
	var res []UpdateReq
	iter.All(&res)
	for i := 0; i < len(res); i++ {
		fmt.Printf("[%v]: %v\n", i, res[i])
	}

}
