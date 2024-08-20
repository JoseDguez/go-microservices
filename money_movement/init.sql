DROP USER IF EXISTS 'money_movement_user'@'%';
CREATE USER 'money_movement_user'@'%' IDENTIFIED BY 'Admin123';

DROP DATABASE IF EXISTS money_movement;
CREATE DATABASE money_movement;

GRANT ALL PRIVILEGES ON money_movement.* TO 'money_movement_user'@'%';

USE money_movement;

DROP TABLE IF EXISTS wallets;
CREATE TABLE wallets (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    wallet_type VARCHAR(255) NOT NULL,
    INDEX(user_id),
);

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    cents INT NOT NULL DEFAULT 0,
    account_type VARCHAR(255) NOT NULL,
    wallet_id INT NOT NULL,
    FOREIGN KEY (wallet_id) REFERENCES wallet(id)
);

DROP TABLE IF EXISTS transactions;
CREATE TABLE transactions (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    pid VARCHAR(255) NOT NULL,
    src_user_id VARCHAR(255) NOT NULL,
    dst_user_id VARCHAR(255) NOT NULL,
    src_wallet_id INT NOT NULL,
    dst_wallet_id INT NOT NULL,
    src_account_id INT NOT NULL,
    dst_account_id INT NOT NULL,
    src_account_type VARCHAR(255) NOT NULL,
    dst_account_type VARCHAR(255) NOT NULL,
    final_dst_wallet_id INT NOT NULL,
    amount INT NOT NULL,
    INDEX(pid),
);

-- customer and merchant wallets
INSERT INTO wallets(user_id, wallet_type) VALUES ('jose@email.com', 'CUSTOMER');
INSERT INTO wallets(user_id, wallet_type) VALUES ('merchant_id', 'MERCHANT');

-- customer accounts
INSERT INTO accounts(cents, account_type, wallet_id) VALUES (5000000, 'DEFAULT', 1);
INSERT INTO accounts(cents, account_type, wallet_id) VALUES (0, 'PAYMENT', 1);

-- merchant accounts
INSERT INTO accounts(cents, account_type, wallet_id) VALUES (0, 'INCOMING', 2);