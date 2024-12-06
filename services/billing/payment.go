package main

//curl.exe -X GET "http://localhost:5000/api/v1/payment/{membershipId}"
//curl.exe -X GET "http://localhost:5000/api/v1/payment/jax.doe@example.com/NEWUSER25"

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/customer"
)

// Structure for stripe process payment
type EmailInput struct {
	Email string `json:"email"`
}

type SessionOutput struct {
	Id string `json:"id"`
}

// Structure for promotions
type Promotion struct {
	PromotionID        int       `json:"promotion_id"`        // Primary key, auto-incremented
	Code               string    `json:"code"`                // Unique and not null
	DiscountPercentage float64   `json:"discount_percentage"` // Decimal(5, 2), not null
	StartDate          time.Time `json:"start_date"`          // DATETIME, not null
	EndDate            time.Time `json:"end_date"`            // DATETIME, not null
	IsActive           int       `json:"is_active"`           // Boolean, default true
}

// Structure for bill calculation
type BillCaculation struct {
	UserID             int     `json:"user_id"`             // Identity of user
	Email              string  `json:"email"`               // Email of user
	Phone              string  `json:"phone"`               // Phone contact of user
	MembershipTierID   int     `json:"membership_tier_id"`  // Membership tier of user
	MembershipPoint    int     `json:"membership_point"`    // Membership points of user
	Code               string  `json:"code"`                // Unique and not null
	DiscountPercentage float64 `json:"discount_percentage"` // Decimal(5, 2), not null
	PaymentMethod      string  `json:"payment_method"`      // could be "Credit Card", "Debit Card", "PayPal", or "Other"
	Amount             float64 `json:"amount"`              // Final amount for bill payment
	DiscountRate       float64 `json:"discount_rate"`
	BookingLimit       int     `json:"booking_limit"`
}

