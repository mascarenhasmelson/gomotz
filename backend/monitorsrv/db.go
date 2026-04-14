package monitorsrv

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mascarenhasmelson/gomotz/utils"
)

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
func (p *PostgresDB) CreatePortMonitor(ctx context.Context, m *utils.PortMonitor) error {
	query := `
        INSERT INTO port_monitors 
        (friendly_name, hostname, port, heartbeat_interval, retries, heartbeat_retry_interval)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, status, created_at, updated_at
    `
	return p.pool.QueryRow(ctx, query,
		m.FriendlyName, m.Hostname, m.Port,
		m.HeartbeatInterval, m.Retries, m.HeartbeatRetryInterval,
	).Scan(&m.ID, &m.Status, &m.CreatedAt, &m.UpdatedAt)
}

func (p *PostgresDB) GetAllPortMonitors(ctx context.Context) ([]*utils.PortMonitor, error) {
	rows, err := p.pool.Query(ctx, `
        SELECT id, friendly_name, hostname, port, 
               heartbeat_interval, retries, heartbeat_retry_interval, 
               status, last_tcp_status,        -- ✅ must be here
               last_checked_at, last_response_ms,
               created_at, updated_at
        FROM port_monitors 
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []*utils.PortMonitor
	for rows.Next() {
		m := &utils.PortMonitor{}
		if err := rows.Scan(
			&m.ID, &m.FriendlyName, &m.Hostname, &m.Port,
			&m.HeartbeatInterval, &m.Retries, &m.HeartbeatRetryInterval,
			&m.Status, &m.LastTCPStatus, // ✅ must be here
			&m.LastCheckedAt, &m.LastResponseMs,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}
	return monitors, rows.Err()
}

func (p *PostgresDB) GetPortMonitorByID(ctx context.Context, id int) (*utils.PortMonitor, error) {
	m := &utils.PortMonitor{}
	err := p.pool.QueryRow(ctx, `
        SELECT id, friendly_name, hostname, port,
               heartbeat_interval, retries, heartbeat_retry_interval,
               status, last_tcp_status,       
               last_checked_at, last_response_ms,
               created_at, updated_at
        FROM port_monitors WHERE id = $1
    `, id).Scan(
		&m.ID, &m.FriendlyName, &m.Hostname, &m.Port,
		&m.HeartbeatInterval, &m.Retries, &m.HeartbeatRetryInterval,
		&m.Status, &m.LastTCPStatus,
		&m.LastCheckedAt, &m.LastResponseMs,
		&m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}

func (p *PostgresDB) UpdatePortMonitor(ctx context.Context, m *utils.PortMonitor) error {
	_, err := p.pool.Exec(ctx, `
        UPDATE port_monitors
        SET friendly_name = $1, hostname = $2, port = $3,
            heartbeat_interval = $4, retries = $5,
            heartbeat_retry_interval = $6, updated_at = NOW()
        WHERE id = $7
    `, m.FriendlyName, m.Hostname, m.Port,
		m.HeartbeatInterval, m.Retries, m.HeartbeatRetryInterval, m.ID)
	return err
}

func (p *PostgresDB) UpdatePortMonitorStatus(ctx context.Context, id int, status string, tcpStatus string, responseMs *int) error {
	_, err := p.pool.Exec(ctx, `
        UPDATE port_monitors
        SET status = $1,
            last_tcp_status = $2,
            last_checked_at = NOW(),
            last_response_ms = $3
        WHERE id = $4
    `, status, tcpStatus, responseMs, id)
	return err
}

func (p *PostgresDB) DeletePortMonitor(ctx context.Context, id int) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM port_monitors WHERE id = $1`, id)
	return err
}

func (p *PostgresDB) InsertPortMonitorLog(ctx context.Context, log *utils.PortMonitorLog) error {
	_, err := p.pool.Exec(ctx, `
        INSERT INTO port_monitor_logs (monitor_id, status, response_ms, error_message)
        VALUES ($1, $2, $3, $4)
    `, log.MonitorID, log.Status, log.ResponseMs, log.ErrorMessage)
	return err
}

