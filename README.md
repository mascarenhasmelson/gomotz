<div align="center">

<img src="https://img.shields.io/badge/version-0.5.5--beta-blueviolet?style=for-the-badge" />
<img src="https://img.shields.io/badge/platform-Raspberry%20Pi-c51a4a?style=for-the-badge&logo=raspberry-pi" />
<img src="https://img.shields.io/badge/arm64-blue?style=for-the-badge&logo=arm" />
<img src="https://img.shields.io/badge/license-MIT-green?style=for-the-badge" />

# GoMotz

### Open-Source Network Management Tool

**Self-hosted, simple network monitoring system built for Raspberry Pi.**

*Monitor, control, and secure your entire network from a single dashboard.*

---

</div>

## What is GoMotz?

**GoMotz** is a free, open-source network monitoring and management system designed to run on a **Raspberry Pi**. It gives you full visibility and control over your local network, from device discovery to port forwarding and DNS lookups to service checks, all from a real-time web dashboard.

Whether you're a homelab enthusiast or a network engineer, GoMotz puts network monitoring in your hands, **for free**.

---

## Features

### Device Monitoring

Scan and track every device on your network.

- **Network Device Discovery**: Automatically detect all connected devices
- **IP & MAC Address Tracking**: Full inventory with hostname and vendor info
- **Status Filtering**: Instantly view Online, Offline, and Conflict devices
- **Open Port Discovery**: See which ports are active on each device
- **Search & Filter**: Find devices by IP, MAC, hostname, or vendor name
- **Domain Expiry Monitor**: Stay ahead of expiring domain registrations
- **TCP Port Scanning**
- **Port Monitoring**

### Port Forwarding & Service Exposure

Expose your internal services for remote access using port forwarding with Tailscale.

> **Note:** GoMotz does not yet include a built-in reverse proxy. Remote access is currently handled through Tailscale. To access your LAN services remotely, install Tailscale on your Raspberry Pi and connect your devices to the same Tailscale network.
>
> Tailscale is a temporary solution. Once the GoMotz VPN-based remote access tool is completed, it will replace Tailscale for remote access.

When adding a port forwarding rule, the local IP can be set in one of two ways:

| Local IP | Result |
|---|---|
| **Tailscale IP** (e.g. `100.x.x.x`) | Binds the rule to your Tailscale interface, so it is accessible only over your Tailscale network |
| **`0.0.0.0`** | Binds to all interfaces, so it is accessible from both your local network and Tailscale |

This gives you flexible, secure control over which services are exposed and how they are reached remotely.

- **Port Forwarding Rules**: Translate internal IP:Port combinations to custom external ports
- **Tailscale Interface Binding**: Bind rules to your Tailscale IP for secure, private remote access
- **Multiple Device IP/Port Mapping**: Manage complex multi-service environments with ease
- **Real-Time Service Monitoring**: Track the live status of all forwarded services

### Network Tools

A full suite of diagnostic utilities built right in.

| Tool | Description | Status |
|---|---|---|
| ~~Port Scan~~ | ~~Scan any host for open TCP ports~~ | Completed |
| ~~TCP Check~~ | ~~Verify TCP connectivity to any IP and port~~ | Completed |
| ~~DNS Lookup~~ | ~~Resolve domains with support for all record types (A, AAAA, MX, CNAME, TXT, SOA, NS, SRV, PTR, and more)~~ | Completed |
| ~~Traceroute~~ | ~~Visualize the network path to any destination~~ | Completed |
| ~~Ping Monitor~~ | ~~Continuous latency monitoring for any host~~ | Completed |
| ~~HTTP(S) Check~~ | ~~Validate HTTP/HTTPS endpoint availability and response~~ | Completed |
| ~~Speed Test~~ | ~~Measure your real-time internet bandwidth~~ | Completed |

---

## Roadmap & Upcoming Features

- [ ] Android app: remote access and monitoring *(not completed)*
- [ ] Remote access through client software for Windows, Linux, and Android *(in progress)*
- [ ] User role-based access control (RBAC)
- [ ] Auto VLAN detection
- [ ] L2/L3 topology mapping
- [ ] Multiple location agent access through unified device access in a single cloud dashboard
- [ ] Device asset inventory
- [ ] Wake-on-LAN (WoL)

---

## Frequently Asked Questions

### What devices can GoMotz monitor?
GoMotz can discover and monitor any IP-enabled device on your network, including routers, switches, firewalls, access points, servers, workstations, cameras, printers, UPS systems, smart plugs, PDUs, POS terminals, digital signage, and IoT devices.

### Is GoMotz cloud-based?
No. GoMotz is self-hosted. A lightweight agent runs on your Raspberry Pi, and the real-time dashboard is served from your own network. Remote access is currently provided through Tailscale as a temporary solution, until the GoMotz VPN-based remote access tool is completed. A unified cloud dashboard is on the roadmap.

### What is GoMotz used for?
GoMotz is used to discover, monitor, and remotely manage network-connected devices. Common use cases include network monitoring, device inventory, remote troubleshooting, and network security monitoring for homelabs and small networks.

**[Read the Setup Guide →](docs/SETUP.md)**

---

## Contributing

GoMotz is open source and contributions are welcome! Whether it's bug reports, feature suggestions, documentation, or code, all help is appreciated.

---

## License

GoMotz is released under the [MIT License](LICENSE).
