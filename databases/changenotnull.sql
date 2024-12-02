-- Add Reservations constraint
USE Assignment;
ALTER TABLE Reservations MODIFY UserID int NOT NULL;
ALTER TABLE Reservations MODIFY VehicleID int NOT NULL;