# 🛠️ GoMotz Setup Guide

This guide walks you through installing GoMotz, configuring your network interfaces for monitoring, and setting up VLANs on your Raspberry Pi.

---

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

---

## Prerequisites

Before installing GoMotz, make sure you have:

- Raspberry Pi 4 (2GB RAM or more recommended)
- Raspberry Pi OS 64-bit (Bookworm or Bullseye)
- Docker and Docker Compose installed
- Your Pi connected to the network via Ethernet (`eth0`) or Wi-Fi (`wlan0`)

**Install Docker if you haven't already:**

```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker
```

---

## Installation

```bash
git clone https://github.com/mascarenhasmelson/gomotz.git
cd gomotz
docker-compose up -d
```

Once running, open your browser and navigate to:

```
http://<your-pi-ip>:8000
```

---

## First Run — Network Configuration

After launching GoMotz, the first thing to do is tell it which network interfaces to monitor. Navigate to **Settings → Network Configuration** from the sidebar.

---

### Step 1: Open Network Configuration

GoMotz will automatically detect all available network interfaces on your Raspberry Pi. You will see a list under **Network Interfaces** on the left panel.

Each interface shows:
- **Interface name** (e.g. `eth0`)
- **IP address** and **MAC address**
- **Monitor status** — `MONITORED` or `NOT MONITORED`

Click on any interface to view its full details on the right panel, including CIDR, default gateway, and current monitor status.

---

### Step 2: Create a Network to Monitor

If an interface shows **NOT MONITORED**, click it and you will see the **Create Network** form on the right.

Fill in the following:

| Field | Description | Example |
|-------|-------------|---------|
| **Network Name** | A friendly label for this network | `wifi`, `lan`, `guest` |
| **Scan Interval** | How often GoMotz scans this network (10–3600 seconds) | `30` |
| **Enable monitoring** | Toggle to activate monitoring immediately | ✅ Checked |

Click **Create Network** to save. GoMotz will begin scanning the interface at the interval you set.

> 💡 **Tip:** Use descriptive names like `home-lan`, `iot-vlan`, or `office-wifi` — especially when monitoring multiple interfaces.

---

### Step 3: Verify Monitoring is Active

After creating the network, the interface card on the left will update to show **MONITORED** with a green dot. The right panel will display:

- Interface name, IP, MAC, CIDR, and gateway
- Network name you assigned
- IP range being scanned
- **Monitoring Active** badge
- Options to **Stop Monitoring** or **Delete Network**

Your devices on that network will now begin appearing in the **Device Monitoring** section.

---

## VLAN Setup on Raspberry Pi

If you want GoMotz to monitor traffic across multiple VLANs, you first need to configure VLAN sub-interfaces on the Raspberry Pi itself. Here's a complete step-by-step guide.

---

### Step 1: Install VLAN Package

```bash
sudo apt-get update
sudo apt-get install vlan -y

# Load the 8021q kernel module on boot
sudo su -c "echo '8021q' >> /etc/modules"
```

---

### Step 2: Create VLAN Interfaces

Create the file `/etc/network/interfaces.d/vlans` and define your VLAN sub-interfaces. The naming convention is `<physical-interface>.<VLAN-ID>`.

```bash
sudo nano /etc/network/interfaces.d/vlans
```

Add the following (adjust VLAN IDs to match your network):

```
# VLAN 10
auto eth0.10
iface eth0.10 inet manual
  vlan-raw-device eth0

# VLAN 20
auto eth0.20
iface eth0.20 inet manual
  vlan-raw-device eth0
```

> You can add as many VLANs as needed by repeating this block with different IDs.

---

### Step 3: Assign Static IPs

Edit `/etc/dhcpcd.conf` to assign static IP addresses to each VLAN interface:

```bash
sudo nano /etc/dhcpcd.conf
```

Add at the end of the file:

```bash
# Main interface
interface eth0
static ip_address=10.0.20.125/24
static routers=10.0.20.1
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

---

### Step 4: Reboot and Verify

```bash
sudo reboot
```

After rebooting, verify the VLAN interfaces are up:

```bash
sudo ifconfig
```

You should see output similar to:

```
eth0.10: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 10.0.10.125  netmask 255.255.255.0  broadcast 10.0.10.255
        ether b8:27:eb:10:70:13  txqueuelen 1000  (Ethernet)

eth0.20: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 10.0.20.125  netmask 255.255.255.0  broadcast 10.0.20.255
        ether b8:27:eb:10:70:13  txqueuelen 1000  (Ethernet)
```

---

### Step 5: Add VLAN Interfaces to GoMotz

Once your VLAN interfaces are up, go back to GoMotz **Settings → Network Configuration**. The new VLAN interfaces (e.g. `eth0.10`, `eth0.20`) will appear automatically in the interface list.

Click each one and follow [Step 2](#step-2-create-a-network-to-monitor) to create a named network and enable monitoring.

> 💡 **Tip:** Name your VLAN networks clearly — e.g. `iot-vlan10`, `cameras-vlan20` — so they are easy to identify in the Device Monitoring view.

---

## Troubleshooting

**Interface not appearing in GoMotz?**
- Make sure the interface is UP: `ip link show`
- Restart GoMotz: `docker-compose restart`

**VLAN interface not coming up after reboot?**
- Confirm `8021q` is in `/etc/modules`
- Run `sudo modprobe 8021q` manually to load without rebooting
- Check `/etc/network/interfaces.d/vlans` for typos

**Devices not appearing after enabling monitoring?**
- Check that GoMotz has the necessary permissions to run ARP scans
- Verify the scan interval isn't set too high
- Ensure GoMotz container is running with `docker ps`

---

<div align="center">

[⬅️ Back to README](../README.md) · [🗺️ View Roadmap](ROADMAP.md)

</div>