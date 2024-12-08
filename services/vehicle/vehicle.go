package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
	_ "github.com/go-sql-driver/mysql" // import lib
	"github.com/gorilla/mux"
	"encoding/base64"
	"github.com/rs/cors"
	"github.com/joho/godotenv"
	"fmt"
)
		

type Vehicle struct {
	VehicleID   int       `json:"vehicle_id"`
	Make        string    `json:"make"`
	Model       string    `json:"model"`
	Year        int       `json:"year"`
	Status      string    `json:"status"` 
	ChargeLevel int       `json:"charge_level"` //declare struct
	Cleanliness string    `json:"cleanliness"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
	Location    string    `json:location`
	VehiclePicture []byte `json:-`
	VehiclePicB64 string  `json:"vehicle_pic"`
}

type Reservation struct {
	ReservationID int    `json:"reservation_id"`
	UserID        int    `json:"user_id"`
	VehicleID     int    `json:"vehicle_id"`
	StartTime     string `json:"start_time"` // Use string for JSON
	EndTime       string `json:"end_time"`   // Use string for JSON
}

// GetAvailableVehicles retrieves available vehicles
func getAvailableVehicles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	var vehicles []Vehicle
	rows, err := db.Query("SELECT * FROM Vehicles WHERE Status = 'Available'") // run sql query
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // if error, error
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cTime string
		var uTime string // declare vars
		var err1 error
		var err2 error
		var vehicle Vehicle
		//insert data
		if err := rows.Scan(&vehicle.VehicleID, &vehicle.Make, &vehicle.Model, &vehicle.Year, &vehicle.Status, &vehicle.ChargeLevel, &vehicle.Cleanliness, &cTime, &uTime, &vehicle.Location, &vehicle.VehiclePicture); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// parse time
		vehicle.UpdatedTime, err1 = time.Parse(time.DateTime, uTime)
		if err1 != nil {
			http.Error(w, err1.Error(), http.StatusInternalServerError)
			return
		}
		vehicle.CreatedTime, err2 = time.Parse(time.DateTime, cTime)
		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
			return
		}
		vehicle.VehiclePicB64 = base64.StdEncoding.EncodeToString(vehicle.VehiclePicture)// get b64 image
		vehicles = append(vehicles, vehicle) // add to list
	}

	w.Header().Set("Content-Type", "application/json") 
	json.NewEncoder(w).Encode(vehicles) // send to client
}

// GetAvailableVehicles retrieves available vehicles
func getAllVehicles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	var vehicles []Vehicle
	rows, err := db.Query("SELECT * FROM Vehicles")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // sql query, if error.
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cTime string
		var uTime string // declare vars.
		var err1 error
		var err2 error
		var vehicle Vehicle
		// get data
		if err := rows.Scan(&vehicle.VehicleID, &vehicle.Make, &vehicle.Model, &vehicle.Year, &vehicle.Status, &vehicle.ChargeLevel, &vehicle.Cleanliness, &cTime, &uTime, &vehicle.Location, &vehicle.VehiclePicture); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// set data
		vehicle.UpdatedTime, err1 = time.Parse(time.DateTime, uTime)
		if err1 != nil {
			http.Error(w, err1.Error(), http.StatusInternalServerError)
			return
		}
		vehicle.CreatedTime, err2 = time.Parse(time.DateTime, cTime)
		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
			return
		}
		vehicle.VehiclePicB64 = base64.StdEncoding.EncodeToString(vehicle.VehiclePicture) // pic into b64
		vehicles = append(vehicles, vehicle) // append to list
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vehicles) // send data
}

// ReserveVehicle handles vehicle reservations
func reserveVehicle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    // set variable
	var reservation Reservation
	if err := json.NewDecoder(r.Body).Decode(&reservation); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // get from client, if err
		return
	}

	// Check if the vehicle is available
	var count int /
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM Reservations 
		WHERE VehicleID = ? AND Status IN ('Confirmed', 'Pending') 
		AND ((StartTime < ? AND EndTime > ?) OR (StartTime < ? AND EndTime > ?))`,
		reservation.VehicleID, reservation.EndTime, reservation.StartTime, reservation.StartTime, reservation.EndTime).Scan(&count)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // if error, error
		return
	}

	if count > 0 { // if not avail, err
		http.Error(w, "Vehicle is not available for the selected time range. Please enter a different time slot. Thank You!", http.StatusConflict)
		return
	}

	// Insert the new reservation and get the reservation ID
	result, err := db.Exec("INSERT INTO Reservations (UserID, VehicleID, StartTime, EndTime, Status) VALUES (?, ?, ?, ?, 'Pending')",
		reservation.UserID, reservation.VehicleID, reservation.StartTime, reservation.EndTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // error out
		return
	}

	// Retrieve the last inserted reservation ID
	reservationID, err := result.LastInsertId() // get reservation id
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // if error, error
		return
	}

	// Update the vehicle status to 'Reserved'
	_, err = db.Exec("UPDATE Vehicles SET Status = 'Reserved' WHERE VehicleID = ?", reservation.VehicleID)
	// set reserved
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // if error.
		return
	}
 
	w.WriteHeader(http.StatusCreated) // else created
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "Vehicle reserved successfully", // send back data
		"reservation_id": reservationID,
	})
}

