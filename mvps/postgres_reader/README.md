# PostgreSQL Reader

A simple Go application that demonstrates connecting to PostgreSQL, running migrations, and performing basic CRUD operations on meter data.

## Features

- PostgreSQL database connection with connection pooling
- Automatic database migrations using embedded SQL files
- Meter data CRUD operations (Insert, Get, GetAll)
- Simple and clean architecture

## Prerequisites

- Go 1.24+ installed
- Docker and Docker Compose (for running PostgreSQL)

## Quick Start

1. **Start PostgreSQL database:**
   ```bash
   docker-compose up -d
   ```

2. **Run the application:**
   ```bash
   go run cmd/main.go
   ```

## What the application does

1. **Connects to PostgreSQL** using configuration settings
2. **Runs database migrations** to create the `meters` table
3. **Inserts a test meter** record with sample data
4. **Retrieves the meter** by ID to verify insertion
5. **Gets all meters** to show the complete dataset

## Database Schema

The application creates a `meters` table with the following structure:

```sql
CREATE TABLE meters (
    id TEXT PRIMARY KEY,
    ipv6 TEXT NOT NULL,
    port INTEGER NOT NULL,
    system_title TEXT NOT NULL,
    auth_password TEXT NOT NULL,
    auth_key TEXT NOT NULL,
    block_cipher_key TEXT NOT NULL
);
```

## Configuration

The application uses hardcoded configuration for simplicity:
- Host: localhost
- Port: 5432
- Database: meter_db
- Username: postgres
- Password: password

## Project Structure

- `cmd/main.go` - Main application entry point
- `postgres/dao_postgres.go` - Database access layer with CRUD operations
- `postgres/migrate.go` - Database migration functionality
- `postgres/scripts/migrations/` - SQL migration files
- `docker-compose.yaml` - PostgreSQL container configuration
