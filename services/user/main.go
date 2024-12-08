package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"time"
	"strconv"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/beevik/ntp"
	"github.com/joho/godotenv"
)

func print(s any) {
	fmt.Println(s)
}

type User struct {
	UserID           int       `json:"user_id"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	PasswordHash     string    `json:"-"` // Exclude password hash from JSON output for security
	MembershipTierID int       `json:"membership_tier_id"`
	MembershipPoint  int       `json:"membership_point"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	IsVerified       bool      `json:"is_verified"` // for user
	TOTPRandomSecret string    `json:"-"`
}


// generateTOTPWithSecret generates a TOTP using a provided secret, displays it, and generates a QR code.
func generateTOTPWithSecret(email string) *otp.Key {
	key, err := totp.Generate(totp.GenerateOpts{ // totp new 
		Issuer:      "CNAD_Assignment1",
		AccountName: email, // fill info
	})
	if err != nil {
		log.Fatal(err) // if err, return err
	}
	return key // return obj
}

func verifyTOTPwithsecret(secret string, code string) bool {
	ntptime, _ := ntp.Time("0.sg.pool.ntp.org")  // use NTP for precsion
	rv, _ := totp.ValidateCustom( // make a new TOTP obj
		code,
		secret,
		ntptime,
		totp.ValidateOpts{
			Period:    30,
			Skew:      1, // fill in params
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)
	return rv // return if true or false
}


// registerUser handles user registration.
func registerUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS") // cors
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Parse JSON request body
	var reqData struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"` // struct body
		Password string `json:"password"`
	}
	var err error; // error

	err = json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		fmt.Println(err)  						// if invalid
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid JSON data")
		return
	}

	// Validate inputs
	if reqData.Email == "" || reqData.Phone == "" || reqData.Password == "" { // if no valid
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Email, phone, and password are required") // error
		return 
	}

	// Hash the password
	hashedPassword, err := hashPassword(reqData.Password) // hash pass
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // error
		fmt.Fprintf(w, "Error hashing password")
		return
	}

	// Generate and display TOTP and QR code
	key := generateTOTPWithSecret(reqData.Email) // totp obj
	qrdata, err := key.Image(200, 200) // qr data
	var buf bytes.Buffer // qr buffer
	png.Encode(&buf, qrdata) // turn into byte
	if err != nil {
		log.Fatal(err)
	}
	b64str := base64.StdEncoding.EncodeToString(buf.Bytes()) // b64

	// Insert user into database
	query := `
	INSERT INTO Users (Email, Phone, PasswordHash, MembershipTierID, MembershipPoint, CreatedAt, UpdatedAt, IsVerified, TOTPRandomSecret) 
	VALUES (?, ?, ?, 1, 0, NOW(), NOW(), ?, ?)
	` // insert db

	result, err := db.Exec(query, reqData.Email, reqData.Phone, hashedPassword, false, key.Secret())
	// get result
	if err != nil {
		print(err) 
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error inserting user into database") // print err
		return
	}

	userID, _ := result.LastInsertId() // get id

	// Respond with success
	w.WriteHeader(http.StatusOK) // okay
	message := fmt.Sprintf("User registered successfully with ID: %d", userID) // success
	jsondata, err := json.Marshal(map[string]string{"message": message, "qrdata": b64str}) // send success
	if err != nil {
		log.Fatal("Unable to convert to JSON") // error
	}
	w.Write(jsondata) // send json.
} 

// hashPassword hashes a plain text password.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // use bcrypt
	if err != nil {
		return "", err // if err, error
	}
	return string(hash), nil //else okay
}


// loginUser handles user login.
func loginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// Database connection

	// Parse JSON request body
	var reqData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		TOTPCode string `json:"totp"`
	}
	var err error;

	err = json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid JSON data")
		return
	}

	// Validate inputs
	if reqData.Email == "" || reqData.Password == "" || reqData.TOTPCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Email and password are required")
		return
	}

	// Retrieve user details from the database
	var user User
	var TOTPString string // make temp obj
	query := `
        SELECT UserID, Email, PasswordHash, TOTPRandomSecret
        FROM Users
        WHERE Email = ?
    `
    // sql query
	err = db.QueryRow(query, reqData.Email).Scan(&user.UserID, &user.Email, &user.PasswordHash, &TOTPString)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized) // if error, no creds
			fmt.Fprintf(w, "Invalid email or password") 
		} else {
			w.WriteHeader(http.StatusInternalServerError) // or other err
			fmt.Fprintf(w, "Error retrieving user details")
		}
		return
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(reqData.Password)) // check hash
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Invalid email or password") // if err, invalid
		return
	}
	if !verifyTOTPwithsecret(TOTPString, reqData.TOTPCode) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Invalid TOTP") // check TOTP
		return 
	}

	// Store UserID in session or context (example: using a cookie)
	http.SetCookie(w, &http.Cookie{
		Name:    "user_id",
		Value:   fmt.Sprintf("%d", user.UserID), // add cookie
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour), // Cookie expires in 1 day
		SameSite: http.SameSiteLaxMode,

	})

	// Respond with success
	w.WriteHeader(http.StatusOK) // 200 ok
	fmt.Fprintf(w, "Login successful! User ID: %d", user.UserID) // print success
}

