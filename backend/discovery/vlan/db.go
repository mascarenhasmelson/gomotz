package vlan

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mascarenhasmelson/gomotz/utils"
)

type ARPScanDevice struct {
	NetworkID     int
	IPAddress     string
	MACAddress    string
	CurrentStatus string
	InterfaceName string
}

type PostgresDB struct {
	pool *pgxpool.Pool
	conn *pgx.Conn
}

func (p *PostgresDB) GetPool() *pgxpool.Pool {
	return p.pool
}
func NewPostgresDB(pool *pgxpool.Pool) (*PostgresDB, error) {
	return &PostgresDB{pool: pool}, nil
}

func (p *PostgresDB) Close() error {
	if p.conn != nil {
		return p.conn.Close(context.Background())
	}
	return nil
}

func (p *PostgresDB) CreateVLANNetwork(ctx context.Context, vlan *utils.VLANNetwork) error {
	if vlan.NetworkMode == "dhcp" {
		query := `
			INSERT INTO vlan_networks (vlan_id, vlan_name, network_mode, monitoring_enabled, scan_interval_seconds)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at
		`
		return p.pool.QueryRow(ctx, query,
			vlan.VLANId, vlan.VLANName, vlan.NetworkMode,
			vlan.MonitoringEnabled, vlan.ScanIntervalSeconds,
		).Scan(&vlan.ID, &vlan.CreatedAt, &vlan.UpdatedAt)
	}

	query := `
		INSERT INTO vlan_networks
		(vlan_id, vlan_name, network_mode, ip_address, cidr_notation, cidr_full, default_gateway, monitoring_enabled, scan_interval_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	return p.pool.QueryRow(ctx, query,
		vlan.VLANId, vlan.VLANName, vlan.NetworkMode,
		vlan.IPAddress, vlan.CIDRNotation, vlan.CIDRFull, vlan.DefaultGateway,
		vlan.MonitoringEnabled, vlan.ScanIntervalSeconds,
	).Scan(&vlan.ID, &vlan.CreatedAt, &vlan.UpdatedAt)
}

// func scanVLANRow(rows interface {
// 	Scan(dest ...interface{}) error
// }) (*utils.VLANNetwork, error) {
// 	network := &utils.VLANNetwork{}
// 	var ipAddress, cidrNotation, cidrFull, defaultGateway string

// 	err := rows.Scan(
// 		&network.ID, &network.VLANId, &network.VLANName, &network.NetworkMode,
// 		&ipAddress, &cidrNotation, &cidrFull, &defaultGateway,
// 		&network.MonitoringEnabled, &network.ScanIntervalSeconds,
// 		&network.CreatedAt, &network.UpdatedAt,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if ipAddress != "" {
// 		network.IPAddress = &ipAddress
// 	}
// 	if cidrNotation != "" {
// 		network.CIDRNotation = &cidrNotation
// 	}
// 	if cidrFull != "" {
// 		network.CIDRFull = &cidrFull
// 	}
// 	if defaultGateway != "" {
// 		network.DefaultGateway = &defaultGateway
// 	}

//		return network, nil
//	}
func scanVLANRow(rows interface {
	Scan(dest ...interface{}) error
}) (*utils.VLANNetwork, error) {
	network := &utils.VLANNetwork{}
	var vlanID *int
	var interfaceName *string // declared but never passed to Scan()!
	var ipAddress, cidrNotation, cidrFull, defaultGateway string

	err := rows.Scan(
		&network.ID, &vlanID, &interfaceName, &network.VLANName, &network.NetworkMode, //  added &interfaceName
		&ipAddress, &cidrNotation, &cidrFull, &defaultGateway,
		&network.MonitoringEnabled, &network.ScanIntervalSeconds,
		&network.CreatedAt, &network.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if vlanID != nil {
		network.VLANId = *vlanID
	}
	if interfaceName != nil {
		network.InterfaceName = *interfaceName
	}
	if ipAddress != "" {
		network.IPAddress = &ipAddress
	}
	if cidrNotation != "" {
		network.CIDRNotation = &cidrNotation
	}
	if cidrFull != "" {
		network.CIDRFull = &cidrFull
	}
	if defaultGateway != "" {
		network.DefaultGateway = &defaultGateway
	}

	return network, nil
}

// const vlanSelectCols = `
//
//	SELECT
//		id, vlan_id, vlan_name, network_mode,
//		COALESCE(ip_address::text, '')     AS ip_address,
//		COALESCE(cidr_notation, '')        AS cidr_notation,
//		COALESCE(cidr_full, '')            AS cidr_full,
//		COALESCE(default_gateway::text, '') AS default_gateway,
//		monitoring_enabled, scan_interval_seconds,
//		created_at, updated_at
//	FROM vlan_networks
//
// `
const vlanSelectCols = `
	SELECT
		id, vlan_id, interface_name, vlan_name, network_mode,
		COALESCE(ip_address::text, '')     AS ip_address,
		COALESCE(cidr_notation, '')        AS cidr_notation,
		COALESCE(cidr_full, '')            AS cidr_full,
		COALESCE(default_gateway::text, '') AS default_gateway,
		monitoring_enabled, scan_interval_seconds,
		created_at, updated_at
	FROM vlan_networks
`

func (p *PostgresDB) GetAllVLANs(ctx context.Context) ([]*utils.VLANNetwork, error) {
	rows, err := p.pool.Query(ctx, vlanSelectCols+"ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var networks []*utils.VLANNetwork
	for rows.Next() {
		n, err := scanVLANRow(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		networks = append(networks, n)
	}
	return networks, rows.Err()
}

func (p *PostgresDB) GetVLANNetwork(ctx context.Context, vlanID int) (*utils.VLANNetwork, error) {
	row := p.pool.QueryRow(ctx, vlanSelectCols+"WHERE vlan_id = $1", vlanID)
	return scanVLANRow(row)
}
func (p *PostgresDB) GetNetworkByID(ctx context.Context, id int) (*utils.VLANNetwork, error) {
	row := p.pool.QueryRow(ctx, vlanSelectCols+"WHERE id = $1", id)
	return scanVLANRow(row)
}

func (p *PostgresDB) DeleteNetwork(ctx context.Context, interfaceName string) error {
	_, err := p.pool.Exec(ctx,
		`DELETE FROM vlan_networks WHERE interface_name = $1`, interfaceName)
	return err
}

func (p *PostgresDB) UpdateNetwork(ctx context.Context, vlan *utils.VLANNetwork) error {
	var vlanID *int
	if vlan.VLANId != 0 {
		vlanID = &vlan.VLANId
	}

	if vlan.NetworkMode == "dhcp" || vlan.NetworkMode == "auto" {
		query := `
            UPDATE vlan_networks
            SET vlan_id = $1, vlan_name = $2, monitoring_enabled = $3, 
                scan_interval_seconds = $4,
                ip_address = NULL, cidr_notation = NULL, cidr_full = NULL, 
                default_gateway = NULL, updated_at = NOW()
            WHERE interface_name = $5
        `
		_, err := p.pool.Exec(ctx, query,
			vlanID, vlan.VLANName, vlan.MonitoringEnabled,
			vlan.ScanIntervalSeconds, vlan.InterfaceName)
		return err
	}

	query := `
        UPDATE vlan_networks
        SET vlan_id = $1, vlan_name = $2, ip_address = $3, cidr_notation = $4, 
            cidr_full = $5, default_gateway = $6, monitoring_enabled = $7, 
            scan_interval_seconds = $8, updated_at = NOW()
        WHERE interface_name = $9
    `
	_, err := p.pool.Exec(ctx, query,
		vlanID, vlan.VLANName, vlan.IPAddress, vlan.CIDRNotation, vlan.CIDRFull,
		vlan.DefaultGateway, vlan.MonitoringEnabled, vlan.ScanIntervalSeconds,
		vlan.InterfaceName)
	return err
}

func (p *PostgresDB) GetEnabledVLANs(ctx context.Context) ([]*utils.VLANNetwork, error) {
	rows, err := p.pool.Query(ctx,
		vlanSelectCols+"WHERE monitoring_enabled = true ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var networks []*utils.VLANNetwork
	for rows.Next() {
		n, err := scanVLANRow(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		networks = append(networks, n)
	}
	return networks, rows.Err()
}

func (p *PostgresDB) UpdateVLANNetwork(ctx context.Context, vlan *utils.VLANNetwork) error {
	if vlan.NetworkMode == "dhcp" {
		query := `
			UPDATE vlan_networks
			SET vlan_name = $1, monitoring_enabled = $2, scan_interval_seconds = $3,
			    ip_address = NULL, cidr_notation = NULL, cidr_full = NULL, default_gateway = NULL
			WHERE vlan_id = $4
		`
		_, err := p.pool.Exec(ctx, query,
			vlan.VLANName, vlan.MonitoringEnabled, vlan.ScanIntervalSeconds, vlan.VLANId)
		return err
	}

	query := `
		UPDATE vlan_networks
		SET vlan_name = $1, ip_address = $2, cidr_notation = $3, cidr_full = $4,
		    default_gateway = $5, monitoring_enabled = $6, scan_interval_seconds = $7
		WHERE vlan_id = $8
	`
	_, err := p.pool.Exec(ctx, query,
		vlan.VLANName, vlan.IPAddress, vlan.CIDRNotation, vlan.CIDRFull,
		vlan.DefaultGateway, vlan.MonitoringEnabled, vlan.ScanIntervalSeconds, vlan.VLANId)
	return err
}

func (p *PostgresDB) DeleteVLANNetwork(ctx context.Context, vlanID int) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM vlan_networks WHERE vlan_id = $1`, vlanID)
	return err
}

func (p *PostgresDB) UpsertDevice(ctx context.Context, device *utils.DiscoveredDevice) error {
	// query := `
	// 	INSERT INTO discovered_devices
	// 	(vlan_id, ip_address, mac_address, hostname, vendor, device_status, first_seen, last_seen)
	// 	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	// 	ON CONFLICT (vlan_id, ip_address)
	// 	DO UPDATE SET
	// 		mac_address   = EXCLUDED.mac_address,
	// 		hostname      = EXCLUDED.hostname,
	// 		vendor        = EXCLUDED.vendor,
	// 		device_status = EXCLUDED.device_status,
	// 		last_seen     = EXCLUDED.last_seen
	// 	RETURNING id, created_at, updated_at
	// `
	// return p.pool.QueryRow(ctx, query,
	// 	device.VLANId, device.IPAddress, device.MACAddress,
	// 	device.Hostname, device.Vendor, device.DeviceStatus,
	// 	device.FirstSeen, device.LastSeen,
	// ).Scan(&device.ID, &device.CreatedAt, &device.UpdatedAt)
	// query := `
	// 	INSERT INTO discovered_devices (
	// 		vlan_id, ip_address, mac_address, hostname, vendor,
	// 		device_status, first_seen, last_seen
	// 	)
	// 	VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	// 	ON CONFLICT (vlan_id, mac_address)
	// 	DO UPDATE SET
	// 		ip_address = EXCLUDED.ip_address,
	// 		hostname = CASE
	// 			WHEN EXCLUDED.hostname IS NOT NULL AND EXCLUDED.hostname != '' THEN EXCLUDED.hostname
	// 			WHEN discovered_devices.hostname IS NOT NULL AND discovered_devices.hostname != '' THEN discovered_devices.hostname
	// 			ELSE EXCLUDED.hostname
	// 		END,
	// 		vendor = CASE
	// 			WHEN EXCLUDED.vendor IS NOT NULL AND EXCLUDED.vendor != '' THEN EXCLUDED.vendor
	// 			WHEN discovered_devices.vendor IS NOT NULL AND discovered_devices.vendor != '' THEN discovered_devices.vendor
	// 			ELSE EXCLUDED.vendor
	// 		END,
	// 		device_status = EXCLUDED.device_status,
	// 		last_seen = NOW()
	// `
	query := `
    INSERT INTO discovered_devices (
        network_id, ip_address, mac_address, hostname, vendor, 
        device_status, first_seen, last_seen
    )
    VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
    ON CONFLICT (network_id, mac_address)
    DO UPDATE SET
        ip_address = EXCLUDED.ip_address,
        hostname = CASE 
            WHEN EXCLUDED.hostname IS NOT NULL AND EXCLUDED.hostname != '' THEN EXCLUDED.hostname
            WHEN discovered_devices.hostname IS NOT NULL AND discovered_devices.hostname != '' THEN discovered_devices.hostname
            ELSE EXCLUDED.hostname
        END,
        vendor = CASE 
            WHEN EXCLUDED.vendor IS NOT NULL AND EXCLUDED.vendor != '' THEN EXCLUDED.vendor
            WHEN discovered_devices.vendor IS NOT NULL AND discovered_devices.vendor != '' THEN discovered_devices.vendor
            ELSE EXCLUDED.vendor
        END,
        device_status = EXCLUDED.device_status,
        last_seen = NOW()
`
	_, err := p.pool.Exec(ctx, query,
		device.NetworkId,
		device.IPAddress,
		device.MACAddress,
		device.Hostname,
		device.Vendor,
		device.DeviceStatus,
	)

	return err

}

// const deviceSelectCols = `
//
//	SELECT
//		id, vlan_id,
//		host(ip_address)      AS ip_address,
//		mac_address,
//		COALESCE(hostname, '') AS hostname,
//		COALESCE(vendor, '')   AS vendor,
//		device_status,
//		first_seen, last_seen, created_at, updated_at
//	FROM discovered_devices
//
// `
const deviceSelectCols = `
    SELECT
        id, network_id,
        host(ip_address)      AS ip_address,
        mac_address,
        COALESCE(hostname, '') AS hostname,
        COALESCE(vendor, '')   AS vendor,
        device_status,
        first_seen, last_seen, created_at, updated_at
    FROM discovered_devices
`

func scanDeviceRows(rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
	Close()
}) ([]*utils.DiscoveredDevice, error) {
	defer rows.Close()

	var devices []*utils.DiscoveredDevice
	for rows.Next() {
		var d utils.DiscoveredDevice
		var hostname, vendor string
		err := rows.Scan(
			&d.ID, &d.NetworkId, &d.IPAddress, &d.MACAddress,
			&hostname, &vendor,
			&d.DeviceStatus, &d.FirstSeen, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		d.Hostname = hostname
		d.Vendor = vendor
		devices = append(devices, &d)
	}
	return devices, rows.Err()
}

func (db *PostgresDB) GetDevicesByVLAN(ctx context.Context, vlanID int) ([]*utils.DiscoveredDevice, error) {
	rows, err := db.pool.Query(ctx,
		deviceSelectCols+"WHERE network_id = $1 ORDER BY last_seen DESC", vlanID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return scanDeviceRows(rows)
}

func (db *PostgresDB) GetAllDevices(ctx context.Context) ([]*utils.DiscoveredDevice, error) {
	rows, err := db.pool.Query(ctx, deviceSelectCols+"ORDER BY last_seen DESC")
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return scanDeviceRows(rows)
}

func (p *PostgresDB) MarkOfflineDevices(ctx context.Context, vlanID int, thresholdMinutes int) (int64, error) {
	query := `
    UPDATE discovered_devices
    SET device_status = 'offline'
    WHERE network_id = $1
      AND device_status NOT IN ('offline', 'conflict')
      AND last_seen < NOW() - ($2 || ' minutes')::INTERVAL
`
	result, err := p.pool.Exec(ctx, query, vlanID, thresholdMinutes)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (p *PostgresDB) CreateScanLog(ctx context.Context, vlanID int) (int64, error) {
	var id int64
	err := p.pool.QueryRow(ctx,
		`INSERT INTO scan_logs (network_id, scan_started_at, scan_status) VALUES ($1, $2, 'running') RETURNING id`,
		vlanID, time.Now(),
	).Scan(&id)
	return id, err
}

func (p *PostgresDB) CompleteScanLog(ctx context.Context, logID int64, totalIPs, found, newDevs, offline, durationMs int) error {
	query := `
		UPDATE scan_logs
		SET scan_completed_at = $1, total_ips_scanned = $2, devices_found = $3,
		    devices_new = $4, devices_offline = $5, scan_duration_ms = $6, scan_status = 'completed'
		WHERE id = $7
	`
	_, err := p.pool.Exec(ctx, query, time.Now(), totalIPs, found, newDevs, offline, durationMs, logID)
	return err
}

func (p *PostgresDB) GetScanLogsByVLAN(ctx context.Context, vlanID int, limit int) ([]*utils.ScanLog, error) {
	query := `
    SELECT id, network_id, scan_started_at, scan_completed_at, total_ips_scanned,
           devices_found, devices_new, devices_offline, scan_duration_ms,
           scan_status, error_message, created_at
    FROM scan_logs
    WHERE network_id = $1
    ORDER BY scan_started_at DESC
    LIMIT $2
`
	rows, err := p.pool.Query(ctx, query, vlanID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*utils.ScanLog
	for rows.Next() {
		sl := &utils.ScanLog{}
		err := rows.Scan(
			&sl.ID, &sl.NetworkId, &sl.ScanStartedAt, &sl.ScanCompletedAt,
			&sl.TotalIPsScanned, &sl.DevicesFound, &sl.DevicesNew, &sl.DevicesOffline,
			&sl.ScanDurationMs, &sl.ScanStatus, &sl.ErrorMessage, &sl.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, sl)
	}
	return logs, rows.Err()
}

func (p *PostgresDB) GetVendorByOUI(ctx context.Context, oui string) (*utils.MACVendor, error) {
	query := `
		SELECT id, oui, vendor_name, fetched_from_api, created_at, updated_at, last_seen
		FROM mac_vendors WHERE oui = $1
	`
	vendor := &utils.MACVendor{}
	err := p.pool.QueryRow(ctx, query, oui).Scan(
		&vendor.ID, &vendor.OUI, &vendor.VendorName, &vendor.FetchedFromAPI,
		&vendor.CreatedAt, &vendor.UpdatedAt, &vendor.LastSeen,
	)
	if err != nil {
		return nil, err
	}
	return vendor, nil
}

func (p *PostgresDB) UpsertVendor(ctx context.Context, vendor *utils.MACVendor) error {
	query := `
		INSERT INTO mac_vendors (oui, vendor_name, fetched_from_api, last_seen)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (oui)
		DO UPDATE SET vendor_name = EXCLUDED.vendor_name, last_seen = EXCLUDED.last_seen, updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	return p.pool.QueryRow(ctx, query,
		vendor.OUI, vendor.VendorName, vendor.FetchedFromAPI, vendor.LastSeen,
	).Scan(&vendor.ID, &vendor.CreatedAt, &vendor.UpdatedAt)
}

func (p *PostgresDB) UpdateVendorLastSeen(ctx context.Context, oui string) error {
	_, err := p.pool.Exec(ctx, `UPDATE mac_vendors SET last_seen = NOW() WHERE oui = $1`, oui)
	return err
}

func (p *PostgresDB) GetAllVendors(ctx context.Context) ([]*utils.MACVendor, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, oui, vendor_name, fetched_from_api, created_at, updated_at, last_seen
		FROM mac_vendors ORDER BY vendor_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vendors []*utils.MACVendor
	for rows.Next() {
		v := &utils.MACVendor{}
		err := rows.Scan(
			&v.ID, &v.OUI, &v.VendorName, &v.FetchedFromAPI,
			&v.CreatedAt, &v.UpdatedAt, &v.LastSeen,
		)
		if err != nil {
			return nil, err
		}
		vendors = append(vendors, v)
	}
	return vendors, rows.Err()
}

func (p *PostgresDB) GetVendorStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*)                                       AS total_vendors,
			COUNT(*) FILTER (WHERE fetched_from_api = true) AS from_api,
			COUNT(*) FILTER (WHERE fetched_from_api = false) AS manual,
			MAX(last_seen)                                 AS most_recent_seen
		FROM mac_vendors
	`
	var total, fromAPI, manual int64
	var mostRecent *time.Time

	err := p.pool.QueryRow(ctx, query).Scan(&total, &fromAPI, &manual, &mostRecent)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_vendors":    total,
		"from_api":         fromAPI,
		"manual":           manual,
		"most_recent_seen": mostRecent,
	}, nil
}

func (p *PostgresDB) DeleteOldVendors(ctx context.Context, daysOld int) (int64, error) {
	result, err := p.pool.Exec(ctx,
		`DELETE FROM mac_vendors WHERE last_seen < NOW() - ($1 || ' days')::INTERVAL`,
		daysOld)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (p *PostgresDB) StartListening(ctx context.Context, connString string, callback func(*utils.DeviceNotification)) error {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return fmt.Errorf("failed to create listener connection: %w", err)
	}
	p.conn = conn

	if _, err := conn.Exec(ctx, "LISTEN device_changes"); err != nil {
		return fmt.Errorf("failed to LISTEN: %w", err)
	}

	log.Println("PostgreSQL listener started on channel 'device_changes'")

	go func() {
		for {
			notification, err := conn.WaitForNotification(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Notification error: %v", err)
				continue
			}

			var n utils.DeviceNotification
			if err := json.Unmarshal([]byte(notification.Payload), &n); err != nil {
				log.Printf("Failed to parse notification: %v", err)
				continue
			}
			callback(&n)
		}
	}()

	return nil
}

// GetVLANNetworkByInterface gets network config by interface name
func (p *PostgresDB) GetVLANNetworkByInterface(ctx context.Context, interfaceName string) (*utils.VLANNetwork, error) {
	row := p.pool.QueryRow(ctx, vlanSelectCols+"WHERE interface_name = $1", interfaceName)
	return scanVLANRow(row)
}

// CreateVLANNetworkByInterface creates a network entry for any interface type
func (p *PostgresDB) CreateVLANNetworkByInterface(ctx context.Context, vlan *utils.VLANNetwork, interfaceName string) error {
	var vlanID *int
	if vlan.VLANId != 0 {
		vlanID = &vlan.VLANId
	}

	if vlan.NetworkMode == "dhcp" || vlan.NetworkMode == "auto" {
		query := `
			INSERT INTO vlan_networks 
			(vlan_id, interface_name, vlan_name, network_mode, monitoring_enabled, scan_interval_seconds)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at, updated_at
		`
		return p.pool.QueryRow(ctx, query,
			vlanID, interfaceName, vlan.VLANName, vlan.NetworkMode,
			vlan.MonitoringEnabled, vlan.ScanIntervalSeconds,
		).Scan(&vlan.ID, &vlan.CreatedAt, &vlan.UpdatedAt)
	}

	query := `
		INSERT INTO vlan_networks
		(vlan_id, interface_name, vlan_name, network_mode, ip_address, cidr_notation, cidr_full, default_gateway, monitoring_enabled, scan_interval_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	return p.pool.QueryRow(ctx, query,
		vlanID, interfaceName, vlan.VLANName, vlan.NetworkMode,
		vlan.IPAddress, vlan.CIDRNotation, vlan.CIDRFull, vlan.DefaultGateway,
		vlan.MonitoringEnabled, vlan.ScanIntervalSeconds,
	).Scan(&vlan.ID, &vlan.CreatedAt, &vlan.UpdatedAt)
}

// UpdateVLANNetworkByInterface updates by interface name
func (p *PostgresDB) UpdateVLANNetworkByInterface(ctx context.Context, vlan *utils.VLANNetwork, interfaceName string) error {
	if vlan.NetworkMode == "auto" {
		//  For auto mode — persist IP/CIDR so recovery works on restart
		query := `
			UPDATE vlan_networks
			SET vlan_name = $1,
				monitoring_enabled = $2,
				scan_interval_seconds = $3,
				ip_address = $4,
				cidr_notation = $5,
				cidr_full = $6,
				default_gateway = $7,
				updated_at = NOW()
			WHERE interface_name = $8
		`
		_, err := p.pool.Exec(ctx, query,
			vlan.VLANName,
			vlan.MonitoringEnabled,
			vlan.ScanIntervalSeconds,
			vlan.IPAddress,
			vlan.CIDRNotation,
			vlan.CIDRFull,
			vlan.DefaultGateway,
			interfaceName,
		)
		return err
	}

	if vlan.NetworkMode == "dhcp" {
		// dhcp — no static IP to store
		query := `
			UPDATE vlan_networks
			SET vlan_name = $1,
				monitoring_enabled = $2,
				scan_interval_seconds = $3,
				ip_address = NULL,
				cidr_notation = NULL,
				cidr_full = NULL,
				default_gateway = NULL,
				updated_at = NOW()
			WHERE interface_name = $4
		`
		_, err := p.pool.Exec(ctx, query,
			vlan.VLANName,
			vlan.MonitoringEnabled,
			vlan.ScanIntervalSeconds,
			interfaceName,
		)
		return err
	}

	// static mode
	query := `
		UPDATE vlan_networks
		SET vlan_name = $1,
			ip_address = $2,
			cidr_notation = $3,
			cidr_full = $4,
			default_gateway = $5,
			monitoring_enabled = $6,
			scan_interval_seconds = $7,
			updated_at = NOW()
		WHERE interface_name = $8
	`
	_, err := p.pool.Exec(ctx, query,
		vlan.VLANName,
		vlan.IPAddress,
		vlan.CIDRNotation,
		vlan.CIDRFull,
		vlan.DefaultGateway,
		vlan.MonitoringEnabled,
		vlan.ScanIntervalSeconds,
		interfaceName,
	)
	return err
}

func (p *PostgresDB) DeleteVLANNetworkByInterface(ctx context.Context, interfaceName string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM vlan_networks WHERE interface_name = $1`, interfaceName)
	return err
}

// db scan arp

func (p *PostgresDB) GetAllDevicesForARPScan(ctx context.Context) ([]*ARPScanDevice, error) {
	query := `
		SELECT 
			d.network_id,
			host(d.ip_address) AS ip_address,
			d.mac_address::text,
			d.device_status,
			v.interface_name
		FROM discovered_devices d
		JOIN vlan_networks v ON v.id = d.network_id
		WHERE d.device_status != 'conflict'   -- skip conflict devices
		  AND v.monitoring_enabled = true      -- only monitored networks
		ORDER BY d.network_id, d.ip_address
	`
	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*ARPScanDevice
	for rows.Next() {
		d := &ARPScanDevice{}
		if err := rows.Scan(
			&d.NetworkID, &d.IPAddress, &d.MACAddress,
			&d.CurrentStatus, &d.InterfaceName,
		); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// UpdateDeviceStatusByNetworkAndMAC updates device status using network_id + mac
func (p *PostgresDB) UpdateDeviceStatusByNetworkAndMAC(ctx context.Context, networkID int, macAddress, status string) error {
	_, err := p.pool.Exec(ctx, `
		UPDATE discovered_devices
		SET device_status = $1, last_seen = NOW()
		WHERE network_id = $2 AND mac_address = $3::macaddr
	`, status, networkID, macAddress)
	return err
}
