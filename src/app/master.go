
package main

import (
    "fmt"
    "net/http"
    "time"
    _ "net/url"
)

var fsHandler = http.FileServer(http.Dir("resfolder"))
type fsStru struct {}
func (*fsStru)ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fsHandler.ServeHTTP(w, r)
}
var stripperHandler = http.StripPrefix("/resfolder/", &fsStru{})

func main() {
    //~ 精确匹配
    http.HandleFunc("/version.php", func(w http.ResponseWriter, r *http.Request){
        fmt.Printf("%s: request for version profile.\n", time.Now())
        /*
        r2 := new(http.Request)
        *r2 = *r // Copy
        r2.URL = new(url.URL)
        *r2.URL = *r.URL
        r2.URL.Path = "version.txt"
        fsHandler.ServeHTTP(w, r2)
        */
        http.Redirect(w, r, "/resfolder/version.txt", http.StatusMovedPermanently)
    })
    //~ 默认匹配
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
        //fmt.Printf("In <%s>\n", r.URL.Path)
        stripperHandler.ServeHTTP(w, r)
    })
    http.ListenAndServe(":12000", nil)
}

