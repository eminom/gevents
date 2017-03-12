package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UpdateReq struct {
	ClientIP string
	Time     int
}

func main() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	change := mgo.Change{
		Upsert:    true, //create new if not found
		Update:    bson.M{"$inc": bson.M{"time": 1}},
		ReturnNew: true,
	}
	/*
		do_set := mgo.Change{
			Upsert:    true,
			Update:    bson.M{"$set": bson.M{"time": 0}},
			ReturnNew: true,
		}
	*/
	var req UpdateReq
	c := session.DB("gevents").C("ips")
	//~ The first one: need to be decapitalized.
	q := c.Find(bson.M{"clientip": "[::1]"})
	q.Apply(change, &req)

	/*
		n, _ := q.Count()
		if 0 == n {
			fmt.Println("Not found")
			q.Apply(do_set, &req)
		} else {
			fmt.Println("There")
			q.Apply(change, &req)
		}*/
	fmt.Printf("%v\n", req)
}
