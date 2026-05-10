package vlan

import (
	"context"
	"time"

	"github.com/mascarenhasmelson/gomotz/utils"
)

type VLANConfig struct {
	VLANId       int
	VLANName     string
	CIDR         string
	Interface    string
	Gateway      string
	ScanInterval time.Duration
}

type Database interface {
	UpsertDevice(ctx context.Context, device *utils.DiscoveredDevice) error
	GetEnabledVLANs(ctx context.Context) ([]*utils.VLANNetwork, error)
	GetVendorByOUI(ctx context.Context, oui string) (*utils.MACVendor, error)
	UpsertVendor(ctx context.Context, vendor *utils.MACVendor) error
	UpdateVendorLastSeen(ctx context.Context, oui string) error
}

type NotificationHandler interface {
	HandleNotification(notification *utils.DeviceNotification)
}
