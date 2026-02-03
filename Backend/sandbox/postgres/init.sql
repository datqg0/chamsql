-- PostgreSQL Sandbox Init Script
-- Creates readonly user and sample tables for testing

-- Create sandbox user with limited permissions
CREATE USER sandbox_user WITH PASSWORD 'sandbox123';

-- Grant connect to database
GRANT CONNECT ON DATABASE sandbox_db TO sandbox_user;

-- Grant usage on schema
GRANT USAGE ON SCHEMA public TO sandbox_user;

-- Grant SELECT on all tables (readonly)
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO sandbox_user;

-- Grant CREATE on schema public (Required for init scripts)
GRANT CREATE ON SCHEMA public TO sandbox_user;

-- Sample tables for practice problems
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    department VARCHAR(50),
    salary DECIMAL(10,2),
    hire_date DATE
);

CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    location VARCHAR(100)
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
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

-- Grant SELECT on existing tables
GRANT SELECT ON ALL TABLES IN SCHEMA public TO sandbox_user;
