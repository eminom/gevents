package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"
)

var (
	fsHandler   = http.FileServer(http.Dir("resfolder"))
	userAddress = regexp.MustCompile(`^(\d+\.\d+\.\d+\.\d+):(\d+)$`)
	localAddr   = regexp.MustCompile(`^(\[\:\:\d+\])\:(\d+)$`)
)

type AppendMsg struct {
	IpString string
}

type ReqUpdate struct {
	ClientIP string
	Time     int
}

func dbServiceGo(shutdown <-chan bool, appendMsg <-chan *AppendMsg, wg *sync.WaitGroup) {
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
		case appMsg := <-appendMsg:
			if nil == myCollect {
				break
			}
			ipAddr := filterIP(appMsg.IpString)
			if `` == ipAddr {
			}
			var reqUpdate ReqUpdate
			myCollect.Find(bson.M{"clientip": ipAddr}).Apply(doIncFor, &reqUpdate)
			fmt.Printf("update to %v\n", reqUpdate)
		}
	}
}

func startDbService(shutdown <-chan bool, appendMsg <-chan *AppendMsg, wg *sync.WaitGroup) {
	wg.Add(1)
	go dbServiceGo(shutdown, appendMsg, wg)
}

type fsStru struct{}

func (*fsStru) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fsHandler.ServeHTTP(w, r)
}

var stripperHandler = http.StripPrefix("/resfolder/", &fsStru{})

func filterIP(tcpAddr string) string {
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
	return ipAddr
}

//func StartMiniServer() {
//~ Exactly
/*
	//~ That's all it takes.
		http.HandleFunc("/version.php", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("%s: request for version profile.\n", time.Now())
			//According to http.Redirect's implementation
			//~Method 1 begin
			m1 := func(w http.ResponseWriter, r *http.Request) {
				r2 := new(http.Request)
				*r2 = *r // Copy
				r2.URL = new(url.URL)
				*r2.URL = *r.URL
				r2.URL.Path = "version.txt"
				fsHandler.ServeHTTP(w, r2)
			}
			m2 := func() {
				http.Redirect(w, r, "/resfolder/version.txt", http.StatusMovedPermanently)
			}
			_ = m2
			filterIP(r.RemoteAddr)
			m1(w, r)
			// Method 2:
			//
		})
		//~ The default match
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			//fmt.Printf("In <%s>\n", r.URL.Path)
			stripperHandler.ServeHTTP(w, r)
		})
		http.ListenAndServe(":12000", nil)
*/
//}

func startMiniServer(nuevo chan<- *AppendMsg, wg *sync.WaitGroup) *http.Server {
	wg.Add(1)
	mux := http.NewServeMux()
	mux.HandleFunc("/version.php", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s: request for version profile.\n", time.Now())
		//According to http.Redirect's implementation
		//~Method 1 begin
		m1 := func(w http.ResponseWriter, r *http.Request) {
			r2 := new(http.Request)
			*r2 = *r // Copy
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = "version.txt"
			fsHandler.ServeHTTP(w, r2)
		}
		//~ If react this way, curl won't the final result.
		//~ m1 is better than m2 in debug mode.
		m2 := func() {
			http.Redirect(w, r, "/resfolder/version.txt", http.StatusMovedPermanently)
		}
		_ = m2
		select {
		case nuevo <- &AppendMsg{r.RemoteAddr}:
		}
		m1(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		stripperHandler.ServeHTTP(w, r)
	})
	ta, ea := net.ResolveTCPAddr("tcp", "0.0.0.0:12000")
	if ea != nil {
		panic("No")
	}
	l, el := net.ListenTCP("tcp", ta)
	if el != nil {
		panic("Cannot listen on target address")
	}
	svr := &http.Server{
		Handler: mux,
	}
	go func() {
		defer wg.Done()
		svr.Serve(l)
	}()
	return svr
}

func main() {
	shut, apChan := make(chan bool), make(chan *AppendMsg, 1024)
	var wg sync.WaitGroup
	startDbService(shut, apChan, &wg)
	svr := startMiniServer(apChan, &wg)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Kill, os.Interrupt)
	fmt.Println(<-c)
	shut <- true
	svr.Shutdown(nil) // close http server gracefully.
	fmt.Println("Quitting")
	wg.Wait()
	close(shut)
	close(apChan)
	fmt.Println("task complete")
}