// Handle different http methods
func handleMethod(w http.ResponseWriter, r *http.Request, dbUser, dbPassword, dbHost, dbName string) {
	// Handle different methods
	switch r.Method {
	case "GET":
		billCaculation(w, r, dbUser, dbPassword, dbHost, dbName)
	case "PUT":
		// tierUpgrade(w, r, dbUser, dbPassword, dbHost, dbName)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// Billing and Payment Processing

/*

ID 1 = Basic(0 points)
ID 2 = Premium(500 points)
ID 3 =  VIP(1000 Points)
Basic tier: Access to standard vehicles with booking time of 10hrs
Premium tier: Access to premium vehicles with booking time of 48hrs and discount rate of 10%
VIP tier: Access to all vehicles with booking time of 120hrs and discount rate of 20%

*/

// errorStatus logs the error and sends the appropriate HTTP status response.
func errorStatus(w http.ResponseWriter, err error, message string, statusCode int) {
	log.Println(message, err)
	http.Error(w, message, statusCode)
}

// Bill caculation function
func billCaculation(w http.ResponseWriter, r *http.Request, dbUser, dbPassword, dbHost, dbName string) {
	// Temp variable for testing without front-end
	var rentalDuration int = 11
	var vehiclePrice int = 50

	// Create a map based on course detail struct
	paymentDetails := make(map[int]BillCaculation)
	var tempBillCalculation BillCaculation

	// Build Data Source Name for secure database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbHost, dbName)
	// Database connection to MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		// Call errorStatus function to handle error
		errorStatus(w, err, "Database connection error: ", http.StatusInternalServerError)
		return
	}
	// Close database connection when function completes
	defer db.Close()

	// Extract the URL parameters from request r
	var params = mux.Vars(r)
	// Retrieves the membership id from the parameters /payment/{membershipId}
	var membershipId = params["memberShipId"]
	var promoCode = params["promoCode"]

	// Check for the current user to obtain details
	userQuery := "SELECT UserID, Email, Phone, MembershipTierID, MembershipPoint FROM Users WHERE Email = ?"
	userResults, err := db.Query(userQuery, membershipId)
	// Check if the member could be found in the database
	if err != nil {
		fmt.Println("Member not found")
		return
	}
	// Store user data details
	for userResults.Next() {
		// Add data into user detail struct
		if err := userResults.Scan(&tempBillCalculation.UserID, &tempBillCalculation.Email, &tempBillCalculation.Phone, &tempBillCalculation.MembershipTierID, &tempBillCalculation.MembershipPoint); err != nil {
			log.Println("Row scan error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	// Close results when function completes
	defer userResults.Close()
	// Check for hourly rate and bookinglimit of the current user
	memberQuery := "SELECT DiscountRate, BookingLimit FROM membershiptiers WHERE MembershipTierID = ?"
	memberResults, err := db.Query(memberQuery, tempBillCalculation.MembershipTierID)
	// Check if the member could be found in the database
	if err != nil {
		fmt.Println("Membership not found")
		return
	}
	// Store member data details
	for memberResults.Next() {
		// Add data into user detail struct
		if err := memberResults.Scan(&tempBillCalculation.DiscountRate, &tempBillCalculation.BookingLimit); err != nil {
			log.Println("Row scan error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	// Close results when function completes
	defer memberResults.Close()
	// Check if the current user is applicable for discounts
	promoQuery := "SELECT Code, DiscountPercentage, IsActive FROM promotions WHERE Code = ?"
	promoResults, err := db.Query(promoQuery, promoCode)
	// Check if the member could be found in the database
	if err != nil {
		fmt.Println("No promotion code available")
	}
	// Create a tempUser to hold user data
	var tempPromo Promotion
	for promoResults.Next() {
		// Add data into user detail struct
		if err := promoResults.Scan(&tempBillCalculation.Code, &tempBillCalculation.DiscountPercentage, &tempPromo.IsActive); err != nil {
			log.Println("Row scan error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	// Close results when function completes
	defer promoResults.Close()

	// Calculate bill
	if tempPromo.IsActive == 1 {
		billingPayment := rentalDuration * vehiclePrice
		// fmt.Println(rentalDuration)
		// fmt.Println(tempBillCalculation.HourlyRate)
		// fmt.Println(billingPayment)
		finalPayment := (float64(billingPayment) - (float64(billingPayment) * (tempBillCalculation.DiscountPercentage / 100)))
		fmt.Println(finalPayment)
		tempBillCalculation.Amount = finalPayment
	} else {
		finalPayment := rentalDuration * vehiclePrice
		fmt.Println(finalPayment)
		tempBillCalculation.Amount = float64(finalPayment)
	}

	// Add structs with data into paymentDetails map
	paymentDetails[tempBillCalculation.UserID] = tempBillCalculation

	// Secure HTTP headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// Add structs with data into paymentDetails map
	paymentDetails[tempBillCalculation.UserID] = tempBillCalculation
	// Pass the map to the display invoice function
	displayInvoiceAndReceipt(w, paymentDetails)
}

// (Bonus) Stripe payment processing

// Cross Origin Resource Sharing function
func CORSCheck(handler func(w http.ResponseWriter, req *http.Request)) func(w http.ResponseWriter, req *http.Request) {
	res := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
		// Set all access for testing at the moment
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "3600")
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		handler(w, req)
	}
	return res
}

// Vehicle product identifier to stripe sandbox product testing
var VehicleCarId = "price_1QSBSDCZpoakQ6AVCTTEStkv"

// Checkout function
func checkout(email string) (*stripe.CheckoutSession, error) {
	var discounts []*stripe.CheckoutSessionDiscountParams

	discounts = []*stripe.CheckoutSessionDiscountParams{
		&stripe.CheckoutSessionDiscountParams{
			//Coupon: stripe.String(""),
			Coupon: stripe.String("NEWCARERENTAL10"),
		},
	}

	customerParams := &stripe.CustomerParams{
		Email: stripe.String(email),
	}
	customerParams.AddMetadata("UserEmail", email)
	newCustomer, err := customer.New(customerParams)

	if err != nil {
		return nil, err
	}

	meta := map[string]string{
		"UserEmail": email,
	}

	log.Println("Creating meta for user: ", meta)

	params := &stripe.CheckoutSessionParams{
		// Email of user
		Customer: &newCustomer.ID,
		// Webpage referrals
		SuccessURL: stripe.String("http://localhost/"),
		CancelURL:  stripe.String("http://localhost/"),
		// Payment methods for stripe
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		// Discount details
		Discounts: discounts,
		// Checkout payment details with car
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				// Identification for car
				Price: stripe.String(VehicleCarId),
				// Quantity of rentals
				Quantity: stripe.Int64(1),
			},
		},
	}
	return session.New(params)
}

// Checkout with user email function
func CheckoutCreator(w http.ResponseWriter, req *http.Request) {
	input := &EmailInput{}
	err := json.NewDecoder(req.Body).Decode(input)
	if err != nil {
		log.Fatal(err)
	}

	stripeSession, err := checkout(input.Email)
	if err != nil {
		log.Fatal(err)
	}
	err = json.NewEncoder(w).Encode(&SessionOutput{Id: stripeSession.ID})

	if err != nil {
		log.Fatal(err)
	}
}

// Display invoice and receipt
func displayInvoiceAndReceipt(w http.ResponseWriter, paymentDetails map[int]BillCaculation) {
	// Encode the response and handle errors
	if err := json.NewEncoder(w).Encode(paymentDetails); err != nil {
		log.Println("JSON encoding error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	fmt.Println(paymentDetails)
}

func main() {
	// Load the environment credentials
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Get database connection details and stripe API key from environment variables
	dbUser, dbPassword, dbHost, dbName, stripeApiKey := os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"), os.Getenv("STRIPE_BACKEND_KEY")
	// Check if all necessary environment variables are set
	if dbUser == "" || dbPassword == "" || dbHost == "" || dbName == "" || stripeApiKey == "" {
		log.Fatal("Database credentials not fully set in environment variables")
	}
	// Store stripe api key into stripe key struct
	stripe.Key = stripeApiKey

	// Mux router for routing HTTP request
	router := mux.NewRouter()

	// Mux router to map path to different functions
	router.HandleFunc("/api/v1/payment", func(w http.ResponseWriter, r *http.Request) { handleMethod(w, r, dbUser, dbPassword, dbHost, dbName) })
	router.HandleFunc("/api/v1/payment/{memberShipId}", func(w http.ResponseWriter, r *http.Request) { handleMethod(w, r, dbUser, dbPassword, dbHost, dbName) })
	router.HandleFunc("/api/v1/payment/{memberShipId}/{promoCode}", func(w http.ResponseWriter, r *http.Request) { handleMethod(w, r, dbUser, dbPassword, dbHost, dbName) })

	// Payment procesing
	router.HandleFunc("/checkout", CORSCheck(CheckoutCreator))

	// Listen at port 5000
		fmt.Println("Listening at port 8082")
	log.Fatal(http.ListenAndServe(":8082", router))
}
