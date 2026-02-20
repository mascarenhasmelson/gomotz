package bgservices

import (
	"time"

	"github.com/mascarenhasmelson/gomotz/utils"
)

func TcpCheck(req utils.TCPCheckRequest) utils.TCPCheckResponse {
	if req.Host == "" {
		return utils.TCPCheckResponse{
			Success: false,
			Status:  "error",
			Message: "Host cannot be empty",
		}
	}
	if req.Port < 1 || req.Port > 65535 {
		return utils.TCPCheckResponse{
			Success: false,
			Status:  "error",
			Message: "Invalid port number (must be 1-65535)",
		}
	}
	timeoutSec := req.Timeout
	if timeoutSec <= 0 {
		timeoutSec = 3
	}
	timeout := time.Duration(timeoutSec) * time.Second
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
	if elapsed >= timeout.Milliseconds() {
		return utils.TCPCheckResponse{
			Success:      true,
			Status:       "filtered",
			Host:         req.Host,
			Port:         req.Port,
			ResponseTime: elapsed,
			Message:      "Port may be filtered by firewall (connection timeout)",
		}
	}
	return utils.TCPCheckResponse{
		Success:      true,
		Status:       "closed",
		Host:         req.Host,
		Port:         req.Port,
		ResponseTime: elapsed,
		Message:      "Port is closed (connection refused)",
	}
}
