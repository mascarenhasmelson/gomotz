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
	"github.com/mascarenhasmelson/gomotz/monitorsrv"
)

const dbMaxConns = 10

func main() {
	log.Println("-----starting VLAN ARP scanner Service---------")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\n->Shutdown signal received...")
		cancel()
	}()
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://admin:StrongPassword123@localhost:5432/tunnel_services"
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf(" Unable to parse DATABASE URL: %v", err)
	}
	config.MaxConns = int32(dbMaxConns)

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalf(" Unable to connect to database: %v", err)
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
			log.Printf("postgres listen notify error: %v", err)
		}
	}()
	log.Println("PostgreSQL LISTEN/NOTIFY started on 'device_changes' channel")

	parentInterface := os.Getenv("PARENT_INTERFACE")
	if parentInterface == "" {
		parentInterface = "eth0"
	}

	go bgservices.StartPortMonitor(ctx, pool)
	scanManager := vlan.NewVLANScanManager(database, parentInterface)
	log.Println("VLAN scan manager created")
	time.Sleep(500 * time.Millisecond)
	if err := scanManager.RecoverFromRestart(); err != nil {
		log.Printf("  VLAN scanner recovery warning: %v", err)
	} else {
		log.Println("->VLAN scanners recovered from database")
	}
	monitorSvc := monitorsrv.NewPortMonitorService(monitorDB, func(payload []byte) {
		wsHub.Broadcast <- payload
	})

	if err := monitorSvc.RecoverFromDB(); err != nil {
		log.Printf(" Port monitor recovery warning: %v", err)
	} else {
		log.Println("->Port monitors recovered from database")
	}
	snmpSvc := monitorsrv.NewSNMPMonitorService(monitorDB, func(payload []byte) {
		wsHub.Broadcast <- payload
	})
	if err := snmpSvc.RecoverFromDB(); err != nil {
		log.Printf(" SNMP monitor recovery: %v", err)
	} else {
		log.Println("->SNMP monitors recovered from database")
	}
	pingSvc := monitorsrv.NewPingMonitorService(monitorDB, func(payload []byte) {
		wsHub.Broadcast <- payload
	})
	if err := pingSvc.RecoverFromDB(); err != nil {
		log.Printf(" Ping monitor recovery: %v", err)
	} else {
		log.Println("->Ping monitors recovered from database")
	}

	sslSvc := monitorsrv.NewSSLMonitorService(monitorDB, func(payload []byte) {
		wsHub.Broadcast <- payload
	})
	if err := sslSvc.RecoverFromDB(); err != nil {
		log.Printf(" SSL monitor recovery: %v", err)
	} else {
		log.Println("->SSL monitors recovered from database")
	}

	domainSvc := monitorsrv.NewDomainExpiryService(monitorDB, func(payload []byte) {
		wsHub.Broadcast <- payload
	})
	if err := domainSvc.RecoverFromDB(); err != nil {
		log.Printf(" Domain monitor recovery: %v", err)
	} else {
		log.Println("->Domain monitors recovered from database")
	}
	server := &http.Server{
		Addr: ":8082",
		Handler: api.NewRouter(
			ctx,
			pool,
			database,
			monitorDB,
			scanManager,
			monitorSvc,
			snmpSvc,
			pingSvc,
			sslSvc,
			domainSvc,
			wsHub,
		),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		<-ctx.Done()
		log.Println("\n----stopping services...")
		snmpSvc.Shutdown()
		log.Println("->SNMP scanners stopped")
		scanManager.Shutdown()
		log.Println("->VLAN scanners stopped")

		monitorSvc.Shutdown()
		log.Println("->Port monitors stopped")
		pingSvc.Shutdown()
		log.Println("->Ping monitors stopped")

		sslSvc.Shutdown()
		log.Println("->ssl monitors stopped")
		domainSvc.Shutdown()
		log.Println("->domain monitors stopped")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("  HTTP server shutdown error: %v", err)
		}
		log.Println("->HTTP server stopped")
	}()

	log.Println("------------------------------------------------------")
	log.Println(" server running on http://localhost:8082")
	log.Println("------------------------------------------------------")
	log.Println("------------------------------------------------------")
	log.Println(" need sudo root privileges")
	log.Println("------------------------------------------------------")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(" HTTP server error:", err)
	}
	log.Println("->Graceful shutdown complete")
}
