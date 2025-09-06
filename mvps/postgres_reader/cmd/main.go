package main

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	"postgres_reader/postgres"
	dao "postgres_reader/postgres"
)

func main() {
	slog.Info("Starting PostgreSQL Reader application")

	// Database configuration
	config := &dao.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "meter_db",
		Username: "postgres",
		Password: "password",
		SSLMode:  "disable",
		Pool: dao.PoolConfig{
			MaxOpenConnections:    10,
			MaxIdleConnections:    5,
			ConnectionMaxLifetime: 30 * time.Minute,
		},
	}

	// Build connection string for migrations
	dbURL := config.GetPostgresDSN()
	slog.Info("Database URL", "url", dbURL)

	// Run migrations
	slog.Info("Running database migrations...")
	if err := dao.MigrateDatabase(dbURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	slog.Info("Database migrations completed successfully")

	// Create DAO instance
	slog.Info("Creating PostgreSQL DAO...")
	dao, err := dao.NewPostgresDAO(config)
	if err != nil {
		log.Fatalf("Failed to create PostgreSQL DAO: %v", err)
	}
	defer dao.Close()

	// Test insert meter
	slog.Info("Testing meter insertion...")
	testMeter := &postgres.Meter{
		ID:             "meter-001",
		IPv6:           "2001:db8::1",
		Port:           4059,
		SystemTitle:    "TestMeter",
		AuthPassword:   "secret123",
		AuthKey:        "authkey123",
		BlockCipherKey: "cipherkey123",
	}

	if err := dao.InsertMeter(testMeter); err != nil {
		log.Fatalf("Failed to insert meter: %v", err)
	}
	slog.Info("Meter inserted successfully", "id", testMeter.ID)

	// Test get meter
	slog.Info("Testing meter retrieval...")
	retrievedMeter, err := dao.GetMeter("meter-001")
	if err != nil {
		log.Fatalf("Failed to get meter: %v", err)
	}
	slog.Info("Meter retrieved successfully",
		"id", retrievedMeter.ID,
		"ipv6", retrievedMeter.IPv6,
		"port", retrievedMeter.Port,
		"system_title", retrievedMeter.SystemTitle)

	// Test get all meters
	slog.Info("Testing get all meters...")
	allMeters, err := dao.GetAllMeters()
	if err != nil {
		log.Fatalf("Failed to get all meters: %v", err)
	}
	slog.Info("Retrieved all meters", "count", len(allMeters))

	for _, meter := range allMeters {
		fmt.Printf("Meter: %s - %s:%d\n", meter.ID, meter.IPv6, meter.Port)
	}

	slog.Info("Application completed successfully")
}
