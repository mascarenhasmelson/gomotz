// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/mascarenhasmelson/gomotz/bgservices"
// 	"github.com/mascarenhasmelson/gomotz/discovery/vlan"

// 	"github.com/mascarenhasmelson/gomotz/api"

// 	"github.com/jackc/pgx/v4/pgxpool"
// )

// const dbMaxConns = 10

// func main() {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	sig := make(chan os.Signal, 1)
// 	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

// 	go func() {
// 		<-sig
// 		fmt.Println(" Shutdown signal received...")
// 		cancel()
// 	}()
// 	connString := os.Getenv("DATABASE_URL")
// 	if connString == "" {
// 		connString = "postgres://admin:StrongPassword123@localhost:5432/tunnel_services"
// 	}

// 	config, err := pgxpool.ParseConfig(connString)
// 	if err != nil {
// 		log.Fatalf(" Unable to parse DATABASE URL: %v", err)
// 	}
// 	config.MaxConns = int32(dbMaxConns)

// 	pool, err := pgxpool.ConnectConfig(ctx, config)
// 	if err != nil {
// 		log.Fatalf(" Unable to connect to database: %v", err)
// 	}
// 	defer pool.Close()

// 	fmt.Println("Connected to PostgreSQL!")

// 	database, err := vlan.NewPostgresDB(pool)
// 	if err != nil {
// 		log.Fatalf(" Failed to create VLAN database wrapper: %v", err)
// 	}
// 	defer database.Close()
// 	wsHub := vlan.NewHub()
// 	go wsHub.Run()
// 	log.Println(" WebSocket hub started")

// 	// start postgreSQL LISTEN/NOTIFY for device changes
// 	err = database.StartListening(ctx, connString, wsHub.HandleNotification)
// 	if err != nil {
// 		log.Fatalf(" Failed to start PostgreSQL listener: %v", err)
// 	}
// 	log.Println("PostgreSQL LISTEN/NOTIFY started on 'device_changes' channel")

// 	scanManager := vlan.NewVLANScanManager(database)
// 	log.Println(" VLAN scan manager created")

// 	time.Sleep(500 * time.Millisecond)
// 	if err := scanManager.RecoverFromRestart(); err != nil {
// 		log.Printf("VLAN scanner recovery warning: %v", err)
// 	} else {
// 		log.Println("VLAN scanners recovered from database")
// 	}
// 	go bgservices.StartPortMonitor(ctx, pool)
// 	log.Println("Port monitor started")

// 	server := &http.Server{
// 		Addr:    ":8082",
// 		Handler: api.NewRouter(ctx, pool, database, scanManager, wsHub),
// 	}

// 	go func() {
// 		<-ctx.Done()
// 		fmt.Println("\nStopping services...")

// 		scanManager.Shutdown()
// 		log.Println("VLAN scanners stopped")

// 		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer shutdownCancel()

// 		if err := server.Shutdown(shutdownCtx); err != nil {
// 			log.Printf("HTTP server shutdown error: %v", err)
// 		}
// 		log.Println(" HTTP server stopped")
// 	}()

// 	// fmt.Println(" Server running on http://localhost:8082")
// 	// fmt.Println("WebSocket endpoint: ws://localhost:8082/ws")
// 	// fmt.Println("VLAN API: http://localhost:8082/api/vlans")
// 	// fmt.Println("Devices API: http://localhost:8082/api/devices")

// 	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 		log.Fatal("HTTP server error:", err)
// 	}

// 	fmt.Println("gracefully shhutdown complete.")
// }

// // const dbMaxConns = 5

// // func main() {
// // 	ctx, cancel := context.WithCancel(context.Background())
// // 	defer cancel()
// // 	sig := make(chan os.Signal, 1)
// // 	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
// // 	go func() {
// // 		<-sig
// // 		fmt.Println("Kill...")
// // 		cancel()
// // 	}()
// // 	connString := "postgres://admin:StrongPassword123@localhost:5432/tunnel_services"
// // 	//connString := os.Getenv("DATABASE_URL")

