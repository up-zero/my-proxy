<div align="center">
    <img src="./assets/logo.png" alt="my-proxy logo" width="300">
</div>

<p align="center">
   <a href="https://github.com/up-zero/my-proxy/fork" target="blank">
      <img src="https://img.shields.io/github/forks/up-zero/my-proxy?style=for-the-badge" alt="GitHub forks"/>
   </a>
   <a href="https://github.com/up-zero/my-proxy/stargazers" target="blank">
      <img src="https://img.shields.io/github/stars/up-zero/my-proxy?style=for-the-badge" alt="GitHub stars"/>
   </a>
   <a href="https://github.com/up-zero/my-proxy/pulls" target="blank">
      <img src="https://img.shields.io/github/issues-pr/up-zero/my-proxy?style=for-the-badge" alt="GitHub pull requests"/>
   </a>
   <a href='https://github.com/up-zero/my-proxy/releases'>
      <img src='https://img.shields.io/github/release/up-zero/my-proxy?&label=Latest&style=for-the-badge' alt="Latest release">
   </a>
</p>

<p align="center">
   English | <a href="./README_zh.md">中文</a>
</p>

A LAN proxy tool that supports proxy forwarding for TCP, UDP, HTTP, SOCKS5, and other protocols, making it suitable for most network environments. It provides both CLI and WebUI configuration modes, greatly simplifying proxy setup. Supports multi-node management — control multiple my-proxy instances from a single dashboard.

## WebUI Mode

+ Dashboard

![webui_dashboard.png](./assets/webui_dashboard.png)

+ Proxy Management

![webui_proxy.png](./assets/webui_proxy.png)

## CLI Mode

+ Start the service

```bash
# Default service port: 12312
my-proxy serve
# Specify a custom service port
my-proxy serve -p 12312
```

+ Proxy status

```bash
# View the status of all proxies by default
my-proxy status
# View the status of a specific proxy
my-proxy status <name>
```

![cli_status.png](./assets/cli_status.png)

+ Terminal dashboard

```bash
# Open terminal stats
my-proxy stats

# Customize the refresh interval
my-proxy stats --interval 2s
```

The terminal stats view mirrors the core Web dashboard elements, including summary metrics, rate trends, connection trends, system resource usage, and the Top N node load list. Press `r` to refresh immediately and `q` to exit.

+ Proxy management

```bash
# Start a proxy
my-proxy start <name>

# Stop a proxy
my-proxy stop <name>

# Restart a proxy
my-proxy restart <name>

# Create a proxy with the TUI
my-proxy create <name>
# Quick create
my-proxy create --name my_proxy --type TCP --lport 9090 --taddr 192.168.1.1 --tport 9000

# Edit a proxy
my-proxy edit <name>

# Delete a proxy
my-proxy delete <name>
```

Interactive command-line interfaces are provided for creating and editing proxies, making operation more convenient.

![cli_create.png](./assets/cli_create.png)

## Deployment

### One-click install on Linux / macOS

Install the latest release directly from GitHub Releases:

```bash
curl -fsSL https://raw.githubusercontent.com/up-zero/my-proxy/master/scripts/install.sh | bash
```

Install a specific version, custom directory, or custom service port:

```bash
# Install a specific release
curl -fsSL https://raw.githubusercontent.com/up-zero/my-proxy/master/scripts/install.sh | MY_PROXY_VERSION=v1.0.0 bash

# Install to a user directory
curl -fsSL https://raw.githubusercontent.com/up-zero/my-proxy/master/scripts/install.sh | INSTALL_DIR="$HOME/.local/bin" bash

# Change the service port
curl -fsSL https://raw.githubusercontent.com/up-zero/my-proxy/master/scripts/install.sh | MY_PROXY_SERVICE_PORT=12312 bash
```

After installation, run `my-proxy version` to verify the binary, then use the printed service status command to inspect the running service.

### Manual deployment with supervisor (Linux)

1. Upload the `my-proxy` executable to `/usr/local/bin`.

2. Install `supervisor`, then create `/etc/supervisor/conf.d/my-proxy.conf` (note: the configuration path may vary by supervisor version; for example, on CentOS you may need to create `/etc/supervisord.d/my-proxy.ini`). Use the following content:

```conf
[program:my-proxy]
# start command
command=/usr/local/bin/my-proxy serve
# start automatically
autostart=true
# restart automatically
autorestart=true
# environment variable
environment=HOME="/root"
```

3. Reload the `supervisor` configuration and start the service.

```bash
sudo supervisorctl reread
sudo supervisorctl update
sudo supervisorctl restart my-proxy
```

4. Run the following command to retrieve version information. If it returns output similar to the example below, the installation was successful.

```bash
sudo my-proxy info

# Example output
my-proxy 1.1.0
+----------+-------------------------+
| Address  | http://10.0.0.11:12312  |
|          | http://172.17.0.1:12312 |
| Username | admin                   |
| Password | KDi7tW6Y                |
+----------+-------------------------+
```

## Docker Deployment

Run it with `docker run`:

```bash
# Create the mount directory
mkdir -p my-proxy/data

# Start the container
docker run -d \
    --name my-proxy-service \
    --restart always \
    --network host \
    -v "./my-proxy/data:/root/.config/my-proxy" \
    getcharzp/my-proxy:1.1.0

# View the login account
docker logs my-proxy-service | grep "admin"
```

## Build

```bash
# Linux / macOS
./scripts/build-release.sh --version 1.1.0 --clean
```
