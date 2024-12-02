-- Add TOTP Random Secret
USE Assignment;

ALTER TABLE Users ADD TOTPRandomSecret VARCHAR(50) NOT NULL;  