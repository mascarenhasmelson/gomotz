-- ============================================
-- DROP EXISTING TABLES (for clean setup)
-- ============================================
DROP TABLE IF EXISTS device_notifications CASCADE;
DROP TABLE IF EXISTS discovered_devices CASCADE;
DROP TABLE IF EXISTS scan_logs CASCADE;
DROP TABLE IF EXISTS services CASCADE;
DROP TABLE IF EXISTS mac_vendors CASCADE;
DROP TABLE IF EXISTS vlan_networks CASCADE;

-- ============================================
-- ENABLE EXTENSIONS
-- ============================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- VLAN NETWORKS TABLE
-- ============================================
CREATE TABLE vlan_networks (
    id SERIAL PRIMARY KEY,
    vlan_id INTEGER UNIQUE NOT NULL CHECK (vlan_id >= 1 AND vlan_id <= 4094),
    vlan_name VARCHAR(100),
    network_mode VARCHAR(10) NOT NULL CHECK (network_mode IN ('static', 'dhcp')),
    
    -- Static configuration fields (NULL if DHCP)
    ip_address INET,
    cidr_notation VARCHAR(50), -- e.g., /24, /16
    cidr_full VARCHAR(50), -- e.g., 192.168.10.0/24
    default_gateway INET,
    
    -- Monitoring/Scanning
    monitoring_enabled BOOLEAN DEFAULT true,
    scan_interval_seconds INTEGER DEFAULT 60, -- 1 minute default between scans
    
    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT static_fields_check CHECK (
        (network_mode = 'dhcp') OR
        (network_mode = 'static' AND ip_address IS NOT NULL AND cidr_notation IS NOT NULL AND cidr_full IS NOT NULL)
    )
);

-- ============================================
-- MAC VENDORS TABLE
-- ============================================
CREATE TABLE mac_vendors (
    id SERIAL PRIMARY KEY,
    oui VARCHAR(6) UNIQUE NOT NULL, -- First 6 hex chars of MAC (e.g., 'FCDDF2')
    vendor_name VARCHAR(255) NOT NULL,
    fetched_from_api BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_seen TIMESTAMP DEFAULT NOW()
);

