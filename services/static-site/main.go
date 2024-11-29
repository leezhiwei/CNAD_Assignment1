package main
import (
 "fmt" 
 "log" 
 "net/http"
)
func main() { 
	fs := http.FileServer(http.Dir("static/")) 
	http.Handle("/", fs) 
	var port int = 80 
	fmt.Printf("Server Running on port %d...", port)              
	// print out running server 
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil)) 
	// run server, with fatal logs
} 