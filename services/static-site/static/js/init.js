let host = window.location.host // get hostname
let endpoints = {
	"login": `http://${host}:8080/api/v1`,
	"billing": `http://${host}:8082`, // declare endpoints
	"vehicles": `http://${host}:8081`
}