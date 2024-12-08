package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log" // Please populate .env file with correct creds to DB, run SQL script.
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"strconv"
)

var conn *sql.DB // global connection


type membershipTier struct {
	ID				int		`json:"ID"`
	TierName		string	`json:"TierName"`
	Benefits		string	`json:"Benefits"`
	DiscountRate	float64	`json:"DiscRate"`
	BookingLimit	int		`json:"BookingLimit"`
}

type Payment struct {
    UserID   string `json:"userid"`
    ReservationID    string `json:"resid"`
    Amount   string `json:"amount"`
    PaymentMethod  string `json:"pmethod"`
}


func dbgetalltier() ([]membershipTier, error) {
	var rows, dberr = conn.Query("SELECT * FROM MembershipTiers")          // query
	var returndata []membershipTier = []membershipTier{} // final var
	if dberr != nil {                                              // if error
		return nil, dberr // no result, return err
	}
	defer rows.Close() // close DB
	for rows.Next() {  // iterate row
		var data membershipTier                                                                                   // declare temp vars
		var scanerr error = rows.Scan(&data.ID, &data.TierName, &data.Benefits, &data.DiscountRate, &data.BookingLimit) // scan into memory
		if scanerr != nil {
			return nil, scanerr // if error, return err
		}
		returndata = append(returndata, data) // put into map
	}
	defer rows.Close()     // close DB
	return returndata, nil // if working, return map
}

func dbaddbilling(p Payment) (int64, error) {
	var result sql.Result   // final result
	var dberr, dberr1 error // error variables
	var lastinsertid int64  // set return variable
	result, dberr = conn.Exec("INSERT INTO Billing (UserID, ReservationID, Amount, Status, CreatedAt, UpdatedAt) VALUES (?, ?, ?, 'Paid', NOW(), NOW())", p.UserID, p.ReservationID, p.Amount)
	// query
	if dberr != nil { // if got error
		log.Fatal(dberr) // log this error
		return 0, dberr  // return fn
	}
	lastinsertid, dberr1 = result.LastInsertId() // get lastinsertid
	if dberr1 != nil {                           // if second eror
		log.Fatal(dberr1) // log this error
		return 0, dberr1  // return fn
	}
	return lastinsertid, nil // return fn
}

func dbaddpayment(p Payment, bid int64) (int64, error) {
	var result sql.Result   // final result
	var dberr, dberr1 error // error variables
	var lastinsertid int64  // set return variable
	result, dberr = conn.Exec("INSERT INTO Payments (BillingID, PaymentMethod, Amount, PaymentDate) VALUES (?, ?, ?, NOW())", bid, p.PaymentMethod, p.Amount)
	// query
	if dberr != nil { // if got error
		log.Fatal(dberr) // log this error
		return 0, dberr  // return fn
	}
	lastinsertid, dberr1 = result.LastInsertId() // get lastinsertid
	if dberr1 != nil {                           // if second eror
		log.Fatal(dberr1) // log this error
		return 0, dberr1  // return fn
	}
	return lastinsertid, nil // return fn
}

func dbupdateuser(p Payment) error {
	var dberr error // error variables
	var amt, _ = strconv.Atoi(p.Amount)
	_, dberr = conn.Exec("UPDATE Users SET MembershipPoint = MembershipPoint + ? WHERE UserID = ?", amt * 100, p.UserID)
	// query
	if dberr != nil { // if got error
		log.Fatal(dberr) // log this error
		return dberr  // return fn
	}
	return nil // return fn
}

func dbinctier(p Payment) error{
	var basic int = 100000
	var premium int = 300000
	var vip int = 600000
	var tierid int;
	var result int
	var row = conn.QueryRow("SELECT MembershipPoint FROM Users WHERE UserID = ?", p.UserID) // query
	var queryerr error = row.Scan(&result)                                           // scan single row
	if queryerr != nil { 
		log.Fatal(queryerr)                                                            // if err
		return queryerr
	}
	if result < basic{
		tierid = 1
	} else if (result > basic) && (result < premium){
		tierid = 1
	} else if (result > premium) && (result < vip) {
		tierid = 2
	} else if (result > vip) {
		tierid = 3
	}
	_, queryerr = conn.Exec("UPDATE Users SET MembershipTierID = ? WHERE UserID = ?", tierid , p.UserID);
	if queryerr != nil{
		log.Fatal(queryerr)
		return queryerr
	}
	return nil
}