func (p *PostgresDB) GetPortMonitorLogs(ctx context.Context, monitorID, limit int) ([]*utils.PortMonitorLog, error) {
	rows, err := p.pool.Query(ctx, `
        SELECT id, monitor_id, status, response_ms, error_message, checked_at
        FROM port_monitor_logs
        WHERE monitor_id = $1
        ORDER BY checked_at DESC LIMIT $2
    `, monitorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*utils.PortMonitorLog
	for rows.Next() {
		l := &utils.PortMonitorLog{}
		if err := rows.Scan(&l.ID, &l.MonitorID, &l.Status,
			&l.ResponseMs, &l.ErrorMessage, &l.CheckedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (p *PostgresDB) CreateSNMPMonitor(ctx context.Context, m *utils.SNMPMonitor) error {
	query := `
        INSERT INTO snmp_monitors
        (friendly_name, hostname, port, community_string, oid, snmp_version,
         polling_interval, timeout, retries, expected_value_type)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, status, created_at, updated_at
    `
	return p.pool.QueryRow(ctx, query,
		m.FriendlyName, m.Hostname, m.Port, m.CommunityString,
		m.OID, m.SNMPVersion, m.PollingInterval, m.Timeout,
		m.Retries, m.ExpectedValueType,
	).Scan(&m.ID, &m.Status, &m.CreatedAt, &m.UpdatedAt)
}

func (p *PostgresDB) GetAllSNMPMonitors(ctx context.Context) ([]*utils.SNMPMonitor, error) {
	rows, err := p.pool.Query(ctx, `
        SELECT id, friendly_name, hostname, port, community_string, oid,
               snmp_version, polling_interval, timeout, retries,
               expected_value_type, status, last_value, last_checked_at,
               last_response_ms, error_message, created_at, updated_at
        FROM snmp_monitors ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []*utils.SNMPMonitor
	for rows.Next() {
		m := &utils.SNMPMonitor{}
		if err := rows.Scan(
			&m.ID, &m.FriendlyName, &m.Hostname, &m.Port,
			&m.CommunityString, &m.OID, &m.SNMPVersion,
			&m.PollingInterval, &m.Timeout, &m.Retries,
			&m.ExpectedValueType, &m.Status, &m.LastValue,
			&m.LastCheckedAt, &m.LastResponseMs, &m.ErrorMessage,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}
	return monitors, rows.Err()
}

func (p *PostgresDB) GetSNMPMonitorByID(ctx context.Context, id int) (*utils.SNMPMonitor, error) {
	m := &utils.SNMPMonitor{}
	err := p.pool.QueryRow(ctx, `
        SELECT id, friendly_name, hostname, port, community_string, oid,
               snmp_version, polling_interval, timeout, retries,
               expected_value_type, status, last_value, last_checked_at,
               last_response_ms, error_message, created_at, updated_at
        FROM snmp_monitors WHERE id = $1
    `, id).Scan(
		&m.ID, &m.FriendlyName, &m.Hostname, &m.Port,
		&m.CommunityString, &m.OID, &m.SNMPVersion,
		&m.PollingInterval, &m.Timeout, &m.Retries,
		&m.ExpectedValueType, &m.Status, &m.LastValue,
		&m.LastCheckedAt, &m.LastResponseMs, &m.ErrorMessage,
		&m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}

func (p *PostgresDB) UpdateSNMPMonitor(ctx context.Context, m *utils.SNMPMonitor) error {
	_, err := p.pool.Exec(ctx, `
        UPDATE snmp_monitors
        SET friendly_name = $1, hostname = $2, port = $3,
            community_string = $4, oid = $5, snmp_version = $6,
            polling_interval = $7, timeout = $8, retries = $9,
            expected_value_type = $10, updated_at = NOW()
        WHERE id = $11
    `, m.FriendlyName, m.Hostname, m.Port, m.CommunityString,
		m.OID, m.SNMPVersion, m.PollingInterval, m.Timeout,
		m.Retries, m.ExpectedValueType, m.ID)
	return err
}

func (p *PostgresDB) UpdateSNMPMonitorStatus(ctx context.Context, id int, status, value string, responseMs *int, errMsg *string) error {
	_, err := p.pool.Exec(ctx, `
        UPDATE snmp_monitors
        SET status = $1, last_value = $2, last_checked_at = NOW(),
            last_response_ms = $3, error_message = $4
        WHERE id = $5
    `, status, value, responseMs, errMsg, id)
	return err
}

func (p *PostgresDB) DeleteSNMPMonitor(ctx context.Context, id int) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM snmp_monitors WHERE id = $1`, id)
	return err
}

func (p *PostgresDB) InsertSNMPMonitorLog(ctx context.Context, l *utils.SNMPMonitorLog) error {
	_, err := p.pool.Exec(ctx, `
        INSERT INTO snmp_monitor_logs (monitor_id, status, value, response_ms, error_message)
        VALUES ($1, $2, $3, $4, $5)
    `, l.MonitorID, l.Status, l.Value, l.ResponseMs, l.ErrorMessage)
	return err
}

func (p *PostgresDB) GetSNMPMonitorLogs(ctx context.Context, monitorID, limit int) ([]*utils.SNMPMonitorLog, error) {
	rows, err := p.pool.Query(ctx, `
        SELECT id, monitor_id, status, value, response_ms, error_message, checked_at
        FROM snmp_monitor_logs
        WHERE monitor_id = $1
        ORDER BY checked_at DESC LIMIT $2
    `, monitorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*utils.SNMPMonitorLog
	for rows.Next() {
		l := &utils.SNMPMonitorLog{}
		if err := rows.Scan(
			&l.ID, &l.MonitorID, &l.Status,
			&l.Value, &l.ResponseMs, &l.ErrorMessage, &l.CheckedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (p *PostgresDB) GetAllEnabledSNMPMonitors(ctx context.Context) ([]*utils.SNMPMonitor, error) {
	rows, err := p.pool.Query(ctx, `
        SELECT id, friendly_name, hostname, port, community_string, oid,
               snmp_version, polling_interval, timeout, retries,
               expected_value_type, status, last_value, last_checked_at,
               last_response_ms, error_message, created_at, updated_at
        FROM snmp_monitors ORDER BY id
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []*utils.SNMPMonitor
	for rows.Next() {
		m := &utils.SNMPMonitor{}
		if err := rows.Scan(
			&m.ID, &m.FriendlyName, &m.Hostname, &m.Port,
			&m.CommunityString, &m.OID, &m.SNMPVersion,
			&m.PollingInterval, &m.Timeout, &m.Retries,
			&m.ExpectedValueType, &m.Status, &m.LastValue,
			&m.LastCheckedAt, &m.LastResponseMs, &m.ErrorMessage,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}
	return monitors, rows.Err()
}
