<p align="center">
  <img src="images/gomotz.png" alt="GoMotz Logo" width="300">
</p>

<div align="center">

<img src="https://img.shields.io/badge/version-0.1.0--beta-blueviolet?style=for-the-badge" />
<img src="https://img.shields.io/badge/platform-Raspberry%20Pi-c51a4a?style=for-the-badge&logo=raspberry-pi" />
<img src="https://img.shields.io/badge/license-MIT-green?style=for-the-badge" />
<img src="https://img.shields.io/badge/status-Coming%20Soon-orange?style=for-the-badge" />

# GoMotz

### Open-Source Network Monitoring & Management Platform

**A powerful, self-hosted alternative to Domotz — built for Raspberry Pi.**

*Monitor, control, and secure your entire network from a single beautiful dashboard.*

---

</div>

## 🚀 What is GoMotz?

GoMotz is a free, open-source network monitoring and management system designed to run on a **Raspberry Pi**. It gives you full visibility and control over your local network — from device discovery to port forwarding, DNS lookups to service health checks — all from an elegant real-time web dashboard.

Whether you're a homelab enthusiast, a small business owner, or a network engineer, GoMotz puts enterprise-grade network monitoring in your hands, **for free**.

---

## 🏗️ Architecture

GoMotz follows a lightweight, modular architecture optimized to run efficiently on Raspberry Pi hardware.

```
┌─────────────────────────────────────────────────────────────┐
│                        GoMotz Agent                         │
│                    (Raspberry Pi Device)                     │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌────────────────────┐  │
│  │   Network   │  │   Service   │  │   Port Forwarding  │  │
│  │  Discovery  │  │  Monitors   │  │  & Tailscale Bind  │  │
│  └──────┬──────┘  └──────┬──────┘  └─────────┬──────────┘  │
│         │                │                   │              │
│  ┌──────▼────────────────▼───────────────────▼──────────┐   │
│  │                    Core Engine                       │   │
│  │         (SNMP · ARP · ICMP · TCP · HTTP)             │   │
│  └──────────────────────────┬───────────────────────────┘   │
│                             │                               │
│  ┌──────────────────────────▼───────────────────────────┐   │
│  │                 REST API / WebSocket                 │   │
│  └──────────────────────────┬───────────────────────────┘   │
└─────────────────────────────│───────────────────────────────┘
                              │
              ┌───────────────▼────────────────┐
              │          Web Dashboard         │
              │  (Real-time UI · Any Browser)  │
              └────────────────────────────────┘
```

**Key design principles:**
- **Self-contained** — Everything runs on the Pi; no cloud dependency
- **Modular** — Each monitoring service is independent and optional
- **Real-time** — WebSocket-driven live updates across all dashboard panels
- **Lightweight** — Designed to run efficiently on Raspberry Pi 4 (2GB+)
- **Secure** — Tailscale interface binding keeps exposed services private

---

## ✨ Features

### 📊 Network Dashboard
Real-time overview of your entire network at a glance.

- **Public IP Detection** — Instantly see your external IP address with one-click copy
- **ISP Information** — Identify your Internet Service Provider and ASN details
- **Connection Status** — Live uptime tracking with percentage metrics
- **Network Statistics** — Total checks, success rate, average latency, and response time
- **Connection History** — Timestamped log of recent network events

---

### 🖥️ Device Monitoring
Scan and track every device on your network.

- **Network Device Discovery** — Automatically detect all connected devices
- **IP & MAC Address Tracking** — Full inventory with hostname and vendor info
- **Status Filtering** — View Online, Offline, and Conflict devices instantly
- **Open Port Discovery** — See which ports are active on each device
- **Search & Filter** — Find any device by IP, MAC, hostname, or vendor name
- **LAN Wakeup (WoL)** — Remotely wake devices on your local network
- **Domain Expiry Monitor** — Stay ahead of expiring domain registrations

---

### 🔁 Port Forwarding & Service Exposure
Expose internal services securely using intelligent port translation.

- **Port Forwarding Rules** — Translate internal IP:Port combinations to custom external ports
- **Tailscale Interface Binding** — Bind forwarding rules directly to your Tailscale VPN interface for secure, private exposure
- **Multiple IP/Port Mapping** — Manage complex multi-service environments with ease
- **Real-Time Service Monitoring** — Track the live status of all forwarded services

---

### 🔬 Network Tools
A full suite of diagnostic utilities built right in.

| Tool | Description |
|------|-------------|
| **Port Scan** | Scan any host for open TCP/UDP ports |
| **TCP Check** | Verify TCP connectivity to any IP and port |
| **DNS Lookup** | Resolve domains with support for all record types (A, AAAA, MX, CNAME, TXT, SOA, NS, SRV, PTR, and more) |
| **Traceroute** | Visualize the network path to any destination |
| **Ping Monitor** | Continuous latency monitoring for any host |
| **HTTP(S) Check** | Validate HTTP/HTTPS endpoint availability and response |
| **Speed Test** | Measure your real-time internet bandwidth |

---

### 📡 Service Monitoring
Keep tabs on all your critical services.

- **TCP Port Monitoring** — Persistent monitoring of any TCP service
- **HTTP(S) Monitor** — Uptime and status code tracking for web services
- **Ping Monitor** — ICMP-based reachability monitoring
- **SNMP Monitor** — Poll network devices using SNMP for deep hardware metrics
- **Reverse Proxy** — Built-in reverse proxy to route and expose internal services
- **Add/Remove Services** — Simple UI to manage your entire monitoring setup

---

## 🗺️ Roadmap

GoMotz is actively growing. Big things are planned — from a custom peer-to-peer VPN to an Android app and Layer 2 device discovery.

📄 **[View the full Roadmap →](docs/ROADMAP.md)**

---

## 🛠️ Getting Started

> ⚠️ **GoMotz is currently in pre-release.** Installation instructions will be published at launch.

**Planned requirements:**
- Raspberry Pi 4 (2GB RAM or more recommended) or any Linux-based device
- Raspberry Pi OS (64-bit) or compatible Debian-based distribution
- Node.js 18+
- Docker *(optional, for containerized deployment)*

```bash
# Coming soon
git clone https://github.com/yourusername/gomotz.git
cd gomotz
./install.sh
```

---

## 🤝 Contributing

GoMotz is open source and contributions are welcome! Whether it's bug reports, feature suggestions, documentation, or code — all help is appreciated.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

GoMotz is released under the [MIT License](LICENSE).

---

## 💡 Why I Built This

Honestly, this started out of frustration.

[Domotz](https://www.domotz.com/) — while a great product — comes with an expensive hardware cost *and* an ongoing subscription fee that just isn't justifiable for personal or small-scale use. On top of that, I found myself juggling **multiple dashboards and tools**, mentally mapping ports, switching between browser tabs, and losing track of what was running where.

So I built GoMotz.

A single, self-hosted platform that gives me — and now you — **full visibility and control over your network**, without the subscription, without the fragmentation, and without the frustration. Everything in one place, running on hardware you already own.

> *"The best tool is the one you build yourself — because it solves exactly your problem."*

---

## 💬 Acknowledgements

GoMotz is an open-source alternative to [Domotz](https://www.domotz.com/). This project is not affiliated with or endorsed by Domotz. The project draws inspiration from Domotz, Tailscale, MikroTik WinBox, and the broader open-source networking community.

---

<div align="center">

**Built with ❤️ for the open-source community**

⭐ Star this repo if you find it useful — it helps others discover GoMotz!

</div>