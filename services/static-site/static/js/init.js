//curl -X POST http://localhost:8080/api/v1/login -d "{\"email\":\"yqyasd@gmail.com\", \"password\":\"password\", \"totp\": \"{code}\"}"
//curl -X POST http://localhost:8080/api/v1/register -d "{\"email\":\"yqyasd@gmail.com\", \"password\":\"password\", \"phone\":\"1234567890\"}"
let endpoints = {
	"login": "http://localhost:8080/api/v1",
	"billing": "http://localhost:8080/api/billing",
	"vehicles": "http://localhost:8080/api/vehicles"
}