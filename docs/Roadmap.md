# 🗺️ GoMotz Roadmap

This document tracks the planned features and future direction of GoMotz. Items are grouped by theme and roughly ordered by priority.

> 💡 Have an idea or want to contribute to a roadmap item? [Open an issue](https://github.com/yourusername/gomotz/issues) or submit a PR!

---

## 🔐 GoMotz VPN — Custom Peer-to-Peer Remote Access

**Inspired by Tailscale. Fully self-hosted. Zero third-party infrastructure.**

The goal is to build a lightweight,  VPN layer directly into GoMotz — letting you securely reach your home or office network from anywhere in the world without relying on any external service.

**Planned capabilities:**
- Peer-to-peer encrypted tunnels between the GoMotz agent and remote clients
- NAT traversal for connectivity behind firewalls and CGNAT
- Self-hosted coordination server — no Tailscale, no cloud middleman
- Automatic key exchange and peer registration via the GoMotz dashboard
- Per-device access control and revocation

---

## 📱 Android App — Remote Access & Command Control

A native Android application that puts your entire GoMotz dashboard in your pocket.

**Planned capabilities:**
- Real-time network status and alerts on your phone
- Remote device management — ping, wake, scan from anywhere
- Service monitoring notifications (push alerts for downtime)
- Port forwarding rule management on the go
- Secure connection via GoMotz VPN integration
- Command & control interface for advanced operations

---

## 🔎 L2 Discovery — UniPI Mode *(WinBox-style Zero-Config Detection)*

**Find and control Raspberry Pi devices on your network — even before they're configured.**

Inspired by MikroTik's WinBox Layer 2 neighbor discovery, UniPI mode will let you detect any GoMotz-capable Raspberry Pi on the same physical network segment — no IP address, no configuration required.

**Planned capabilities:**
- Broadcast-based Layer 2 discovery of unconfigured Pi devices
- One-click initial setup and onboarding directly from the dashboard
- Display device info: hostname, MAC address, hardware model, firmware version
- Remotely configure network settings on a newly flashed Pi
- Works even when the target device has no IP address assigned

---

## 🏷️ VLAN Detection & Auto-Configuration

Automatic discovery and management of VLANs on your network.

**Planned capabilities:**
- Passive VLAN detection via SNMP and 802.1Q tag inspection
- Auto-discovery of existing VLAN segments and their members
- UI-driven VLAN creation, tagging, and management
- Per-VLAN device grouping in the device monitoring view
- VLAN-aware port forwarding rules

---

## 🗺️ Network Topology Mapping

A live, visual representation of how every device on your network connects.

**Planned capabilities:**
- Auto-generated network topology diagrams using SNMP polling
- Interactive graph view — click any node to inspect device details
- Layer 2 and Layer 3 topology views
- Switch port mapping — see which device is on which port
- Topology change detection and alerting
- Export topology as image or JSON

---

## 🔄 Reverse Proxy Enhancements

Expanding the built-in reverse proxy with more advanced routing features.

**Planned capabilities:**
- Domain-based virtual hosting (route by hostname)
- SSL/TLS termination with automatic Let's Encrypt certificate management
- Basic auth and token-based access control per proxy rule
- Request/response header manipulation
- Integration with GoMotz VPN for private-only proxy exposure

---

## 📅 Release Timeline

| Phase | Focus | Status |
|-------|-------|--------|
| **v0.1.0** | Core dashboard, device monitoring, network tools, service monitoring, port forwarding | 🚧 In Progress |
| **v0.2.0** | Android app (basic remote access), GoMotz VPN (alpha) | 📋 Planned |
| **v0.3.0** | UniPI L2 Discovery, VLAN detection | 📋 Planned |
| **v0.4.0** | Network topology mapping, SNMP enhancements | 📋 Planned |
| **v1.0.0** | Full feature set, stable API, production-ready | 🎯 Goal |

---

<div align="center">

*Roadmap is subject to change as the project evolves.*
*Community feedback and contributions directly shape what gets built next.*

[⬅️ Back to README](../README.md)

</div>