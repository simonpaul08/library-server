# Library Management System (Server)
A role-based Library Management System with distinct user roles and permissions, allowing admins to onboard and manage readers efficiently. Implements RESTful APIs for user authentication, book inventory, lending workflows, and admin operations. Designed for performance, scalability, and clean code architecture.

# Installation
Clone the repo and install the dependencies(if any).

# Environment Variables 
Create a `.env` file in the root and declare the following variables with appropritae values, check .env.example file for reference

```bash
EMAIL="your_smtp_email_address"
EMAIL_PASSWORD="your_smtp_email_password"
SECRET="jwt_secret"
```

# Run Command 
Use the command below to run the server
```bash
go run main.go
```