// // 	config, err := pgxpool.ParseConfig(connString)
// // 	if err != nil {
// // 		log.Fatalf("Unable to parse DATABASE URL: %v", err)
// // 	}
// // 	config.MaxConns = int32(dbMaxConns)
// // 	pool, err := pgxpool.ConnectConfig(ctx, config)
// // 	if err != nil {
// // 		log.Fatalf("Unable to connect to database: %v", err)
// // 	}
// // 	defer pool.Close()

// // 	fmt.Println("connected to PostgreSQL!")
// // 	go bgservices.StartPortMonitor(ctx, pool)
// // 	server := &http.Server{
// // 		Addr:    ":8082",
// // 		Handler: api.NewRouter(ctx, pool),
// // 	}
// // 	go func() {
// // 		<-ctx.Done()
// // 		fmt.Println("Stopping Backend server...")
// // 		server.Shutdown(context.Background())
// // 	}()
// // 	fmt.Println("running on http://localhost:8082")

// // 	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// // 		log.Fatal("HTTP server error:", err)
// // 	}

// // }
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mascarenhasmelson/gomotz/api"
	"github.com/mascarenhasmelson/gomotz/bgservices"
	"github.com/mascarenhasmelson/gomotz/discovery/vlan"
)

const dbMaxConns = 10

func main() {
	log.Println("🚀 Starting VLAN ARP Scanner Service...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Println("\n🛑 Shutdown signal received...")
		cancel()
	}()

	// Database connection
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://admin:StrongPassword123@localhost:5432/tunnel_services"
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("❌ Unable to parse DATABASE URL: %v", err)
	}
	config.MaxConns = int32(dbMaxConns)

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalf("❌ Unable to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println(" _____________connected to PostgreSQL________")

	// Create VLAN database wrapper
	database, err := vlan.NewPostgresDB(pool)
	if err != nil {
		log.Fatalf(" Failed to create VLAN database wrapper: %v", err)
	}
	defer database.Close()

	// Create WebSocket hub
	wsHub := vlan.NewHub()
	go wsHub.Run()
	log.Println("WebSocket hub started")

	// Start PostgreSQL LISTEN/NOTIFY for real-time device changes
	// FIX: Run in goroutine to avoid blocking startup
	go func() {
		if err := database.StartListening(ctx, connString, wsHub.HandleNotification); err != nil {
			log.Printf("  PostgreSQL LISTEN/NOTIFY error: %v", err)
		}
	}()
	log.Println("PostgreSQL LISTEN/NOTIFY started on 'device_changes' channel")

	// Create VLAN scan manager
	scanManager := vlan.NewVLANScanManager(database)
	log.Println(" VLAN scan manager created")

	// Small delay to let database settle
	time.Sleep(500 * time.Millisecond)

	// Recover all enabled VLANs from database
	if err := scanManager.RecoverFromRestart(); err != nil {
		log.Printf("⚠️  VLAN scanner recovery warning: %v", err)
	} else {
		log.Println("✅ VLAN scanners recovered from database")
	}

	// Start port monitoring service (background)
	go bgservices.StartPortMonitor(ctx, pool)
	log.Println("✅ Port monitor started")

	// Create HTTP server
	server := &http.Server{
		Addr:         ":8082",
		Handler:      api.NewRouter(ctx, pool, database, scanManager, wsHub),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handler
	go func() {
		<-ctx.Done()
		log.Println("\n🔄 Stopping services...")

		// Stop all VLAN scanners first
		scanManager.Shutdown()
		log.Println("✅ VLAN scanners stopped")

		// Shutdown HTTP server with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("⚠️  HTTP server shutdown error: %v", err)
		}
		log.Println("✅ HTTP server stopped")
	}()

	// Print startup info
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println(" Server running on http://localhost:8082")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println(" WebSocket endpoint:  ws://localhost:8082/ws")
	log.Println(" VLAN API:            http://localhost:8082/api/vlans")
	log.Println(" Devices API:         http://localhost:8082/api/devices")
	log.Println("Status API:          http://localhost:8082/api/status")
	log.Println("Pending VLANs:       http://localhost:8082/api/status/pending")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("  NOTE: ARP scanning requires sudo/root privileges")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Start HTTP server (blocking)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("❌ HTTP server error:", err)
	}

	log.Println("✅ Graceful shutdown complete")
}
