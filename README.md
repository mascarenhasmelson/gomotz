<div align="center">

<img src="https://img.shields.io/badge/version-0.5.5--beta-blueviolet?style=for-the-badge" />
<img src="https://img.shields.io/badge/platform-Raspberry%20Pi-c51a4a?style=for-the-badge&logo=raspberry-pi" />
<img src="https://img.shields.io/badge/arm64-blue?style=for-the-badge&logo=arm" />
<img src="https://img.shields.io/badge/license-MIT-green?style=for-the-badge" />

# GoMotz

### Open-Source Network Management Tool

**Self-hosted, Simple Network Monitoring System built for Raspberry Pi.**

*Monitor, control, and secure your entire network from a single dashboard.*

---

</div>

## What is GoMotz?

**GoMotz** is a free, open-source network monitoring and management system designed to run on a **Raspberry Pi**. It gives you full visibility and control over your local network from device discovery to port forwarding, DNS lookups to service checks all from an elegant real-time web dashboard.

Whether you're a homelab enthusiast or network engineer GoMotz puts network monitoring in your hands, **for free**.

## Features

Gomotz — Device Monitoring & Network Management
Scan and track every device on your network.

Network Device Discovery
Automatically detect all connected devices

IP & MAC address tracking with a full inventory including hostname and vendor info

Status filtering — instantly view Online, Offline, and Conflict devices

Open port discovery — see which ports are active on each device

Search & filter by IP, MAC, hostname, or vendor name

Domain expiry monitor — stay ahead of expiring domain registrations

TCP port scanning

Port monitoring

Port Forwarding & Service Exposure
Expose your internal services for remote access using port forwarding with Tailscale.

⚠️ Note: Gomotz does not yet include a built-in reverse proxy. Remote access is currently handled through Tailscale. To access your LAN services remotely, install Tailscale on your Raspberry Pi and connect your devices to the same Tailscale network.

When adding a port forwarding rule, the local IP can be set in one of two ways:

Tailscale IP (e.g. 100.x.x.x) — binds the rule specifically to your Tailscale interface, making it accessible only over your Tailscale network

0.0.0.0 — binds to all interfaces, making it accessible from both your local network and Tailscale

This gives you flexible, secure control over which services are exposed and how they are reached remotely.

Features:

Port forwarding rules — translate internal IP:Port combinations to custom external ports

Tailscale interface binding — bind rules to your Tailscale IP for secure, private remote access

Multiple device IP/port mapping — manage complex multi-service environments with ease

Real-time service monitoring — track the live status of all forwarded services

Network Tools
A full suite of diagnostic utilities built right in.

Tool	Description
Port Scan	Scan any host for open TCP ports
TCP Check	Verify TCP connectivity to any IP and port
DNS Lookup	Resolve domains with support for all record types (A, AAAA, MX, CNAME, TXT, SOA, NS, SRV, PTR, and more)
Traceroute	Visualize the network path to any destination
Ping Monitor	Continuous latency monitoring for any host
HTTP(S) Check	Validate HTTP/HTTPS endpoint availability and response
Speed Test	Measure your real-time internet bandwidth
Roadmap & Upcoming Features
~~TCP Port Monitoring~~ (removed)

~~HTTP(S) Monitor — uptime tracking with SSL certificate monitoring~~ (removed)

~~Ping Monitor~~ (removed)

~~SNMP Monitor~~ (removed)

Android app — remote access and monitoring (not completed)

~~Network Device Tracking~~ (removed)

Remote access through client software for Windows, Linux, Android, and iPhone (currently working)

User role-based access control (RBAC)

Auto VLAN detection

L2/L3 topology mapping

Unified device access through a cloud single dashboard

Device asset inventory

Wake-on-LAN (WoL)

Frequently Asked Questions
What devices can Gomotz monitor?
Gomotz monitors any IP-enabled device, including routers, switches, firewalls, access points, servers, workstations, cameras, printers, UPS systems, smart plugs, PDUs, POS terminals, and IoT devices. It supports SNMP monitoring, TCP service checks, and over 1,000 hardware and software integrations.

Is Gomotz cloud-based?
Gomotz is a cloud-based SaaS platform. A lightweight agent is installed on the local network on a Raspberry Pi, and all monitoring data is accessible via the Gomotz cloud dashboard and mobile app from anywhere.

What is Gomotz used for?
Gomotz is used to discover, monitor, and remotely manage network-connected devices across one or many sites. Primary use cases include proactive network monitoring, IT asset inventory, remote troubleshooting, network security monitoring, and managed service delivery for MSPs and IT teams.

Is Gomotz good for MSPs?
Yes. Gomotz is purpose-built for MSPs managing multiple client networks. It offers multi-tenant dashboards, per-client network views, PSA and documentation integrations, and per-collector licensing that scales without high overhead. MSPs use Gomotz to monitor all client networks from a single interface.

How long does it take to set up Gomotz?
Most networks can be discovered and monitored within 15 minutes of installing a Gomotz agent. Larger or more complex networks may take longer to map completely.

What can Gomotz monitor?
Gomotz monitors any IP-enabled device on a network, including routers, switches, firewalls, access points, servers, workstations, printers, cameras, UPS systems, smart plugs, PDUs, POS systems, digital signage, and IoT devices. It supports SNMP, TCP service monitoring, NetFlow, and device-specific integrations for vendor-native visibility.

**[Read the Setup Guide →](docs/SETUP.md)**

---



## Contributing

GoMotz is open source and contributions are welcome! Whether it's bug reports, feature suggestions, documentation, or code — all help is appreciated.

---

## License

GoMotz is released under the [MIT License](LICENSE).

---

