# Uptimer — Enterprise Uptime Monitoring Platform

A full-featured clone of UptimeRobot with all enterprise features, built for reliability and performance.

## Features

- **Enterprise Monitoring**: HTTP(S), Ping, TCP Port, UDP, Keyword, API, SSL Certificate, DNS, Domain Expiry, and Heartbeat.
- **High Performance**: Goroutine-based engine capable of checking 1,000+ monitors at 30-second intervals.
- **12 Integrations**: Slate, Microsoft Teams, Discord, Telegram, PagerDuty, Webhooks, Google Chat, Pushbullet, Pushover, Mattermost, Zapier, and Email.
- **Status Pages**: Full-featured public status pages with custom branding and real-time updates.
- **Passwordless Auth**: Secure login via Google, Microsoft, and Apple OAuth.
- **Premium Dashboard**: Glassmorphism design with real-time charts and dark/light mode.
- **Teams & Seats**: Multi-user team support with up to 5 seats included.
- **API v1**: Full REST API for automation and integration.

## Technology Stack

- **Backend**: Go 1.22, Chi Router, pgx, PostgreSQL
- **Frontend**: React 18, TypeScript, Vite, TanStack Query, Zustand, Framer Motion
- **Deployment**: Debian 13, Nginx, Systemd, GitHub Actions

## Installation (Debian 13)

### 1. Initial Server Setup
Run the setup script as root on your Debian 13 host:
```bash
chmod +x deploy/setup.sh
sudo ./deploy/setup.sh
```

### 2. Configure Environment
Copy the example environment file and fill in your OAuth credentials:
```bash
cp .env.example .env
nano .env
```

### 3. Deploy via GitHub Actions
1. Push this repository to your GitHub account.
2. In GitHub Settings > Secrets > Actions, add the following secrets:
   - `DEPLOY_HOST`: Your server IP
   - `DEPLOY_USER`: Your SSH username
   - `DEPLOY_SSH_KEY`: Your SSH private key
3. Push to the `main` branch to trigger the deployment.

## Development

### Prerequisites
- Go 1.22+
- Node.js 20+
- Docker (for local PostgreSQL)

### Running Locally
1. Start the database: `make run-db`
2. Run backend: `cd backend && go run cmd/uptimer/main.go`
3. Run frontend: `cd frontend && npm run dev`

### Testing
Run all tests using the Makefile:
```bash
make test-all
```

## License
MIT
