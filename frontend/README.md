# SvcWatch Frontend 📊

The frontend of SvcWatch is a high-performance, real-time analytics dashboard built with **Vue 3**, **Vite**, and **TypeScript**. It visualizes monitoring data from the Go backend, providing insights into traffic, latency, and error rates.

## ✨ Features

- **Real-time Visualization**: Interactive charts for QPS and status distributions.
- **Modern UI**: Built with **Tailwind CSS 4** for a sleek, responsive experience.
- **Dark Mode Support**: Seamless transition between light and dark themes.
- **Secure Authentication**: Integration with the Passport service for protected monitoring views.

## 🛠️ Tech Stack

- **Framework**: Vue 3 (Composition API)
- **Build Tool**: Vite
- **Styling**: Tailwind CSS 4
- **State Management**: Pinia
- **Language**: TypeScript

## 🚀 Getting Started

### Prerequisites

- Node.js (v20 or higher)
- npm

### Installation

```bash
# Navigate to the frontend directory
cd frontend

# Install dependencies
npm install
```

### Development

```bash
# Run the development server
npm run dev
```

The app will be available at `http://localhost:5173`.

### Production Build

```bash
# Type-check and build for production
npm run build
```

The output will be in the `dist/` directory.

## 📦 Deployment

This project uses GitHub Actions for automated deployment. The workflow is defined in `.github/workflows/deploy-frontend.yml`.

### Required GitHub Secrets

To deploy properly, ensure the following secrets are configured in your GitHub repository:

- `SERVER_SSH_KEY`: Private SSH key for server access.
- `FRONTEND_HOST`: The IP address or hostname of your frontend server.
- `FRONTEND_USER`: The SSH username (e.g., `ubuntu`).

### Target Server Setup

The deployment script expects:
- Nginx installed and configured as a reverse proxy.
- SSL certificates managed by Certbot (at `watch.dongyuhan.com`).
- The app directory at `/home/<user>/app/frontend`.