func modifyReservation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	var reservation Reservation
	if err := json.NewDecoder(r.Body).Decode(&reservation); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Check if the reservation exists
	var existingCount int
	err := db.QueryRow("SELECT COUNT(*) FROM Reservations WHERE ReservationID = ? AND Status IN ('Confirmed', 'Pending')",
		reservation.ReservationID).Scan(&existingCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingCount == 0 {
		http.Error(w, "Reservation not found or already canceled", http.StatusNotFound)
		return
	}

	// 2. Check if the vehicle is available for the new time range
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM Reservations 
		WHERE VehicleID = ? AND Status IN ('Confirmed', 'Pending') 
		AND ReservationID != ? 
		AND ((StartTime < ? AND EndTime > ?) OR (StartTime < ? AND EndTime > ?))`,
		reservation.VehicleID, reservation.ReservationID, reservation.EndTime, reservation.StartTime, reservation.StartTime, reservation.EndTime).Scan(&count)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "Vehicle is not available for the selected time range", http.StatusConflict)
		return
	}

	// 3. Unlock the vehicle during the previous reservation period before modification (if the time range is changed)
	// Note: This assumes you have a mechanism in place to track the old reservation times for rollback purposes
	var oldStartTime, oldEndTime string
	err = db.QueryRow("SELECT StartTime, EndTime FROM Reservations WHERE ReservationID = ?", reservation.ReservationID).Scan(&oldStartTime, &oldEndTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the vehicle status back to 'Available' for the old time range (if it's modified)
	_, err = db.Exec(`
		UPDATE Vehicles 
		SET Status = 'Available' 
		WHERE VehicleID = ? AND NOT EXISTS (
			SELECT 1 
			FROM Reservations 
			WHERE VehicleID = ? 
			AND Status IN ('Confirmed', 'Pending') 
			AND ((StartTime < ? AND EndTime > ?) OR (StartTime < ? AND EndTime > ?))
		)`,
		reservation.VehicleID, reservation.VehicleID, oldEndTime, oldStartTime, oldStartTime, oldEndTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Update the reservation with new times
	_, err = db.Exec("UPDATE Reservations SET StartTime = ?, EndTime = ? WHERE ReservationID = ?",
		reservation.StartTime, reservation.EndTime, reservation.ReservationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Lock the vehicle status as 'Reserved' for the new time range
	_, err = db.Exec("UPDATE Vehicles SET Status = 'Reserved' WHERE VehicleID = ? AND NOT EXISTS (SELECT 1 FROM Reservations WHERE VehicleID = ? AND Status IN ('Confirmed', 'Pending') AND ((StartTime < ? AND EndTime > ?) OR (StartTime < ? AND EndTime > ?)))",
		reservation.VehicleID, reservation.VehicleID, reservation.EndTime, reservation.StartTime, reservation.StartTime, reservation.EndTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success message and the ReservationID
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "Reservation modified successfully",
		"reservation_id": reservation.ReservationID,
	})
}

func cancelReservation(w http.ResponseWriter, r *http.Request) {
    // Get reservation_id from the URL parameters
    vars := mux.Vars(r)
    reservationID := vars["reservation_id"]
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost") // Replace with your actual client origin
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, DELETE, PUT")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
 
    // 1. Check if the reservation exists and is in 'Confirmed' or 'Pending' status
    var existingCount int
    err := db.QueryRow("SELECT COUNT(*) FROM Reservations WHERE ReservationID = ? AND Status IN ('Confirmed', 'Pending')",
        reservationID).Scan(&existingCount)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
 
    if existingCount == 0 {
        http.Error(w, "Reservation not found or already canceled", http.StatusNotFound)
        return
    }
 
    // 2. Get the VehicleID associated with the reservation
    var vehicleID int
    err = db.QueryRow("SELECT VehicleID FROM Reservations WHERE ReservationID = ?", reservationID).Scan(&vehicleID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
 
    // 3. Update the vehicle status back to 'Available'
    _, err = db.Exec("UPDATE Vehicles SET Status = 'Available' WHERE VehicleID = ?", vehicleID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
 
    // 4. Delete the reservation from the database
    _, err = db.Exec("DELETE FROM Reservations WHERE ReservationID = ?", reservationID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
 
    // 5. Send a success response
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Reservation canceled and deleted successfully"})
}
var db *sql.DB

func main() {
	var err error
	var errdb, errenv error
	var env map[string]string
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
	if err != nil {
		log.Fatalln(err)
	}

	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/vehicles", getAvailableVehicles).Methods("GET")
	r.HandleFunc("/allvehicles", getAllVehicles).Methods("GET")
	r.HandleFunc("/reserve", reserveVehicle).Methods("POST")
	r.HandleFunc("/modify", modifyReservation).Methods("PUT")    // New route for modifying reservations
	r.HandleFunc("/cancel/{reservation_id}", cancelReservation).Methods("DELETE") // New route for canceling reservations
	handler := cors.Default().Handler(r)   
	// Start the server
	log.Println("Server is running on port 8081")
	if err := http.ListenAndServe(":8081", handler); err != nil {
		log.Fatal(err)
	}
}
