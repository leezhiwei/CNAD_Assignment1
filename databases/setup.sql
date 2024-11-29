-- CNAD Assignment 1 (MySQL/MariaDB)
DROP DATABASE IF EXISTS Assignment;  
CREATE DATABASE Assignment;  
USE Assignment;  

CREATE TABLE MembershipTiers (  
   MembershipTierID INT PRIMARY KEY AUTO_INCREMENT,  
   TierName VARCHAR(50) NOT NULL,  
   Benefits TEXT NOT NULL,  
   DiscountRate DECIMAL(10, 2) NOT NULL,  
   BookingLimit INT NOT NULL 
);  

CREATE TABLE Users (  
   UserID INT PRIMARY KEY AUTO_INCREMENT,  
   Email VARCHAR(255) UNIQUE NOT NULL,  
   Phone VARCHAR(20) UNIQUE,  
   PasswordHash VARCHAR(255) NOT NULL,  
   MembershipTierID INT,  
   MembershipPoint INT,  
   CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,  
   UpdatedAt DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,  
   IsVerified BOOLEAN DEFAULT FALSE,  
   FOREIGN KEY (MembershipTierID) REFERENCES MembershipTiers(MembershipTierID)  
);  

CREATE TABLE Vehicles (  
   VehicleID INT PRIMARY KEY AUTO_INCREMENT,  
   Make VARCHAR(50) NOT NULL,  
   Model VARCHAR(50) NOT NULL,  
   Year INT NOT NULL,  
   Status ENUM('Available', 'Reserved', 'In Maintenance') NOT NULL,  
   ChargeLevel INT, -- Assuming a percentage (0-100)  
   Cleanliness ENUM('Clean', 'Needs Cleaning') NOT NULL,  
   CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,  
   UpdatedAt DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP  
);  

CREATE TABLE Reservations (  
   ReservationID INT PRIMARY KEY AUTO_INCREMENT,  
   UserID INT,  
   VehicleID INT,  
   StartTime DATETIME NOT NULL,  
   EndTime DATETIME NOT NULL,  
   Status ENUM('Pending', 'Confirmed', 'Cancelled') NOT NULL,  
   CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,  
   UpdatedAt DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,  
   FOREIGN KEY (UserID) REFERENCES Users(UserID),  
   FOREIGN KEY (VehicleID) REFERENCES Vehicles(VehicleID)  
);  

CREATE TABLE Billing (  
   BillingID INT PRIMARY KEY AUTO_INCREMENT,  
   UserID INT,  
   ReservationID INT,  
   Amount DECIMAL(10, 2) NOT NULL,  
   Status ENUM('Pending', 'Paid', 'Refunded') NOT NULL,  
   CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,  
   UpdatedAt DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,  
   FOREIGN KEY (UserID) REFERENCES Users(UserID),  
   FOREIGN KEY (ReservationID) REFERENCES Reservations(ReservationID)  
);  

CREATE TABLE Payments (  
   PaymentID INT PRIMARY KEY AUTO_INCREMENT,  
   BillingID INT,  
   PaymentMethod ENUM('Credit Card', 'Debit Card', 'PayPal', 'Other') NOT NULL,  
   Amount DECIMAL(10, 2) NOT NULL,  
   PaymentDate DATETIME DEFAULT CURRENT_TIMESTAMP,  
   FOREIGN KEY (BillingID) REFERENCES Billing(BillingID)  
);  

CREATE TABLE Promotions (  
   PromotionID INT PRIMARY KEY AUTO_INCREMENT,  
   Code VARCHAR(50) UNIQUE NOT NULL,  
   DiscountPercentage DECIMAL(5, 2) NOT NULL,  
   StartDate DATETIME NOT NULL,  
   EndDate DATETIME NOT NULL,  
   IsActive BOOLEAN DEFAULT TRUE  
); 

INSERT INTO MembershipTiers 
    ( TierName,Benefits,DiscountRate,BookingLimit ) 
    VALUES 
    ('Basic','Original price and booking limit 8 hours', 0 , 8),
    ('Premium','5% discount and booking limit 12 hour ', 5 , 12), 
    ('VIP','10% discount and booking limit 24 hours', 10 , 2);