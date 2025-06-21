# Delivery Management System

A full-stack delivery management system built with Go (Gin) for the backend and Vue (Vite) for the frontend.

---

## Prerequisites

- Go 1.18+
- Node.js 16+
- PostgreSQL

---

## Backend Setup

1. **Clone the repository:**
   ```sh
   git clone https://github.com/yourusername/yourrepo.git
   cd yourrepo
   ```
2. **Configure environment variables:**
    - update `.env` with your database credentials
    ```bash
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=postgres
    DB_PASSWORD=mysecretpassword
    DB_NAME=mydatabase
    DB_SSLMODE=disable
    ```
3. **Backend setup**
    - Install dependencies:
        ```Go
        go mod tidy
        ```

    - Run the backend:
        ```Go
        go run main.go
        ```

    The backend will start on http://localhost:8080.
4. **Frontend Setup**

    - Navigate to the frontend directory:
        ```bash
        cd frontend
        ```

    - Install dependencies:
        ```bash
        npm install
        ```

    - Run the frontend:
        ```bash
        npm run dev
        ```

    The frontend will start on http://localhost:5173.





