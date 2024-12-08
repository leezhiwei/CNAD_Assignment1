package main

import (
 "fmt" 
 "log" 
 "net/http" // import packages
 "os"
 "path"
)

var notFoundFile string = "static/error404.html" // set 404 page

func notFound(w http.ResponseWriter, r *http.Request){ // if 404
	var data []byte // data variable
	var err error // erorr
	data, err = os.ReadFile(notFoundFile) // try to read file
	if err != nil{
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 error not found.") // if error, fallback
		return
	}
	w.Header().Set("Content-Type", "text/html") // else return 404 page
	w.Write(data)
}

func customNotFound(fs http.FileSystem) http.Handler {
    fileServer := http.FileServer(fs) // instantiate fileserver
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _, err := fs.Open(path.Clean(r.URL.Path)) // Do not allow path traversals.
        if os.IsNotExist(err) { // if no file exist
            notFound(w, r) // use above
            return
        }
        fileServer.ServeHTTP(w, r) // else use normal handler
    })
}

func main() { 
	var port int = 80  // normal http port
	fmt.Printf("Server Running on port %d...", port)     // status          
	// print out running server 
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), customNotFound(http.Dir("./static")))) // runs server
	// run server, with fatal logs
} 