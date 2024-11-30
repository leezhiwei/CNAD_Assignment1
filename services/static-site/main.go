package main

import (
 "fmt" 
 "log" 
 "net/http"
 "os"
 "path"
 "io/ioutil"
)

var notFoundFile string = "static/error404.html"

func notFound(w http.ResponseWriter, r *http.Request){
	var data []byte
	var err error
	data, err = ioutil.ReadFile(notFoundFile)
	if err != nil{
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 error not found.")
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func customNotFound(fs http.FileSystem) http.Handler {
    fileServer := http.FileServer(fs)
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _, err := fs.Open(path.Clean(r.URL.Path)) // Do not allow path traversals.
        if os.IsNotExist(err) {
            notFound(w, r)
            return
        }
        fileServer.ServeHTTP(w, r)
    })
}

func main() { 
	var port int = 80 
	fmt.Printf("Server Running on port %d...", port)              
	// print out running server 
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), customNotFound(http.Dir("./static"))))
	// run server, with fatal logs
} 