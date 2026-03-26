# SvcWatch 🚀

SvcWatch is a professional, full-stack Nginx log monitoring and analytics dashboard. It "tails" Nginx access logs in real-time, parses them into structured metrics, provides an authorized REST API, and visualizes data on a modern, responsive web dashboard.

## 🏗️ Monorepo Structure

The project is organized as a monorepo for clean separation of concerns:

- **`/backend`**: High-performance Go service using [Gin](https://github.com/gin-gonic/gin) and [Redis](https://redis.io/).
- **`/frontend`**: Modern analytics dashboard built with [Vue 3](https://vuejs.org/), [Vite](https://vitejs.dev/), and [TypeScript](https://www.typescriptlang.org/).
- **`/.github`**: Multi-server parallel deployment workflows with GitHub Actions.

## ✨ Key Features

- **📡 Real-time Analytics**: Intelligent agent that tails logs, handles log rotation automatically, and parses entries on-the-fly using robust regex.
- **📊 Modern Dashboard**: Visualize QPS, status code distributions, and traffic spikes with interactive charts and real-time updates.
- **🔐 Secure API**: Integrated with external Auth services for per-request permission validation for all monitoring endpoints.
- **💾 Redis Persistence**: Efficient data storage for long-term metric aggregation and persistence across service restarts.
- **⚙️ Matrix Deployment**: Automatic, parallel deployment to multiple production servers via GitHub Actions Matrix strategy.
- **🛡️ Nginx Reverse Proxy**: Secure HTTPS/SSL mapping via Nginx (port 8080) with automated configuration and Certbot integration.

## 🚀 Getting Started

### Backend (Go)

1.  Navigate to the `backend/` directory.
2.  Install dependencies: `go mod tidy`.
3.  Configure your environment in `config/config.yaml`.
4.  Run natively: `go run cmd/server/main.go`.

### Frontend (Vue)

1.  Navigate to the `frontend/` directory.
2.  Install dependencies: `npm install`.
3.  Start the development server: `npm run dev`.

## 📦 Deployment (CI/CD)

The project includes pre-configured CI/CD pipelines for automated, targeted deployments.

1.  **Backend Deployment**: [deploy-backend.yml](.github/workflows/deploy-backend.yml) triggers only on `backend/**` changes.
2.  **Frontend Deployment**: [deploy-frontend.yml](.github/workflows/deploy-frontend.yml) triggers only on `frontend/**` changes.

### Required Setup
To use the CI/CD pipeline, configure the following in GitHub:

**Repository Variables (Settings → Secrets and variables → Actions → Variables):**
- `DEPLOY_TARGETS`: JSON array of server metadata (IP, domain, log paths).

**Repository Secrets (Settings → Secrets and variables → Actions → Secrets):**
- `SERVER_SSH_KEY`: Private RSA/ED25519 key for SSH access.
- `AUTH_PASSPORT_URL`, `AUTH_PERMISSION_URL`, `AUTH_SYS_CODE`: Authentication service integration.
- `REDIS_ADDR`, `REDIS_PASSWORD`: Redis connection credentials.

## 🛠️ Tech Stack

- **Languages**: [Go](https://go.dev/), [TypeScript](https://www.typescriptlang.org/)
- **Backend Framework**: [Gin Gonic](https://github.com/gin-gonic/gin)
- **Frontend Framework**: [Vue 3](https://vuejs.org/) + [Vite](https://vitejs.dev/)
- **Storage**: Redis, In-Memory
- **Automation**: GitHub Actions, Nginx, Certbot
