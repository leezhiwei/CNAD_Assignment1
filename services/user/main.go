package main

//pls go get -u golang.org/x/crypto/bcrypt
//curl -X POST http://localhost:8080/api/v1/register -d "{\"email\":\"yqyasd@gmail.com\", \"password\":\"password\", \"phone\":\"1234567890\"}"
//curl -X POST http://localhost:8080/api/v1/register -d "email=yeqiyangasd@gmail.com&password=password123&phone=1234567890
//curl -X POST http://localhost:8080/api/v1/login -d "{\"email\":\"yqyasd@gmail.com\", \"password\":\"password\", \"totp\": \"{code}\"}"
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
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "CNAD_Assignment1",
		AccountName: email,
	})
	if err != nil {
		log.Fatal(err)
	}
	return key
}

func verifyTOTPwithsecret(secret string, code string) bool {
	ntptime, _ := ntp.Time("0.sg.pool.ntp.org") 
	rv, _ := totp.ValidateCustom(
		code,
		secret,
		ntptime,
		totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)
	return rv
}


// registerUser handles user registration.
func registerUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// Database connection
	db, err := sql.Open("mysql", "aime:aime@tcp(127.0.0.1:3306)/Assignment")
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Database connection error")
		return
	}
	defer db.Close()

	// Parse JSON request body
	var reqData struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}

	err = json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid JSON data")
		return
	}

	// Validate inputs
	if reqData.Email == "" || reqData.Phone == "" || reqData.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Email, phone, and password are required")
		return
	}

	// Hash the password
	hashedPassword, err := hashPassword(reqData.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error hashing password")
		return
	}

	// Generate and display TOTP and QR code
	key := generateTOTPWithSecret(reqData.Email)
	qrdata, err := key.Image(200, 200)
	var buf bytes.Buffer
	png.Encode(&buf, qrdata)
	if err != nil {
		log.Fatal(err)
	}
	b64str := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Insert user into database
	query := `
	INSERT INTO Users (Email, Phone, PasswordHash, MembershipTierID, MembershipPoint, CreatedAt, UpdatedAt, IsVerified, TOTPRandomSecret) 
	VALUES (?, ?, ?, 1, 0, NOW(), NOW(), ?, ?)
	`

	result, err := db.Exec(query, reqData.Email, reqData.Phone, hashedPassword, false, key.Secret())
	if err != nil {
		print(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error inserting user into database")
		return
	}

	userID, _ := result.LastInsertId()

	// Respond with success
	w.WriteHeader(http.StatusOK)
	message := fmt.Sprintf("User registered successfully with ID: %d", userID)
	jsondata, err := json.Marshal(map[string]string{"message": message, "qrdata": b64str})
	if err != nil {
		log.Fatal("Unable to convert to JSON")
	}
	w.Write(jsondata)
}

// hashPassword hashes a plain text password.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}


// loginUser handles user login.
func loginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// Database connection
	db, err := sql.Open("mysql", "aime:aime@tcp(127.0.0.1:3306)/Assignment")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Database connection error")
		return
	}
	defer db.Close()

	// Parse JSON request body
	var reqData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		TOTPCode string `json:"totp"`
	}

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
	var TOTPString string
	query := `
        SELECT UserID, Email, PasswordHash, TOTPRandomSecret
        FROM Users
        WHERE Email = ?
    `
	err = db.QueryRow(query, reqData.Email).Scan(&user.UserID, &user.Email, &user.PasswordHash, &TOTPString)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Invalid email or password")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error retrieving user details")
		}
		return
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(reqData.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Invalid email or password")
		return
	}
	fmt.Println(reqData.TOTPCode)
	if !verifyTOTPwithsecret(TOTPString, reqData.TOTPCode) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Invalid TOTP")
		return
	}

	// Store UserID in session or context (example: using a cookie)
	http.SetCookie(w, &http.Cookie{
		Name:    "user_id",
		Value:   fmt.Sprintf("%d", user.UserID),
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour), // Cookie expires in 1 day
		SameSite: http.SameSiteLaxMode,

	})

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Login successful! User ID: %d", user.UserID)
}

// updateUserProfile allows users to update their personal details.
func updateUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// Database connection
	db, err := sql.Open("mysql", "aime:aime@tcp(127.0.0.1:3306)/Assignment")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Database connection error")
		return
	}
	defer db.Close()

	// Retrieve the user's ID from cookies
	userIDCookie, err := r.Cookie("user_id")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Unauthorized: Please log in")
		return
	}

	userID := userIDCookie.Value

	// Parse JSON request body
	var reqData struct {
		Email string `json:"email"`
		Phone string `json:"phone"`
	}

	err = json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid JSON data")
		return
	}

	// Validate inputs
	if reqData.Email == "" && reqData.Phone == "" {
		w.WriteHeader(http.StatusBadRequest)
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
	_, err = db.Exec(query, reqData.Email, reqData.Phone, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error updating user profile")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Profile updated successfully")
}

// viewUserProfile allows users to view their membership status and rental history.
func viewUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// Database connection
	db, err := sql.Open("mysql", "aime:aime@tcp(127.0.0.1:3306)/Assignment?parseTime=true")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Database connection error")
		return
	}
	defer db.Close()

	// Retrieve the user's ID from cookies
	userIDCookie, err := r.Cookie("user_id")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Unauthorized: Please log in")
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
	err = db.QueryRow(query, userID).Scan(&user.Email, &user.Phone, &user.MembershipTierID, &user.MembershipPoint, &user.CreatedAt, &user.UpdatedAt)
	user.UserID, _ = strconv.Atoi(userID)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error retrieving user profile")
		return
	}

	// Query rental history
	type Reservation struct {
		ReservationID int       `json:"reservation_id"`
		VehicleID     int       `json:"vehicle_id"`
		StartTime     time.Time `json:"start_time"`
		EndTime       time.Time `json:"end_time"`
		Status        string    `json:"status"`
	}

	reservations := []Reservation{}
	query = `
        SELECT ReservationID, VehicleID, StartTime, EndTime, Status
        FROM Reservations
        WHERE UserID = ?
        ORDER BY CreatedAt DESC
    `
	rows, err := db.Query(query, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error retrieving rental history")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var reservation Reservation
		err := rows.Scan(&reservation.ReservationID, &reservation.VehicleID, &reservation.StartTime, &reservation.EndTime, &reservation.Status)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error processing rental history")
			return
		}
		reservations = append(reservations, reservation)
	}

	// Create response
	response := map[string]interface{}{
		"user":          user,
		"rentalHistory": reservations,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Database connection setup
	db, err := sql.Open("mysql", "aime:aime@tcp(127.0.0.1:3306)/Assignment")
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	// Set up HTTP router
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/register", registerUser).Methods("POST")
	// Add verification endpoint
	r.HandleFunc("/api/v1/login", loginUser).Methods("POST")
	r.HandleFunc("/api/v1/profile", viewUserProfile).Methods("GET")
	r.HandleFunc("/api/v1/profile/update", updateUserProfile).Methods("POST")
	// Start HTTP server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
