package bgservices

import (
	"time"

	"github.com/mascarenhasmelson/gomotz/utils"
)

func TcpCheck(req utils.TCPCheckRequest) utils.TCPCheckResponse {
	timeout := time.Duration(req.Timeout) * time.Second
	if timeout == 0 {
		timeout = 3 * time.Second
	}
	start := time.Now()
	open := CheckPort(req.Host, req.Port)
	elapsed := time.Since(start).Milliseconds()
	if open {
		return utils.TCPCheckResponse{
			Success:      true,
			Status:       "open",
			Host:         req.Host,
			Port:         req.Port,
			ResponseTime: elapsed,
			Message:      "Port is open and accepting connections",
		}
	}

	status := "closed"
	message := "Port is closed or not accepting connections"

	if elapsed >= timeout.Milliseconds() {
		status = "filtered"
		message = "No response received - port may be filtered by firewall"
	}

	return utils.TCPCheckResponse{
		Success:      true,
		Status:       status,
		Host:         req.Host,
		Port:         req.Port,
		ResponseTime: elapsed,
		Message:      message,
	}
}
