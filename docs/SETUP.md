# GoMotz — Installation and Setup Guide

## Access the Services

| Service | URL / Address | Description |
|---|---|---|
| Frontend | [http://localhost:8000](http://localhost:8000) | GoMotz Web GUI |
| Backend | [http://localhost:8082](http://localhost:8082) | Go Server |
| Database | `localhost:5432` | PostgreSQL |

### Default Runtime Environment

The `.env` file contains the following default parameters:

```ini
POSTGRES_USER=admin
POSTGRES_PASSWORD=StrongPassword123
POSTGRES_DB=tunnel_services
POSTGRES_PORT=5432
VITE_API_URL=http://localhost:8082
```

> Change default credentials before using GoMotz in a production or exposed environment.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [First Run — Network Configuration](#first-run--network-configuration)
  - [Step 1: Open Network Configuration](#step-1-open-network-configuration)
  - [Step 2: Create a Network to Monitor](#step-2-create-a-network-to-monitor)
  - [Step 3: Verify Monitoring is Active](#step-3-verify-monitoring-is-active)
- [VLAN Setup on Raspberry Pi](#vlan-setup-on-raspberry-pi)
  - [Step 1: Install VLAN Package](#step-1-install-vlan-package)
  - [Step 2: Create VLAN Interfaces](#step-2-create-vlan-interfaces)
  - [Step 3: Assign Static IPs](#step-3-assign-static-ips)
  - [Step 4: Reboot and Verify](#step-4-reboot-and-verify)
  - [Step 5: Add VLAN Interfaces to GoMotz](#step-5-add-vlan-interfaces-to-gomotz)
- [Troubleshooting](#troubleshooting)

## Prerequisites

Before installing GoMotz, make sure you have:

- Raspberry Pi 4 (2 GB RAM or more recommended)
- Raspberry Pi OS 64-bit (Bookworm or Bullseye)
- Docker and Docker Compose installed
- Raspberry Pi connected to the network via Ethernet (`eth0`) or Wi-Fi (`wlan0`)

### Install Docker

If Docker is not installed:

```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker "$USER"
newgrp docker
```

## Installation

Clone the repository and start the services:

```bash
git clone https://github.com/mascarenhasmelson/gomotz.git
cd gomotz
docker-compose up -d
```

Once the containers are running, open the frontend:

```text
http://<your-pi-ip>:8000
```

## First Run — Network Configuration

After launching GoMotz, configure which network interfaces it should monitor. Navigate to **Settings → Network Configuration** in the sidebar.

### Step 1: Open Network Configuration

GoMotz automatically detects available network interfaces on your Raspberry Pi. The **Network Interfaces** list appears in the left panel.

Each interface displays:

- **Interface name** (for example, `eth0`)
- **IP address** and **MAC address**
- **Monitor status** — `MONITORED` or `NOT MONITORED`

Select an interface to view its details in the right panel, including CIDR, default gateway, and monitor status.

### Step 2: Create a Network to Monitor

If an interface is marked **NOT MONITORED**, select it to open the **Create Network** form.

| Field | Description | Example |
|---|---|---|
| **Network Name** | Friendly label for the network | `wifi`, `lan`, `guest` |
| **Scan Interval** | How often GoMotz scans this network (10–3600 seconds) | `30` |
| **Enable monitoring** | Enable monitoring immediately | Checked |

Click **Create Network** to save. GoMotz will begin scanning the interface at the configured interval.

Tip: Use descriptive names such as `home-lan`, `iot-vlan`, or `office-wifi`, especially when monitoring multiple interfaces.

### Step 3: Verify Monitoring is Active

After creating the network, the interface card should show **MONITORED**. The right panel will display:

- Interface name, IP, MAC, CIDR, and gateway
- The network name you assigned
- IP range being scanned
- Monitoring Active status
- Options to **Stop Monitoring** or **Delete Network**

Devices on that network should begin appearing in the **Device Monitoring** section.

## VLAN Setup on Raspberry Pi

To monitor multiple VLANs, configure VLAN sub-interfaces on the Raspberry Pi first. These instructions use `eth0` as the physical interface and VLAN IDs 10 and 20 as examples. Adapt the interface names, VLAN IDs, IP addresses, and gateway addresses to your network.

### Step 1: Install VLAN Package

```bash
sudo apt-get update
sudo apt-get install vlan -y

# Load the 802.1Q kernel module at boot
echo 8021q | sudo tee -a /etc/modules
```

### Step 2: Create VLAN Interfaces

Create `/etc/network/interfaces.d/vlans`:

```bash
sudo nano /etc/network/interfaces.d/vlans
```

Add the following configuration, adjusting VLAN IDs as required:

```text
# VLAN 10
auto eth0.10
iface eth0.10 inet manual
    vlan-raw-device eth0

# VLAN 20
auto eth0.20
iface eth0.20 inet manual
    vlan-raw-device eth0
```

Add more VLANs by repeating the blocks with different VLAN IDs.

> Note: This configuration uses the `ifupdown` networking system. Raspberry Pi OS installations using NetworkManager or another network manager may require a different configuration method.

### Step 3: Assign Static IPs

If your system uses `dhcpcd`, edit `/etc/dhcpcd.conf`:

```bash
sudo nano /etc/dhcpcd.conf
```

Example configuration (replace the addresses with values appropriate for your network):

```text
# Main interface
interface eth0
static ip_address=10.0.30.125/24
static routers=10.0.30.1
static domain_name_servers=1.1.1.1

# VLAN 10
interface eth0.10
static ip_address=10.0.10.125/24
static routers=10.0.10.1
static domain_name_servers=1.1.1.1

# VLAN 20
interface eth0.20
static ip_address=10.0.20.125/24
static routers=10.0.20.1
static domain_name_servers=1.1.1.1
```

Ensure each interface has an IP address appropriate for its VLAN and that the addresses do not conflict with other devices. The example uses different subnets for the physical interface and VLAN interfaces.

### Step 4: Reboot and Verify

Reboot the Raspberry Pi:

```bash
sudo reboot
```

After reboot, verify the VLAN interfaces:

```bash
ip -brief address
```

You can also use:

```bash
ip link show
```

Example output (addresses and MAC values will vary):

```text
eth0.10: <UP,BROADCAST,RUNNING,MULTICAST> mtu 1500
    inet 10.0.10.125/24

eth0.20: <UP,BROADCAST,RUNNING,MULTICAST> mtu 1500
    inet 10.0.20.125/24
```

### Step 5: Add VLAN Interfaces to GoMotz

Once the VLAN interfaces are up, return to **Settings → Network Configuration** in GoMotz. The VLAN interfaces (such as `eth0.10` and `eth0.20`) should appear in the interface list.

Select each interface and follow [Step 2: Create a Network to Monitor](#step-2-create-a-network-to-monitor) to create a named network and enable monitoring.

Use clear names such as `iot-vlan10` or `cameras-vlan20` to identify networks easily in Device Monitoring.

## Troubleshooting

### Interface not appearing in GoMotz

- Confirm that the interface is up: `ip link show`
- Restart GoMotz: `docker-compose restart`

### VLAN interface not coming up after reboot

- Confirm that `8021q` is listed in `/etc/modules`.
- Load the module manually: `sudo modprobe 8021q`
- Check `/etc/network/interfaces.d/vlans` for errors.
- Confirm that your Raspberry Pi OS networking system supports the configuration method used in this guide.

### Devices not appearing after enabling monitoring

- Check that GoMotz has the permissions required to perform ARP scans.
- Verify that the scan interval is not set too high.
- Ensure the GoMotz containers are running: `docker ps`
- Confirm the Raspberry Pi is connected to the VLANs and that the switch port is configured to carry the required VLAN tags.

---

[Back to README](../README.md) · [View Roadmap](ROADMAP.md)
