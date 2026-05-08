# GoMotz Roadmap

This document tracks the planned features and future direction of GoMotz. Items are grouped by theme and roughly ordered by priority.

> 💡 Have an idea or want to contribute to a roadmap item? [Open an issue](https://github.com/mascarenhasmelson/gomotz/issues) or submit a PR!

---

## UI & Backend fix

---
## Network Devices Restart Tracking
Monitor and log devices reboots and unexpected restarts across your network. Get notified when a device goes offline and comes back up, with a full restart history per device.

---

## GoMotz VPN — Custom Peer-to-Peer Remote Access
A self-hosted remote access solution alternative to Domotz's OpenVPN integration. Provides encrypted secure tunnels supporting HTTP/HTTPS, SSH, RDP, Telnet, and TCP connections. Also includes VPN on Demand for full LAN access when you need to reach anything on your network remotely. Inspired by Tailscale, with support for both **Windows and Linux** platforms.

---

## Alerts — NTFY, Slack & Telegram Integration
Push notifications and alerts delivered directly to your preferred platform. Get notified about device status changes, service downtime, high latency, or any monitored event — through NTFY, Slack, or Telegram.

---

## Internet Block — Per-Device Access Control
Block internet access for individual devices on your network directly from the dashboard. Useful for parental controls,device isolation, or restricting guest devices without touching your router configuration.

---

##  LAN Wakeup (WoL) — Remote Wake-on-LAN
Remotely wake any device on your local network using Wake-on-LAN magic packets. Send WoL commands directly from the GoMotz dashboard without needing to be physically present on the network.

---

## Authentication & Login — Security
Secure your GoMotz instance with user authentication. Protect the dashboard behind a login screen so only authorized users can access your network monitoring and management tools.

---

##  Auto VLAN Detection & Configuration
Automatically discover existing VLANs on your network and configure VLAN sub-interfaces with minimal manual setup. GoMotz will detect tagged traffic, identify VLAN segments, and guide you through adding them to your monitoring setup.

---

##  Network Topology Mapping
Real-time, interactive network topology visualization. Automatically map how devices connect to each other using SNMP and ARP data, and display the result as a live, clickable graph  giving you a clear picture of your entire network at a glance.

---

## Reverse Proxy — Cloud Exposure via NAT
A built-in reverse proxy that allows GoMotz — even when running behind NAT — to securely expose internal services to the internet.

---
## Android App — Remote Access & Monitoring
 monitoring and managing your network on the go. Features include real-time network status, device alerts and push notifications control of your GoMotz instance from anywhere in the world.

---

##  Device Asset Inventory
A structured inventory of all devices discovered on your network. Track device details such as hostname, vendor, MAC address, first seen, last seen, open ports, and custom notes — giving you a complete and searchable register of everything on your network.

---
<div align="center">

*Roadmap is subject to change as the project evolves.*
*Community feedback and contributions directly shape what gets built next.*

[Back to README](../README.md)

</div>