-- ============================================
-- DISCOVERED DEVICES TABLE
-- ✅ UPDATED: Added 'conflict' status
-- ============================================
CREATE TABLE discovered_devices (
    id SERIAL PRIMARY KEY,
    vlan_id INTEGER NOT NULL REFERENCES vlan_networks(vlan_id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    mac_address MACADDR NOT NULL,
    hostname VARCHAR(255),
    vendor VARCHAR(100),
    -- ✅ UPDATED: Added 'conflict' to CHECK constraint
    device_status VARCHAR(20) NOT NULL DEFAULT 'online' CHECK (device_status IN ('online', 'offline', 'new', 'conflict')),
    
    -- Timestamps
    first_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Unique constraint
    UNIQUE(vlan_id, ip_address)
);

-- ============================================
-- IP CONFLICTS LOG TABLE
-- ✅ NEW: Track conflict history
-- ============================================
CREATE TABLE ip_conflicts (
    id SERIAL PRIMARY KEY,
    vlan_id INTEGER NOT NULL REFERENCES vlan_networks(vlan_id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    conflicting_macs TEXT[] NOT NULL, -- Array of MAC addresses claiming this IP
    detected_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'resolved', 'ignored')),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================
-- DEVICE NOTIFICATIONS TABLE
-- ============================================
CREATE TABLE device_notifications (
    id SERIAL PRIMARY KEY,
    notification_id UUID DEFAULT uuid_generate_v4(),
    vlan_id INTEGER NOT NULL,
    ip_address INET NOT NULL,
    mac_address MACADDR,
    event_type VARCHAR(50) NOT NULL, -- 'new_device', 'status_change', 'device_info_changed', 'device_removed', 'ip_conflict'
    old_status VARCHAR(20),
    new_status VARCHAR(20),
    change_details JSONB,
    notified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================
-- SCAN LOGS TABLE
-- ============================================
CREATE TABLE scan_logs (
    id SERIAL PRIMARY KEY,
    vlan_id INTEGER NOT NULL REFERENCES vlan_networks(vlan_id) ON DELETE CASCADE,
    scan_started_at TIMESTAMP NOT NULL,
    scan_completed_at TIMESTAMP,
    total_ips_scanned INTEGER,
    devices_found INTEGER,
    devices_new INTEGER,
    devices_offline INTEGER,
    conflicts_detected INTEGER DEFAULT 0, -- ✅ NEW: Track conflicts per scan
    scan_duration_ms INTEGER,
    scan_status VARCHAR(20) CHECK (scan_status IN ('running', 'completed', 'failed')),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================
-- SERVICES TABLE (SEPARATE - No VLAN relation)
-- ============================================
CREATE TABLE services (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL,
    local_ip INET NOT NULL,
    local_port INTEGER NOT NULL CHECK (local_port BETWEEN 1 AND 65535),
    remote_ip INET NOT NULL,
    remote_port INTEGER NOT NULL CHECK (remote_port BETWEEN 1 AND 65535),
    online BOOLEAN DEFAULT FALSE,
    last_seen TIMESTAMPTZ DEFAULT NOW(),
    pid INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Unique constraint
    UNIQUE(service_name, local_ip, local_port)
);

-- ============================================
-- INDEXES
-- ============================================

-- MAC Vendors indexes
CREATE INDEX idx_mac_vendors_oui ON mac_vendors(oui);

-- VLAN Networks indexes
CREATE INDEX idx_vlan_networks_monitoring ON vlan_networks(monitoring_enabled);

-- Discovered Devices indexes
CREATE INDEX idx_discovered_devices_vlan ON discovered_devices(vlan_id);
CREATE INDEX idx_discovered_devices_status ON discovered_devices(device_status);
CREATE INDEX idx_discovered_devices_last_seen ON discovered_devices(last_seen);
CREATE INDEX idx_discovered_devices_mac ON discovered_devices(mac_address);
CREATE INDEX idx_discovered_devices_ip ON discovered_devices(ip_address);

-- ✅ NEW: IP Conflicts indexes
CREATE INDEX idx_ip_conflicts_vlan ON ip_conflicts(vlan_id);
CREATE INDEX idx_ip_conflicts_status ON ip_conflicts(status);
CREATE INDEX idx_ip_conflicts_detected ON ip_conflicts(detected_at);
CREATE INDEX idx_ip_conflicts_ip ON ip_conflicts(ip_address);

-- Device Notifications indexes
CREATE INDEX idx_device_notifications_vlan ON device_notifications(vlan_id);
CREATE INDEX idx_device_notifications_notified ON device_notifications(notified);
CREATE INDEX idx_device_notifications_created ON device_notifications(created_at);
CREATE INDEX idx_device_notifications_event_type ON device_notifications(event_type); -- ✅ NEW

-- Scan Logs indexes
CREATE INDEX idx_scan_logs_vlan ON scan_logs(vlan_id);
CREATE INDEX idx_scan_logs_started ON scan_logs(scan_started_at);

-- Services indexes (separate from VLANs)
CREATE INDEX idx_services_online ON services(online);
CREATE INDEX idx_services_local_ip ON services(local_ip);
CREATE INDEX idx_services_remote_ip ON services(remote_ip);
CREATE INDEX idx_services_last_seen ON services(last_seen);
CREATE INDEX idx_services_name ON services(service_name);

-- ============================================
-- TRIGGER FUNCTIONS (CREATE BEFORE TRIGGERS!)
-- ============================================

-- Update timestamp trigger function - DEFINED FIRST
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- NOW create triggers that use this function
CREATE TRIGGER update_vlan_networks_updated_at 
    BEFORE UPDATE ON vlan_networks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_discovered_devices_updated_at 
    BEFORE UPDATE ON discovered_devices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mac_vendors_updated_at 
    BEFORE UPDATE ON mac_vendors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_services_updated_at 
    BEFORE UPDATE ON services
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- DEVICE CHANGES NOTIFICATION TRIGGER
-- ✅ UPDATED: Handle conflict status changes
-- ============================================

CREATE OR REPLACE FUNCTION notify_device_changes()
RETURNS TRIGGER AS $$
DECLARE
    notification JSON;
    event_type TEXT;
    should_notify BOOLEAN := FALSE;
    severity TEXT := 'info';
BEGIN
    -- Determine event type and whether to notify
    IF (TG_OP = 'INSERT') THEN
        event_type := 'new_device';
        should_notify := TRUE;
        
        notification := json_build_object(
            'event_type', event_type,
            'vlan_id', NEW.vlan_id,
            'ip_address', host(NEW.ip_address),  
            'mac_address', NEW.mac_address::text,
            'hostname', COALESCE(NEW.hostname, ''),
            'vendor', COALESCE(NEW.vendor, ''),
            'status', NEW.device_status,
            'severity', 'info',
            'first_seen', to_char(NEW.first_seen AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US'),
            'last_seen', to_char(NEW.last_seen AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US')
        );
        
    ELSIF (TG_OP = 'UPDATE') THEN
        -- ✅ NEW: Detect conflict status change (HIGHEST PRIORITY)
        IF (NEW.device_status = 'conflict' AND OLD.device_status != 'conflict') THEN
            event_type := 'ip_conflict';
            should_notify := TRUE;
            severity := 'critical';
            
            notification := json_build_object(
                'event_type', event_type,
                'vlan_id', NEW.vlan_id,
                'ip_address', host(NEW.ip_address),
                'mac_address', NEW.mac_address::text,
                'hostname', COALESCE(NEW.hostname, ''),
                'vendor', COALESCE(NEW.vendor, ''),
                'old_status', OLD.device_status,
                'new_status', NEW.device_status,
                'severity', severity,
                'message', 'DUPLICATE IP DETECTED: Multiple devices claiming IP ' || host(NEW.ip_address),
                'last_seen', to_char(NEW.last_seen AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US')
            );
            
        -- ✅ NEW: Conflict resolved
        ELSIF (OLD.device_status = 'conflict' AND NEW.device_status != 'conflict') THEN
            event_type := 'conflict_resolved';
            should_notify := TRUE;
            severity := 'warning';
            
            notification := json_build_object(
                'event_type', event_type,
                'vlan_id', NEW.vlan_id,
                'ip_address', host(NEW.ip_address),
                'mac_address', NEW.mac_address::text,
                'hostname', COALESCE(NEW.hostname, ''),
                'vendor', COALESCE(NEW.vendor, ''),
                'old_status', OLD.device_status,
                'new_status', NEW.device_status,
                'severity', severity,
                'message', 'IP conflict resolved for ' || host(NEW.ip_address),
                'last_seen', to_char(NEW.last_seen AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US')
            );
            
        -- Normal status changes
        ELSIF (OLD.device_status != NEW.device_status) THEN
            event_type := 'status_change';  
            should_notify := TRUE;
            
            notification := json_build_object(
                'event_type', event_type,
                'vlan_id', NEW.vlan_id,
                'ip_address', host(NEW.ip_address), 
                'mac_address', NEW.mac_address::text,
                'hostname', COALESCE(NEW.hostname, ''),
                'vendor', COALESCE(NEW.vendor, ''),
                'old_status', OLD.device_status,
                'new_status', NEW.device_status,
                'severity', 'info',
                'last_seen', to_char(NEW.last_seen AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US')
            );
            
        -- Check if MAC, hostname, or vendor changed
        ELSIF (OLD.mac_address != NEW.mac_address OR 
               COALESCE(OLD.hostname, '') != COALESCE(NEW.hostname, '') OR 
               COALESCE(OLD.vendor, '') != COALESCE(NEW.vendor, '')) THEN
            event_type := 'device_info_changed';
            should_notify := TRUE;
            
            notification := json_build_object(
                'event_type', event_type,
                'vlan_id', NEW.vlan_id,
                'ip_address', host(NEW.ip_address), 
                'mac_address', NEW.mac_address::text,
                'hostname', COALESCE(NEW.hostname, ''),
                'vendor', COALESCE(NEW.vendor, ''),
                'status', NEW.device_status,
                'severity', 'info',
                'last_seen', to_char(NEW.last_seen AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US')
            );
        ELSE
            -- Only last_seen updated - don't notify (too noisy)
            should_notify := FALSE;
        END IF;
        
    ELSIF (TG_OP = 'DELETE') THEN
        event_type := 'device_removed';
        should_notify := TRUE;
        
        notification := json_build_object(
            'event_type', event_type,
            'vlan_id', OLD.vlan_id,
            'ip_address', host(OLD.ip_address), 
            'mac_address', OLD.mac_address::text,
            'hostname', COALESCE(OLD.hostname, ''),
            'vendor', COALESCE(OLD.vendor, ''),
            'status', OLD.device_status,
            'severity', 'warning'
        );
    END IF;

    -- Only send notification if it's meaningful
    IF should_notify THEN
        PERFORM pg_notify('device_changes', notification::text);
    END IF;

    IF (TG_OP = 'DELETE') THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER device_changes_trigger
    AFTER INSERT OR UPDATE OR DELETE ON discovered_devices
    FOR EACH ROW EXECUTE FUNCTION notify_device_changes();

-- ============================================
-- SERVICE CHANGE NOTIFICATION TRIGGER
-- ============================================

CREATE OR REPLACE FUNCTION notify_service_change()
RETURNS TRIGGER AS $$
DECLARE
    notification_payload JSON;
    event_type_val VARCHAR(50);
BEGIN
    IF (TG_OP = 'INSERT') THEN
        event_type_val := 'service_created';
        
        notification_payload := json_build_object(
            'event_type', event_type_val,
            'service_id', NEW.id,
            'service_name', NEW.service_name,
            'local_ip', NEW.local_ip::text,
            'local_port', NEW.local_port,
            'remote_ip', NEW.remote_ip::text,
            'remote_port', NEW.remote_port,
            'online', NEW.online,
            'pid', NEW.pid
        );
        
    ELSIF (TG_OP = 'UPDATE') THEN
        IF (OLD.online != NEW.online) THEN
            event_type_val := 'service_status_change';
        ELSE
            event_type_val := 'service_updated';
        END IF;
        
        notification_payload := json_build_object(
            'event_type', event_type_val,
            'service_id', NEW.id,
            'service_name', NEW.service_name,
            'local_ip', NEW.local_ip::text,
            'local_port', NEW.local_port,
            'remote_ip', NEW.remote_ip::text,
            'remote_port', NEW.remote_port,
            'online', NEW.online,
            'old_online', OLD.online,
            'pid', NEW.pid,
            'last_seen', NEW.last_seen
        );
        
    ELSIF (TG_OP = 'DELETE') THEN
        event_type_val := 'service_deleted';
        
        notification_payload := json_build_object(
            'event_type', event_type_val,
            'service_id', OLD.id,
            'service_name', OLD.service_name,
            'local_ip', OLD.local_ip::text,
            'local_port', OLD.local_port
        );
    END IF;
    
    -- Send PostgreSQL NOTIFY
    PERFORM pg_notify('service_change', notification_payload::text);
    
    IF (TG_OP = 'DELETE') THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER service_change_trigger
    AFTER INSERT OR UPDATE OR DELETE ON services
    FOR EACH ROW EXECUTE FUNCTION notify_service_change();

-- ============================================
-- HELPER FUNCTIONS
-- ============================================

-- Function to mark old devices as offline
CREATE OR REPLACE FUNCTION mark_offline_devices(p_vlan_id INTEGER, p_threshold_minutes INTEGER DEFAULT 5)
RETURNS INTEGER AS $$
DECLARE
    affected_rows INTEGER;
BEGIN
    UPDATE discovered_devices
    SET device_status = 'offline'
    WHERE vlan_id = p_vlan_id
    AND device_status NOT IN ('offline', 'conflict') -- ✅ Don't override conflicts
    AND last_seen < NOW() - (p_threshold_minutes || ' minutes')::INTERVAL;
    
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
    RETURN affected_rows;
END;
$$ LANGUAGE plpgsql;

-- Function to mark old services as offline (SEPARATE from VLANs)
CREATE OR REPLACE FUNCTION mark_offline_services(p_threshold_minutes INTEGER DEFAULT 5)
RETURNS INTEGER AS $$
DECLARE
    affected_rows INTEGER;
BEGIN
    UPDATE services
    SET online = false
    WHERE online = true
    AND last_seen < NOW() - (p_threshold_minutes || ' minutes')::INTERVAL;
    
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
    RETURN affected_rows;
END;
$$ LANGUAGE plpgsql;

-- ✅ NEW: Function to get device count by VLAN (with conflict count)
CREATE OR REPLACE FUNCTION get_device_count_by_vlan(p_vlan_id INTEGER)
RETURNS TABLE(
    total_devices BIGINT,
    online_devices BIGINT,
    offline_devices BIGINT,
    new_devices BIGINT,
    conflict_devices BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*) as total_devices,
        COUNT(*) FILTER (WHERE device_status = 'online') as online_devices,
        COUNT(*) FILTER (WHERE device_status = 'offline') as offline_devices,
        COUNT(*) FILTER (WHERE device_status = 'new') as new_devices,
        COUNT(*) FILTER (WHERE device_status = 'conflict') as conflict_devices
    FROM discovered_devices
    WHERE vlan_id = p_vlan_id;
END;
$$ LANGUAGE plpgsql;

-- ✅ NEW: Function to get active conflicts
CREATE OR REPLACE FUNCTION get_active_conflicts()
RETURNS TABLE(
    vlan_id INTEGER,
    ip_address INET,
    mac_address TEXT,
    hostname TEXT,
    vendor TEXT,
    last_seen TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        d.vlan_id,
        d.ip_address,
        d.mac_address::text,
        COALESCE(d.hostname, '') as hostname,
        COALESCE(d.vendor, '') as vendor,
        d.last_seen
    FROM discovered_devices d
    WHERE d.device_status = 'conflict'
    ORDER BY d.last_seen DESC;
END;
$$ LANGUAGE plpgsql;

-- Function to get all online services (SEPARATE from VLANs)
CREATE OR REPLACE FUNCTION get_online_services()
RETURNS TABLE(
    id INTEGER,
    service_name VARCHAR(100),
    local_ip INET,
    local_port INTEGER,
    remote_ip INET,
    remote_port INTEGER,
    pid INTEGER,
    last_seen TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.id,
        s.service_name,
        s.local_ip,
        s.local_port,
        s.remote_ip,
        s.remote_port,
        s.pid,
        s.last_seen
    FROM services s
    WHERE s.online = true
    ORDER BY s.last_seen DESC;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- TABLE COMMENTS
-- ============================================

COMMENT ON TABLE vlan_networks IS 'VLAN network configurations for ARP scanning';
COMMENT ON TABLE discovered_devices IS 'Devices discovered via ARP scanning on VLANs';
COMMENT ON TABLE ip_conflicts IS 'History of IP address conflicts detected on the network'; -- ✅ NEW
COMMENT ON TABLE device_notifications IS 'Device state change notifications for real-time updates';
COMMENT ON TABLE scan_logs IS 'History of ARP scan operations per VLAN';
COMMENT ON TABLE services IS 'Port forwarding/tunneling services - SEPARATE from VLANs';
COMMENT ON TABLE mac_vendors IS 'MAC address vendor lookup cache - reduces API calls';

COMMENT ON COLUMN vlan_networks.network_mode IS 'static: manual IP config, dhcp: automatic from DHCP server';
COMMENT ON COLUMN vlan_networks.cidr_notation IS 'CIDR prefix only, e.g., /24';
COMMENT ON COLUMN vlan_networks.cidr_full IS 'Full CIDR notation, e.g., 192.168.10.0/24';
COMMENT ON COLUMN vlan_networks.scan_interval_seconds IS 'Time between complete subnet scans (default 60 seconds)';

COMMENT ON COLUMN discovered_devices.device_status IS 'online: actively responding, offline: not seen recently, new: just discovered, conflict: duplicate IP detected'; -- ✅ UPDATED
COMMENT ON COLUMN discovered_devices.last_seen IS 'Last time device responded to ARP request';

COMMENT ON COLUMN ip_conflicts.conflicting_macs IS 'Array of MAC addresses that were detected claiming this IP address'; -- ✅ NEW

COMMENT ON COLUMN services.online IS 'Whether the service/tunnel is currently active';
COMMENT ON COLUMN services.pid IS 'Process ID of the running service/tunnel';
COMMENT ON COLUMN services.local_ip IS 'Local IP address for the service';
COMMENT ON COLUMN services.remote_ip IS 'Remote IP address the service connects to';

COMMENT ON COLUMN mac_vendors.oui IS 'Organizationally Unique Identifier - first 6 hex chars of MAC';
COMMENT ON COLUMN mac_vendors.fetched_from_api IS 'True if fetched from API, false if manually added';