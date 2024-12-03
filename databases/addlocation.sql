-- Add Location of Vehicle
USE Assignment;

ALTER TABLE Vehicles ADD Location ENUM('Carpark', 'RentedOut') DEFAULT 'Carpark' ;  
