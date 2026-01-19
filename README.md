# mdspace

A lightweight markdown sharing platform. Paste or drop markdown files to get instant shareable links with live preview, line comments, and view tracking.

## Features

- **Paste or Drop** - Upload markdown via text paste or file drag-and-drop
- **Live Preview** - See rendered markdown with syntax highlighting in real-time
- **Shareable Links** - Get instant URLs for sharing with anyone
- **Line Comments** - Viewers can add comments on specific lines
- **View Tracking** - See how many people have viewed your markdown

## Tech Stack

- **Backend**: Go with Chi router
- **Storage**: Redis with TTL-based expiration
- **Frontend**: Vanilla JS with marked.js and highlight.js
- **Deployment**: Docker, Railway-ready

## Local Development

### Prerequisites

- Go 1.22+
- Redis (or via Docker)

### Setup

1. Clone the repository
2. Copy environment variables:
   ```bash
   cp .env.example .env
   ```
3. Start Redis:
   ```bash
   docker run -d -p 6379:6379 redis:alpine
   ```
4. Run the server:
   ```bash
   go run ./cmd/server
   ```
5. Open http://localhost:8080

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/markdown` | Create new markdown |
| GET | `/api/markdown/{id}` | Get markdown content |
| DELETE | `/api/markdown/{id}` | Delete markdown (owner only) |
| POST | `/api/markdown/{id}/comments` | Add line comment |
| GET | `/api/markdown/{id}/comments` | Get all comments |
| GET | `/api/user/stats` | Get user file count |

## License

Licensed under the [MIT](LICENSE) license.
