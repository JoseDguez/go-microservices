DROP USER IF EXISTS 'ledger_user'@'localhost';
CREATE USER 'ledger_user'@'%' IDENTIFIED BY 'Admin123';

DROP DATABASE IF EXISTS ledger;
CREATE DATABASE ledger;

GRANT ALL PRIVILEGES ON ledger.* TO 'ledger_user'@'localhost';

USE ledger;

DROP TABLE IF EXISTS ledger;
CREATE TABLE ledger (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    amount INT NOT NULL,
    operation VARCHAR(255) NOT NULL,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);