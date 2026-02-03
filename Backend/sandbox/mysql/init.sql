-- MySQL Sandbox Init Script
-- Creates readonly user and sample tables for testing

-- Note: sandbox_user is created via docker environment variables
-- This script just sets up sample tables and grants

-- Sample tables for practice problems
CREATE TABLE IF NOT EXISTS employees (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    department VARCHAR(50),
    salary DECIMAL(10,2),
    hire_date DATE
);

CREATE TABLE IF NOT EXISTS departments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    location VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    customer_name VARCHAR(100),
    product VARCHAR(100),
    quantity INT,
    price DECIMAL(10,2),
    order_date DATE
);

-- Sample data
INSERT INTO departments (name, location) VALUES 
    ('Engineering', 'Building A'),
    ('Sales', 'Building B'),
    ('HR', 'Building C'),
    ('Marketing', 'Building B');

INSERT INTO employees (name, department, salary, hire_date) VALUES 
    ('John Doe', 'Engineering', 75000, '2020-01-15'),
    ('Jane Smith', 'Sales', 65000, '2019-06-01'),
    ('Bob Johnson', 'Engineering', 80000, '2018-03-22'),
    ('Alice Brown', 'HR', 55000, '2021-08-10'),
    ('Charlie Wilson', 'Marketing', 60000, '2020-11-05'),
    ('Diana Ross', 'Engineering', 85000, '2017-02-14'),
    ('Eve Davis', 'Sales', 70000, '2019-09-30');

INSERT INTO orders (customer_name, product, quantity, price, order_date) VALUES 
    ('Customer A', 'Laptop', 2, 1200.00, '2024-01-10'),
    ('Customer B', 'Mouse', 5, 25.00, '2024-01-11'),
    ('Customer A', 'Keyboard', 3, 75.00, '2024-01-12'),
    ('Customer C', 'Monitor', 1, 350.00, '2024-01-15'),
    ('Customer B', 'Laptop', 1, 1200.00, '2024-01-18');

-- Grant SELECT only to sandbox_user (readonly)
GRANT SELECT ON sandbox_db.* TO 'sandbox_user'@'%';
FLUSH PRIVILEGES;
