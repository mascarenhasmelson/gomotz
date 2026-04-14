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
	"github.com/mascarenhasmelson/gomotz/discovery/vlan"
	"github.com/mascarenhasmelson/gomotz/monitorsrv"
)

const dbMaxConns = 10

func main() {
	log.Println("->Starting VLAN ARP Scanner Service...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Println("\n->Shutdown signal received...")
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
	log.Println("_____________connected to PostgreSQL________")

	database, err := vlan.NewPostgresDB(pool)
	if err != nil {
		log.Fatalf("Failed to create VLAN database wrapper: %v", err)
	}
	defer database.Close()

	monitorDB, err := monitorsrv.NewPostgresDB(pool)
	if err != nil {
		log.Fatalf("Failed to create monitor database wrapper: %v", err)
	}
	defer monitorDB.Close()

	wsHub := vlan.NewHub()
	go wsHub.Run()
	log.Println("WebSocket hub started")

	go func() {
		if err := database.StartListening(ctx, connString, wsHub.HandleNotification); err != nil {
			log.Printf("PostgreSQL LISTEN/NOTIFY error: %v", err)
		}
	}()
	log.Println("PostgreSQL LISTEN/NOTIFY started on 'device_changes' channel")

	parentInterface := os.Getenv("PARENT_INTERFACE")
	if parentInterface == "" {
		parentInterface = "eth0"
	}
	scanManager := vlan.NewVLANScanManager(database, parentInterface)
	log.Println("VLAN scan manager created")

	time.Sleep(500 * time.Millisecond)

	// ✅ Recover VLAN scanners
	if err := scanManager.RecoverFromRestart(); err != nil {
		log.Printf("⚠️  VLAN scanner recovery warning: %v", err)
	} else {
		log.Println("->VLAN scanners recovered from database")
	}

	// ✅ Port monitor service — inject WS broadcast as a plain function
	monitorSvc := monitorsrv.NewPortMonitorService(monitorDB, func(payload []byte) {
		wsHub.Broadcast <- payload
	})
	if err := monitorSvc.RecoverFromDB(); err != nil {
		log.Printf("⚠️  Port monitor recovery warning: %v", err)
	} else {
		log.Println("->Port monitors recovered from database")
	}

	// ✅ HTTP server
	server := &http.Server{
		Addr: ":8082",
		Handler: api.NewRouter(
			ctx,
			pool,
			database,
			monitorDB, // ✅
			scanManager,
			monitorSvc, // ✅
			wsHub,
		),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ✅ Graceful shutdown
	go func() {
		<-ctx.Done()
		log.Println("\n-> Stopping services...")

		scanManager.Shutdown()
		log.Println("->VLAN scanners stopped")

		monitorSvc.Shutdown() // ✅
		log.Println("->Port monitors stopped")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("⚠️  HTTP server shutdown error: %v", err)
		}
		log.Println("->HTTP server stopped")
	}()

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println(" Server running on http://localhost:8082")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println(" WebSocket:     ws://localhost:8082/ws")
	log.Println(" Monitor WS:    ws://localhost:8082/ws/monitors")
	log.Println(" VLAN API:      http://localhost:8082/v1/api/vlans")
	log.Println(" Devices API:   http://localhost:8082/v1/api/devices")
	log.Println(" Monitors API:  http://localhost:8082/v1/api/monitors")
	log.Println(" Status API:    http://localhost:8082/v1/api/status")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("  NOTE: ARP scanning requires sudo/root privileges")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("❌ HTTP server error:", err)
	}

	log.Println("->Graceful shutdown complete")
}

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

// 	"github.com/jackc/pgx/v4/pgxpool"
// 	"github.com/mascarenhasmelson/gomotz/api"
// 	"github.com/mascarenhasmelson/gomotz/bgservices"
// 	"github.com/mascarenhasmelson/gomotz/discovery/vlan"
// )

// const dbMaxConns = 10

// func main() {
// 	log.Println("->Starting VLAN ARP Scanner Service...")

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	// Signal handling
// 	sig := make(chan os.Signal, 1)
// 	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

// 	go func() {
// 		<-sig
// 		fmt.Println("\n->Shutdown signal received...")
// 		cancel()
// 	}()

