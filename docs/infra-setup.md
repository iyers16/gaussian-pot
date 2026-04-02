# Gaussian Pot — Infrastructure Setup Documentation

## Stack Overview

| Layer | Technology |
|---|---|
| Frontend | Next.js 16 (standalone mode) |
| Backend | Go 1.25 + Gin |
| Database | Postgres 16 |
| Reverse proxy | Nginx |
| Containerization | Docker + Docker Compose |
| Container registry | GitHub Container Registry (GHCR) |
| CI/CD | GitHub Actions |
| VPS | DigitalOcean Droplet — Toronto (TOR1) |

---

## 1. DigitalOcean VPS Provisioning

**Droplet specs chosen:**
- Region: Toronto (TOR1)
- OS: Ubuntu 24.04 LTS x64
- Plan: Basic — 2 vCPU, 2GB RAM, 60GB SSD — $18/mo
- Auth: SSH key (ed25519)

**Generate SSH key locally before provisioning:**
```bash
ssh-keygen -t ed25519 -C "gaussian-pot"
cat ~/.ssh/id_ed25519.pub  # paste this into DigitalOcean during setup
```

---

## 2. Initial Server Hardening

SSH in as root for the first and last time:
```bash
ssh root@<ip>
```

**Update the system:**
```bash
apt update && apt upgrade -y
```

**Create a non-root user:**
```bash
adduser user
usermod -aG sudo user
rsync --archive --chown=user:user ~/.ssh /home/user
```

**Disable root SSH login:**
```bash
nano /etc/ssh/sshd_config
# Change: PermitRootLogin yes → PermitRootLogin no
systemctl restart ssh
```

**Verify new user works in a separate terminal before closing root session:**
```bash
ssh user@<ip>
```

From here on, always SSH as `user`. Never root.

---

## 3. Docker Installation

```bash
sudo apt install -y ca-certificates curl gnupg

sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

sudo usermod -aG docker user

# Log out and back in for group change to take effect
exit
ssh user@<ip>

# Verify
docker run hello-world
```

---

## 4. Firewall Configuration

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

sudo ufw status
```

Expected output:
```
Status: active
To                         Action      From
--                         ------      ----
OpenSSH                    ALLOW       Anywhere
80/tcp                     ALLOW       Anywhere
443/tcp                    ALLOW       Anywhere
```

---

## 5. Project Structure

```
gaussian-pot/
├── backend/
│   ├── cmd/server/main.go       # entrypoint, DB init, routes
│   ├── internal/
│   │   ├── handler/number.go    # HTTP handlers
│   │   ├── model/number.go      # data structs
│   │   └── repository/number.go # DB queries
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/app/page.tsx         # main UI
│   ├── Dockerfile
│   ├── next.config.ts           # output: standalone
│   └── ...
├── nginx/
│   └── nginx.conf               # reverse proxy config
├── .github/
│   └── workflows/
│       └── deploy.yml           # CI/CD pipeline
├── docker-compose.yml           # production (pulls from GHCR)
├── docker-compose.dev.yml       # local dev (builds locally)
├── .env                         # never committed
└── .env.example                 # committed, no real values
```

---

## 6. Docker Architecture

### How containers communicate

All four containers run on a Docker bridge network. Docker Compose creates an internal DNS so containers reference each other by service name, not IP:

```
Internet → Nginx (:80) → frontend (:3000)
                       → backend (:8080) → db (:5432)
```

- Nginx is the only container with a public port (80)
- Frontend, backend, db are internal only
- Postgres is never exposed to the internet

### Key Docker Compose decisions

**Production (`docker-compose.yml`)** — pulls pre-built images from GHCR:
```yaml
image: ghcr.io/iyers16/gaussian-pot-frontend:latest
```

**Dev (`docker-compose.dev.yml`)** — builds images locally from Dockerfiles:
```yaml
build:
  context: ./frontend
  dockerfile: Dockerfile
```

### Named volume for Postgres persistence
```yaml
volumes:
  - postgres_data:/var/lib/postgresql/data
```
Data survives container restarts and redeployments.

### Frontend must bind to 0.0.0.0
```yaml
environment:
  - HOSTNAME=0.0.0.0
