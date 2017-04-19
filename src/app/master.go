package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

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

func main() {
	apChan := make(chan interface{}, 1024)
	var wg sync.WaitGroup

	var tasks []Task
	addTask := func(ta ...Task) {
		tasks = append(tasks, ta...)
	}
	addTask(
		NewDbTask(apChan),
		NewHttpTask(apChan),
	)
	for _, t := range tasks {
		t.Start(&wg)
	}

	func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, os.Kill, os.Interrupt)
		fmt.Println(<-c)
	}()
	for _, t := range tasks {
		t.Shutdown()
	}
	fmt.Println("Quitting")
	wg.Wait()
	close(apChan)
	fmt.Println("task complete")
}
