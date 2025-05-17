# DNS Failover for Cloudflare

[![license](https://img.shields.io/badge/license-MIT-blue.svg)](https://www.google.com/search?q=MIT LICENSE)

An automated tool to monitor the health of specified IP addresses and dynamically update Cloudflare DNS records. It removes offline IPs and re-adds them if they become online, ensuring your domain primarily points to healthy servers. This is particularly useful for maintaining high availability for services hosted on servers with dynamic or multiple IP addresses.

## ✨ Key Features

* **Automatic Health Checks:** Periodically checks the reachability of configured IP addresses.
* **Dynamic Cloudflare DNS Updates:** Automatically adds or removes DNS A/AAAA records in Cloudflare based on IP health.
* **Multi-Domain & Multi-IP Support:** Manage multiple IP addresses across several domains simultaneously.
* **Flexible Configuration:** Easy-to-understand YAML configuration for Cloudflare API, health check parameters, and domain settings.
* **Cloudflare Proxy Support:** Configure whether DNS records should be proxied through Cloudflare (orange cloud) or be DNS-only (grey cloud).
* **Enhanced Service Availability:** Helps improve service uptime by ensuring DNS records point only to responsive IP addresses.

## 🚀 How It Works

1.  **Load Configuration:** The application reads its settings from the `config.yaml` file upon startup.
2.  **IP Health Monitoring:**
    * At intervals defined by `ip_test.interval`, the tool checks each IP address listed under `domains`.
    * For each IP, it performs `ip_test.sampling` number of probes (e.g., ICMP pings or other configured checks).
    * An IP is considered "online" if a sufficient number of probes succeed (e.g., more than 50% by default).
    * Each probe has a timeout defined by `ip_test.timeout`.
3.  **DNS Record Synchronization:**
    * The tool fetches the current DNS records for the managed domain name from Cloudflare.
    * **Add IP:** If a monitored IP is detected as "online" but is not present in the Cloudflare DNS records for the domain, the tool adds it.
    * **Remove IP:** If a monitored IP is detected as "offline" (or removed from the config) but still exists in the Cloudflare DNS records, the tool removes it.


## 📄 Configuration

The project is configured via a `config.yaml` file in the root directory. Create one based on the example below:

```yaml
provider:
  name: cloudflare # Currently, only "cloudflare" is supported
  key: "YOUR_CLOUDFLARE_API_TOKEN" # API Token

ip_test:
  interval: 600s   # Health check interval (e.g., 600s = 10 minutes). Supports 's', 'm', 'h'.
  sampling: 10     # Number of health check samples per IP in one cycle (e.g., 10 pings).
  timeout: 5s      # Timeout for each individual sample (e.g., 5s per ping).
  # (Optional) success_threshold_percent: 50 # Percentage of successful samples needed to consider an IP online (default: 50).

domains:
  - name: "xxx.xxxxxx.xyz" # The fully qualified domain name (FQDN) to manage.
    proxied: false         # Cloudflare proxy status: true for orange cloud, false for grey cloud (DNS only).
    ip_type: "A"          # IP address type: "A" (for A records) or "AAAA" (for AAAA records).
    ips:                   # List of IP addresses to monitor and manage for this domain.
      - "1.1.1.1,9.9.9.9"  # Choose the best quality IP.
      - "8.8.8.8"

  - name: "another.example.com"
    proxied: true
    ip_type: "A"
    multiple: false
    ips:
      - "1.2.3.4"
      - "5.6.7.8"
    ip_test:               #Override global configuration
      interval: 60s
      sampling: 1
      timeout: 5s
```

## Docker

```
docker run -it --rm --name dns-failover \
    --restart=always \
    -e LOG_LEVEL=debug \
    -v ./config.yaml:/config.yaml \
    heifeng/dns-failover:latest
```


