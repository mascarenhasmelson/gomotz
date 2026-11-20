# GoMotz Roadmap

This document tracks planned features and the future direction of GoMotz. Items are grouped by theme and roughly ordered by priority.

Have an idea or want to contribute to a roadmap item? [Open an issue](https://github.com/mascarenhasmelson/gomotz/issues) or submit a pull request.

---

## UI and Backend Fixes

Planned improvements, bug fixes, and refinements across the user interface and backend.

---

## Network Device Restart Tracking

Monitor and log device reboots and unexpected restarts across your network. Get notified when a device goes offline and comes back online, with a full restart history for each device.

---

## GoMotz VPN — Custom Peer-to-Peer Remote Access

A self-hosted remote access solution as an alternative to Domotz's OpenVPN integration. Provides encrypted tunnels supporting HTTP/HTTPS, SSH, RDP, Telnet, and TCP connections.

Also includes VPN on Demand for full LAN access when you need to reach devices on your network remotely. Inspired by Tailscale, with support for both Windows and Linux platforms.

---

## Alerts — NTFY, Slack, and Telegram Integration

Push notifications and alerts delivered directly to your preferred platform. Get notified about device status changes, service downtime, high latency, or other monitored events through NTFY, Slack, or Telegram.

---

## Internet Access Control — Per-Device Restrictions

Block internet access for individual devices on your network directly from the dashboard. Useful for parental controls, device isolation, or restricting guest devices without changing your router configuration.

---

## LAN Wakeup — Remote Wake-on-LAN

Remotely wake devices on your local network using Wake-on-LAN (WoL) magic packets. Send WoL commands directly from the GoMotz dashboard without needing to be physically present on the network.

---

## Authentication and Login Security

Secure your GoMotz instance with user authentication. Protect the dashboard behind a login screen so only authorized users can access network monitoring and management tools.

---

## Automatic VLAN Detection and Configuration

Automatically discover existing VLANs on your network and configure VLAN sub-interfaces with minimal manual setup. GoMotz will detect tagged traffic, identify VLAN segments, and guide you through adding them to your monitoring setup.

---

## Network Topology Mapping

Real-time, interactive network topology visualization. Automatically map device connections using SNMP and ARP data, and display the results as a live, clickable graph, giving you a clear view of your network.

---

## Reverse Proxy — Internet Access Through NAT

A built-in reverse proxy that allows GoMotz, even when running behind NAT, to securely expose internal services to the internet.

---

## Android App — Remote Access and Monitoring

Monitor and manage your network on the go. Planned features include real-time network status, device alerts, push notifications, and remote control of your GoMotz instance from anywhere.

---

## Device Asset Inventory

A structured inventory of devices discovered on your network. Track details such as hostname, vendor, MAC address, first seen, last seen, open ports, device image, and custom notes, providing a complete and searchable register of devices on your network.

---
