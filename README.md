# Insider Assessment - Automatic Messaging System

This project is a message scheduler and sender built with Go, PostgreSQL, and Redis. It automatically processes pending messages from the database and sends them to a configured Webhook endpoint.

## Features

-   **Automatic Scheduling:** Sends 2 pending messages every 2 minutes using native Go `time.Ticker` (no cron packages).
-   **Concurrency:** Start/Stop control via API.
-   **Redis Caching:** Caches sent message IDs and timestamps.
-   **Dockerized:** Complete environment setup with Docker Compose.
-   **Swagger Documentation:** Auto-generated API docs.

## Prerequisites

-   Docker & Docker Compose
-   Make (optional, for convenience commands)

## Quick Start

1.  **Configure Environment:**
    The project comes with a default configuration in `docker-compose.yml` and `.env`.
    You can update the `WEBHOOK_URL` in `docker-compose.yml` to your own [Webhook.site](https://webhook.site) URL if desired.

2.  **Run with Docker:**
    ```bash
    make docker-build
    make docker-up
    ```
    Or manually:
    ```bash
    docker-compose up --build -d
    ```

3.  **Seed Data:**
    Populate the database with test messages:
    ```bash
    make seed
    ```
    *Alternatively, you can add messages manually via the API.*

4.  **Monitor:**
    The scheduler starts automatically. Watch the logs to see the worker in action:
    ```bash
    docker-compose logs -f app
    ```

## API Documentation

Swagger UI is available at: **http://localhost:8080/swagger/index.html**

### Endpoints

-   **Control**
    -   `POST /start` - Resumes the automatic message sender.
    -   `POST /stop` - Pauses the automatic message sender.

-   **Messages**
    -   `GET /sent-messages` - Retrieves a list of all successfully sent messages.
    -   `POST /messages` - Adds a new message to the queue (Status: PENDING).
    -   `GET /messages/cache` - Retrieves all sent messages currently stored in Redis.

-   **System**
    -   `GET /health` - Health check endpoint.

## Development

### Project Structure
-   `cmd/server`: Application entry point.
-   `internal/model`: Database models and hooks.
-   `internal/repository`: Database access layer (GORM).
-   `internal/service`: Business logic (Scheduler and Worker).
-   `internal/handler`: HTTP handlers.
-   `internal/config`: Configuration management.

### Useful Commands
-   `make run`: Run the application locally.
-   `make test`: Run unit tests.
-   `make swag`: Regenerate Swagger documentation.
-   `make lint`: Run linter.
-   `make seed`: Insert test data into the running DB.

## Configuration

Environment variables can be set in `.env` or `docker-compose.yml`:

| Variable | Default | Description |
|----------|---------|-------------|
| `WEBHOOK_URL` | (Set in compose) | Target URL for sending messages |
| `WORKER_BATCH_SIZE` | `2` | Number of messages to process per tick |
| `WORKER_INTERVAL` | `2m` | Time between worker runs |
| `REDIS_TTL` | `24h` | Expiration time for Redis cache |