func dbupdateres(p Payment) error {
	var dberr error // error variables
	_, dberr = conn.Exec("UPDATE Reservations SET Status = 'Confirmed' WHERE ReservationID = ?", p.ReservationID)
	// query
	if dberr != nil { // if got error
		log.Fatal(dberr) // log this error
		return dberr  // return fn
	}
	return nil // return fn
}


func getalltier(w http.ResponseWriter, r *http.Request) {
	var dberr error // declare temp vars
	var dbresult []membershipTier
	dbresult, dberr = dbgetalltier()                  // get from db
	var jsonresponse, jsonError = json.Marshal(dbresult) // try convert to JSON
	if jsonError != nil || dberr != nil {                  // if json error
		w.WriteHeader(http.StatusInternalServerError) // 500 status code
		fmt.Fprintf(w, "500 Internal Server Error")   // send error
		return                                        // end function
	}
	w.Header().Set("Content-Type", "application/json") // set JSON datatype
	w.WriteHeader(http.StatusOK)                       // send 200 OK
	w.Write(jsonresponse)                              // send JSON data
	}

func addpayment(w http.ResponseWriter, r *http.Request) {
	var decoder = json.NewDecoder(r.Body) // instantiate a json decoder from the datastream
	var response Payment               // declare a var to put data in
	var err = decoder.Decode(&response)   // use decoder, put decoded output into the var
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest) // set Bad Request status
		fmt.Fprintf(w, "Bad Request")        // return error
		return
	}
	var billingid, dberr = dbaddbilling(response)
	if dberr != nil { // if cannot find in map
		w.WriteHeader(http.StatusInternalServerError)                // set conflict status
		fmt.Fprintf(w, "Error with adding billing, please try again") // return error
		return                                                       // end fn
	}
	_, dberr = dbaddpayment(response, billingid)
	if dberr != nil { // if cannot find in map
		w.WriteHeader(http.StatusInternalServerError)                // set conflict status
		fmt.Fprintf(w, "Error with adding payment, please try again") // return error
		return                                                       // end fn
	}
	dberr = dbupdateuser(response)
	if dberr != nil { // if cannot find in map
		w.WriteHeader(http.StatusInternalServerError)                // set conflict status
		fmt.Fprintf(w, "Error with updating user, please try again") // return error
		return                                                       // end fn
	}
	dberr =  dbinctier(response)
	if dberr != nil { // if cannot find in map
		w.WriteHeader(http.StatusInternalServerError)                // set conflict status
		fmt.Fprintf(w, "Error with incrementing tier, please try again") // return error
		return                                                       // end fn
	}
	dberr = dbupdateres(response)
	if dberr != nil { // if cannot find in map
		w.WriteHeader(http.StatusInternalServerError)                // set conflict status
		fmt.Fprintf(w, "Error with incrementing tier, please try again") // return error
		return                                                       // end fn
	}
	w.WriteHeader(http.StatusOK) // HTTP accepted
	fmt.Fprintf(w, "Payment Added")     // print out status message
}

func main() { // main
	var errdb, errenv error
	var env map[string]string
	env, errenv = godotenv.Read(".env") // attempt to read env file.
	if errenv != nil {
		log.Fatal("Unable to read env file, error: ", errenv) // print err
	}
	var connstring = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", env["DB_USERNAME"], env["DB_PASSWORD"], env["DB_HOST"], env["DB_PORT"], env["DATABASE_NAME"])
	// connection string
	conn, errdb = sql.Open("mysql", connstring) // make sql connection
	if errdb != nil {                           // if error with db
		log.Fatal("Unable to connect to database, error: ", errdb) // print err
	}
	var router mux.Router = *mux.NewRouter() // create new router
	var port int = 8082                      // declare port
	router.HandleFunc("/gettier", getalltier).Methods("GET")
	router.HandleFunc("/payment", addpayment).Methods("POST")
	var corsobj = cors.New(cors.Options{ // new cors object
		AllowedOrigins: []string{"http://localhost", "http://localhost:80"}, // add allowed hosts
		AllowedMethods: []string{"POST", "GET", "PUT", "DELETE"},
	})

	var handler = corsobj.Handler(&router)                            // add new router handler
	fmt.Printf("Server Running on port %d...", port)                  // print out running server
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler)) // run server, with fatal logs
	defer conn.Close()
}
