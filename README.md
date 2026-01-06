# HttpProxyByHost

A lightweight, secure forward proxy designed for distributed teams to access restricted upstream resources (e.g., via a VPN or upstream proxy) simply by modifying local DNS/Hosts configurations.

## Overview

This tool acts as a "Reverse-like" Forward Proxy. It listens on standard HTTP/HTTPS ports (80/443) and forwards traffic to the destination specified in the `Host` header. 

**Core Use Case:** 
Deploy this tool on a server that has access to restricted content (e.g., a server with a VPN connection or in a specific region). Users can then point specific domains (e.g., `google.com`) to this server's IP in their local `/etc/hosts`. The server will proxy the request to the actual destination, optionally tunneling through another upstream proxy (HTTP/SOCKS).

**Key Features:**
*   **Zero-Config Client:** No need to install VPN clients or configure proxy settings on user machines. Just update DNS/Hosts.
*   **Upstream Proxy Support:** Can chain to another proxy (HTTP/SOCKS5).
*   **Automatic SSL:** Can generate self-signed certificates on the fly (for testing or internal use).
*   **HTTPS Support:** Listens on port 443 and handles TLS termination (note: see SSL Limitations).

## Quick Start

### 1. Installation

Download the binary from the [Releases](https://github.com/friddle/http-proxy-by-host/releases) page or build it yourself:

```bash
go build -o http-proxy-by-host
```

### 2. Usage

**Basic Usage (Direct Connection):**
Listens on ports 80 and 443.

```bash
sudo ./http-proxy-by-host --ssl=generate
```

**With Upstream Proxy:**
Forward all traffic through a local SOCKS5 or HTTP proxy (e.g., a VPN client running on port 7897).

```bash
sudo ./http-proxy-by-host --ssl=generate --proxy=http://127.0.0.1:7897
```

**Custom Ports & Certificates:**

```bash
sudo ./http-proxy-by-host \
  --http-port=8080 \
  --https-port=8443 \
  --ssl=on \
  --ssl-crt=/path/to/cert.crt \
  --ssl-key=/path/to/key.key
```

### 3. Client Configuration (The "DNS Spoofing" Method)

On your local machine (Client), edit your hosts file (`/etc/hosts` on macOS/Linux, `C:\Windows\System32\drivers\etc\hosts` on Windows):

```text
# Point google.com to your proxy server IP
192.168.1.100 www.google.com
192.168.1.100 google.com
```

Now, when you visit `http://www.google.com` or `https://www.google.com`, the request goes to `192.168.1.100`. The proxy reads `Host: www.google.com` and forwards the request to the real Google servers (potentially via the upstream proxy).

## CLI Arguments

| Flag | Default | Description |
|------|---------|-------------|
| `--ssl` | `none` | SSL Mode: `generate` (auto-create certs), `on` (use provided files), `none`. |
| `--ssl-crt` | - | Path to SSL certificate file (required if `--ssl=on`). |
| `--ssl-key` | - | Path to SSL key file (required if `--ssl=on`). |
| `--proxy` | - | Upstream proxy URL (e.g., `http://127.0.0.1:7897` or `socks5://...`). |
| `--http-port`| `80` | Port to listen for HTTP traffic. |
| `--https-port`| `443` | Port to listen for HTTPS traffic. |

## SSL/HTTPS Limitations

When using the `--ssl=generate` mode, the server generates a self-signed certificate for `*.reserver.proxy`. 

If you point a real domain (e.g., `google.com`) to this proxy via DNS:
1.  **Browser Warning:** Your browser will show a "Not Secure" or "Your connection is not private" warning because the certificate provided by the proxy (`*.reserver.proxy`) does not match the domain (`google.com`).
2.  **Acceptance:** You must manually proceed (e.g., click "Advanced" -> "Proceed to ...") to access the site.
3.  **HSTS:** Sites with strict HSTS (like Google) might block this entirely.

**Solution:** For production use with specific domains, generate valid certificates (e.g., via Let's Encrypt) for the domains you intend to spoof and run the proxy with `--ssl=on --ssl-crt=...`.

## License

MIT