```
Without this Next.js standalone only listens on the container hostname, Nginx can't reach it.

---

## 7. Nginx Configuration

```nginx
events {}

http {
    server {
        listen 80;

        location /api/ {
            proxy_pass http://backend:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        location / {
            proxy_pass http://frontend:3000;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}
```

All `/api/*` requests go to the Go backend. Everything else goes to Next.js. CORS is never an issue because everything is under one domain/IP.

---

## 8. GitHub Container Registry (GHCR) Setup

**On the VPS — authenticate Docker with GHCR:**

Generate a GitHub Personal Access Token:
- GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
- Scope: `read:packages` only
- Expiry: 90 days

```bash
echo YOUR_PAT | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin
```

**GHCR image naming convention:**
```
ghcr.io/iyers16/gaussian-pot-frontend:latest
ghcr.io/iyers16/gaussian-pot-backend:latest
```

---

## 9. GitHub Actions CI/CD Pipeline

### Secrets required (repo → Settings → Secrets → Actions)

| Secret | Value |
|---|---|
| `DROPLET_HOST` | `<ip>` |
| `DROPLET_USER` | `user` |
| `DROPLET_SSH_KEY` | contents of `~/.ssh/id_ed25519` (private key) |

### Pipeline flow

```
git push to main
      ↓
build-frontend (parallel) ──┐
build-backend  (parallel) ──┤
                            ↓
                         deploy
                            ↓
                  SSH into VPS
                  git pull origin main
                  docker compose pull
                  docker compose up -d
                  docker image prune -f
```

### Path filtering — only rebuild what changed

- `frontend/**` or `docker-compose.yml` changed → rebuild frontend image
- `backend/**` or `docker-compose.yml` changed → rebuild backend image
- Deploy always runs to pick up any compose/config changes

### Build caching

Images are cached in GHCR between runs. First build is slow (~3-5 min), subsequent builds only rebuild changed layers (~30-60s).

---

## 10. VPS Setup for Deployment

After provisioning and hardening, clone the repo and create the env file:

```bash
cd ~
git clone https://github.com/iyers16/gaussian-pot.git
cd gaussian-pot
```

The `.env` file is never in git — it must be manually created on each server.

---

## 11. Development Loop

### Daily workflow

```bash
# 1. Write code locally
# 2. Commit and push
git add .
git commit -m "feat: your change"
git push

# 3. Watch pipeline (GitHub → Actions tab)
# 4. Changes live in ~2-3 minutes
```

### Local development

```bash
# Run full stack locally
docker compose -f docker-compose.dev.yml up --build

# Access at http://localhost (port 80 via Nginx)
# NOT http://localhost:3000 — that bypasses Nginx
```

### Debugging on the VPS

```bash
ssh user@<ip>

# Check container status
docker ps

# Check logs
docker logs gaussian-pot-backend-1 --tail 50
docker logs gaussian-pot-frontend-1 --tail 50
docker logs gaussian-pot-nginx-1 --tail 50
docker logs gaussian-pot-db-1 --tail 50

# Restart all containers
cd ~/gaussian-pot
docker compose restart

# Pull latest images manually
docker compose pull
docker compose up -d

# Clean up old images
docker image prune -f
```

---

## 12. SSL / Domain (Future)

When ready to add HTTPS:

1. Buy a domain (Cloudflare Registrar, Namecheap, etc.)
2. Create an `A record` pointing to `<ip>`
3. SSH into VPS and run:

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com
```

Certbot automatically edits the Nginx config and sets up auto-renewal. No other infrastructure changes needed.

---

## Key Decisions Summary

| Decision | Choice | Reason |
|---|---|---|
| VPS location | Toronto | Closest to Montreal, Canadian data residency |
| OS | Ubuntu 24.04 LTS | Long term support, Docker support |
| Non-root user | Yes | Security hardening |
| Firewall | ufw | Simple, standard |
| Container registry | GHCR | Free, integrated with GitHub |
| Image architecture | x86 | DigitalOcean is x86, no cross-compilation needed |
| DB management | Docker volume | Persistence without managed DB cost |
| Nginx placement | Separate container | Clean separation, single entry point |
| Local dev | docker-compose.dev.yml | Mirrors production exactly |
| Deploy trigger | git push to main | Simple, automatic, auditable |