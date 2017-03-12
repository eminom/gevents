package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"os"
	"regexp"
	"time"
)

var (
	fsHandler   = http.FileServer(http.Dir("resfolder"))
	userAddress = regexp.MustCompile(`^(\d+\.\d+\.\d+\.\d+):(\d+)$`)
	localAddr   = regexp.MustCompile(`^(\[\:\:\d+\])\:(\d+)$`)

	sess        *mgo.Session
	isDBRunning bool
	doIncFor    mgo.Change
	myCollect   *mgo.Collection
)

type ReqUpdate struct {
	ClientIP string
	Time     int
}

func init() {
	var err error
	sess, err = mgo.Dial(`localhost`)
	if nil != err {
		isDBRunning = false
		return
	}
	isDBRunning = true
	myCollect = sess.DB("gevents").C("ips")
	doIncFor = mgo.Change{
		Upsert:    true,
		ReturnNew: true,
		Update:    bson.M{"$inc": bson.M{"time": 1}},
		//~Remove: false
	}

}

type fsStru struct{}

func (*fsStru) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fsHandler.ServeHTTP(w, r)
}

var stripperHandler = http.StripPrefix("/resfolder/", &fsStru{})

func filterIP(tcpAddr string) {
	var ipAddr string = ``
	if ms := userAddress.FindStringSubmatch(tcpAddr); ms != nil {
		fmt.Printf("IP:<%v>, Port<%v>\n", ms[1], ms[2])
		ipAddr = ms[1]
	} else if ms := localAddr.FindStringSubmatch(tcpAddr); ms != nil {
		fmt.Printf("IP:<%v>, Port<%v>\n", ms[1], ms[2])
		ipAddr = ms[1]
	} else {
		fmt.Fprintf(os.Stderr, "Unknown client ip for <%v>\n", tcpAddr)
		fmt.Fprintf(os.Stderr, "%v\n", tcpAddr)
	}
	if `` != ipAddr {
		if isDBRunning {
			var reqUpdate ReqUpdate
			myCollect.Find(bson.M{"clientip": ipAddr}).Apply(doIncFor, &reqUpdate)
			fmt.Printf("update to %v\n", reqUpdate)
		}
	}
}

func main() {
	defer func() {
		//~ But it won't be executed.
		fmt.Println("Quitting")
		if isDBRunning {
			sess.Close()
			fmt.Println("Bye")
		}
	}()

	//~ Exactly
	http.HandleFunc("/version.php", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s: request for version profile.\n", time.Now())
		/*//According to http.Redirect's implementation
		  r2 := new(http.Request)
		  *r2 = *r // Copy
		  r2.URL = new(url.URL)
		  *r2.URL = *r.URL
		  r2.URL.Path = "version.txt"
		  fsHandler.ServeHTTP(w, r2)
		*/
		filterIP(r.RemoteAddr)
		http.Redirect(w, r, "/resfolder/version.txt", http.StatusMovedPermanently)
	})
	//~ The default match
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("In <%s>\n", r.URL.Path)
		stripperHandler.ServeHTTP(w, r)
	})
	http.ListenAndServe(":12000", nil)
}
