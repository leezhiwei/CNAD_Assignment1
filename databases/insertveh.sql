-- Move images into /var/lib/mysql in Linux systems to ingest file
INSERT INTO Vehicles (VehicleID, Make, Model, Year, Status, ChargeLevel, Cleanliness, VehiclePicture) VALUES 
(1,'Nissan', 'SKYLINE GT-R [BNR32]', 2020, 'Available', 85, 'Clean', LOAD_FILE('/var/lib/mysql/images/R32.png')),
(2,'Mitsubishi', 'LANCER Evolution IX MR GSR [CT9A]', 2019, 'Reserved', 60, 'Clean', LOAD_FILE('/var/lib/mysql/images/EVO9.png')),
(3,'Nissan', 'SKYLINE GT-R Vspec II [BNR34]', 2021, 'In Maintenance', 0, 'Needs Cleaning', LOAD_FILE('/var/lib/mysql/images/R34.png')),
(4,'Mazda', 'RX-8 Type S [SE3P]', 2023, 'Available', 100, 'Clean', LOAD_FILE('/var/lib/mysql/images/RX8.png'))
