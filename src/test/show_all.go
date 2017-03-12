package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Person struct {
	Name  string
	Phone string
}

func main() {
	sess, err := mgo.Dial(`localhost`)
	defer sess.Close()
	if nil != err {
		panic(err)
	}
	c := sess.DB(`test`).C(`people`)

	var one Person

	//q:=c.Find(bson.M{}) //find all documents
	q := c.Find(bson.M{"name": "Ale"})
	n, _ := q.Count()
	fmt.Printf("%v in all\n", n)

	//~ 遍历:
	iter := q.Iter()
	var res []Person
	iter.All(&res)
	for i := 0; i < len(res); i++ {
		fmt.Printf("[%v]: %v\n", i, res[i])
	}
	fmt.Println("--------------------")
	err = q.One(&one)
	if nil != err {
		fmt.Println("Not a one to be found")
		return
	} else {
		fmt.Printf("one received:%v\n", one)
	}
}