// 	// Database connection
// 	connString := os.Getenv("DATABASE_URL")
// 	if connString == "" {
// 		connString = "postgres://admin:StrongPassword123@localhost:5432/tunnel_services"
// 	}

// 	config, err := pgxpool.ParseConfig(connString)
// 	if err != nil {
// 		log.Fatalf("❌ Unable to parse DATABASE URL: %v", err)
// 	}
// 	config.MaxConns = int32(dbMaxConns)

// 	pool, err := pgxpool.ConnectConfig(ctx, config)
// 	if err != nil {
// 		log.Fatalf("❌ Unable to connect to database: %v", err)
// 	}
// 	defer pool.Close()

// 	log.Println(" _____________connected to PostgreSQL________")

// 	// Create VLAN database wrapper
// 	database, err := vlan.NewPostgresDB(pool)
// 	if err != nil {
// 		log.Fatalf(" Failed to create VLAN database wrapper: %v", err)
// 	}
// 	defer database.Close()

// 	// Create WebSocket hub
// 	wsHub := vlan.NewHub()
// 	go wsHub.Run()
// 	log.Println("WebSocket hub started")

// 	// Start PostgreSQL LISTEN/NOTIFY for real-time device changes
// 	// FIX: Run in goroutine to avoid blocking startup
// 	go func() {
// 		if err := database.StartListening(ctx, connString, wsHub.HandleNotification); err != nil {
// 			log.Printf("  PostgreSQL LISTEN/NOTIFY error: %v", err)
// 		}
// 	}()
// 	log.Println("PostgreSQL LISTEN/NOTIFY started on 'device_changes' channel")
// 	parentInterface := os.Getenv("PARENT_INTERFACE")
// 	if parentInterface == "" {
// 		parentInterface = "eth0"
// 	}
// 	// Create VLAN scan manager
// 	scanManager := vlan.NewVLANScanManager(database, parentInterface)
// 	log.Println(" VLAN scan manager created")

// 	// Small delay to let database settle
// 	time.Sleep(500 * time.Millisecond)

// 	// Recover all enabled VLANs from database
// 	if err := scanManager.RecoverFromRestart(); err != nil {
// 		log.Printf("⚠️  VLAN scanner recovery warning: %v", err)
// 	} else {
// 		log.Println("->VLAN scanners recovered from database")
// 	}

// 	// Start port monitoring service (background)
// 	go bgservices.StartPortMonitor(ctx, pool)
// 	log.Println("->Port monitor started")

// 	// Create HTTP server
// 	server := &http.Server{
// 		Addr:         ":8082",
// 		Handler:      api.NewRouter(ctx, pool, database, scanManager, wsHub),
// 		ReadTimeout:  15 * time.Second,
// 		WriteTimeout: 15 * time.Second,
// 		IdleTimeout:  60 * time.Second,
// 	}

// 	// Graceful shutdown handler
// 	go func() {
// 		<-ctx.Done()
// 		log.Println("\n-> Stopping services...")

// 		// Stop all VLAN scanners first
// 		scanManager.Shutdown()
// 		log.Println("->VLAN scanners stopped")

// 		// Shutdown HTTP server with timeout
// 		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer shutdownCancel()

// 		if err := server.Shutdown(shutdownCtx); err != nil {
// 			log.Printf("⚠️  HTTP server shutdown error: %v", err)
// 		}
// 		log.Println("->HTTP server stopped")
// 	}()

// 	// Print startup info
// 	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
// 	log.Println(" Server running on http://localhost:8082")
// 	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
// 	log.Println(" WebSocket endpoint:  ws://localhost:8082/ws")
// 	log.Println(" VLAN API:            http://localhost:8082/api/vlans")
// 	log.Println(" Devices API:         http://localhost:8082/api/devices")
// 	log.Println("Status API:          http://localhost:8082/api/status")
// 	log.Println("Pending VLANs:       http://localhost:8082/api/status/pending")
// 	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
// 	log.Println("  NOTE: ARP scanning requires sudo/root privileges")
// 	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

// 	// Start HTTP server (blocking)
// 	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 		log.Fatal("❌ HTTP server error:", err)
// 	}

// 	log.Println("->Graceful shutdown complete")
// }
