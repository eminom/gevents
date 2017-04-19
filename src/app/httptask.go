package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"
)

var (
	fsHandler   = http.FileServer(http.Dir("resfolder"))
	userAddress = regexp.MustCompile(`^(\d+\.\d+\.\d+\.\d+):(\d+)$`)
	localAddr   = regexp.MustCompile(`^(\[\:\:\d+\])\:(\d+)$`)
)

type HttpTask struct {
	out chan<- interface{}
	svr *http.Server
}

func NewHttpTask(outChan chan<- interface{}) *HttpTask {
	return &HttpTask{out: outChan}
}

func (h *HttpTask) Start(wg *sync.WaitGroup) {
	svr := &http.Server{
		Handler: createMux(h.out),
	}
	l := createListener()
	wg.Add(1)
	go func() {
		defer wg.Done()
		svr.Serve(l)
	}()
	h.svr = svr
}

func (h *HttpTask) Shutdown() {
	if h.svr != nil {
		h.svr.Shutdown(nil)
	}
}

func createListener() *net.TCPListener {
	ta, ea := net.ResolveTCPAddr("tcp", "0.0.0.0:12000")
	if ea != nil {
		panic("No")
	}
	l, el := net.ListenTCP("tcp", ta)
	if el != nil {
		panic("Cannot listen on target address")
	}
	return l
}

func createMux(out chan<- interface{}) *http.ServeMux {
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
		case out <- &AppendMsg{r.RemoteAddr}:
		}
		m1(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		stripperHandler.ServeHTTP(w, r)
	})
	return mux
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
