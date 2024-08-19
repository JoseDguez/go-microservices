DELETE USER IF EXISTS 'auth_user'@'localhost';
CREATE USER 'auth_user'@'localhost' IDENTIFIED BY 'Admin123';

DROP DATABASE IF EXISTS auth;
CREATE DATABASE auth;

GRANT ALL PRIVILEGES ON auth.* TO 'auth_user'@'localhost';

USER auth_user;

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
);

INSERT INTO users (user_id, password) VALUES ('jose@email.com', 'Admin123');