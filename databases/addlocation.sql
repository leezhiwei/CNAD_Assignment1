-- Add Location of Vehicle
USE Assignment;

ALTER TABLE Vehicles ADD Location ENUM('Carpark', 'Driving') DEFAULT 'Carpark' ;  
