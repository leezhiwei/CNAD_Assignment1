let host = window.location.host
let endpoints = {
	"login": `http://${host}:8080/api/v1`,
	"billing": `http://${host}:8082`,
	"vehicles": `http://${host}:8081`
}