// updateUserProfile allows users to update their personal details.
func updateUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// Retrieve the user's ID from cookies
	userIDCookie, err := r.Cookie("user_id") // get cookie
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Unauthorized: Please log in") // if no cookie, err
		return
	}

	userID := userIDCookie.Value // store value

	// Parse JSON request body
	var reqData struct {
		Email string `json:"email"` // store data
		Phone string `json:"phone"`
	}

	err = json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // if bad, print err
		fmt.Fprintf(w, "Invalid JSON data")
		return
	}

	// Validate inputs
	if reqData.Email == "" && reqData.Phone == "" {
		w.WriteHeader(http.StatusBadRequest) // get input
		fmt.Fprintf(w, "At least one field (email or phone) must be provided")
		return
	}

	// Update the user's profile in the database
	query := `
        UPDATE Users
        SET Email = COALESCE(NULLIF(?, ''), Email),
            Phone = COALESCE(NULLIF(?, ''), Phone),
            UpdatedAt = NOW()
        WHERE UserID = ?
    `
    //set sql update
	_, err = db.Exec(query, reqData.Email, reqData.Phone, userID) // run 
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error updating user profile") // if err, error
		return
	}

	w.WriteHeader(http.StatusOK) // else success
	fmt.Fprintf(w, "Profile updated successfully")
}

// viewUserProfile allows users to view their membership status and rental history.
func viewUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Retrieve the user's ID from cookies
	userIDCookie, err := r.Cookie("user_id") // get cookie
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Unauthorized: Please log in") // no cookie
		return
	}

	userID := userIDCookie.Value

	// Query user details
	var user User
	query := `
        SELECT Email, Phone, MembershipTierID, MembershipPoint, CreatedAt, UpdatedAt
        FROM Users
        WHERE UserID = ?
    `

    // sql statement 

	err = db.QueryRow(query, userID).Scan(&user.Email, &user.Phone, &user.MembershipTierID, &user.MembershipPoint, &user.CreatedAt, &user.UpdatedAt)
	user.UserID, _ = strconv.Atoi(userID) // store userid
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error retrieving user profile") // if error, return err
		return
	}

	// Query rental history
	type Reservation struct {
		ReservationID int       `json:"reservation_id"`
		VehicleID     int       `json:"vehicle_id"`
		StartTime     time.Time `json:"start_time"` //res struct
		EndTime       time.Time `json:"end_time"`
		Status        string    `json:"status"`
	}

	reservations := []Reservation{} // make array
	query = `
        SELECT ReservationID, VehicleID, StartTime, EndTime, Status
        FROM Reservations
        WHERE UserID = ?
        ORDER BY CreatedAt DESC
    `
    // sql res
	rows, err := db.Query(query, userID) // 
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error retrieving rental history") // if error
		return
	}
	defer rows.Close() // close rows

	for rows.Next() { //for each row
		var reservation Reservation // new obj
		err := rows.Scan(&reservation.ReservationID, &reservation.VehicleID, &reservation.StartTime, &reservation.EndTime, &reservation.Status)
		// scan in data
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error processing rental history") // if error
			return
		}
		reservations = append(reservations, reservation) // add to array
	}

	// Create response
	response := map[string]interface{}{
		"user":          user,
		"rentalHistory": reservations, // response struct
	}

	w.Header().Set("Content-Type", "application/json") // set type
	json.NewEncoder(w).Encode(response) // send data
}

var db *sql.DB // global var 

func main() {
	var errdb, errenv error // error 
	var env map[string]string // to read env into
	env, errenv = godotenv.Read(".env") // attempt to read env file.
	if errenv != nil {
		log.Fatal("Unable to read env file, error: ", errenv) // print err
	}
	var connstring = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", env["DB_USERNAME"], env["DB_PASSWORD"], env["DB_HOST"], env["DB_PORT"], env["DATABASE_NAME"])
	// connection string
	db, errdb = sql.Open("mysql", connstring) // make sql connection
	if errdb != nil {                           // if error with db
		log.Fatal("Unable to connect to database, error: ", errdb) // print err
	}

	// Set up HTTP router
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/register", registerUser).Methods("POST") // register endpoint
	// Add verification endpoint
	r.HandleFunc("/api/v1/login", loginUser).Methods("POST") // login endpoint
	r.HandleFunc("/api/v1/profile", viewUserProfile).Methods("GET") // get user endpoint
	r.HandleFunc("/api/v1/profile/update", updateUserProfile).Methods("POST") // update endpoint
	// Start HTTP server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
