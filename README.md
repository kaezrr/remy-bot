# remy-bot
A versatile WhatsApp bot designed to manage and enhance communication within a college or community group, built using Go and the `whatsmeow` library.

## Prerequisites
To host and run this bot, you need:
1.  **A Dedicated WhatsApp Account:** The bot operates by linking to a WhatsApp account (like WhatsApp Web). This account **must** remain logged in, active, and should ideally be a separate number specifically for the bot.
2.  **Docker & Docker Compose:** For the easiest setup and hosting.

## Quick Start (Docker Compose)
The easiest way to get `remy-bot` running and ensure its session data persists is by using Docker Compose.

### Step 1: Configuration
Before building, you **must** configure the bot's settings by editing the `config.json` file in the root directory.
```json
{
  "database": "data/remy.db",
  "session_dir": "data/session",
  "prefix": ".",
  "target_group_name": "Test Group" // <--- IMPORTANT: Change this to your target group's name
}
```

### Step 2: Build and Run
Run the following command from the root directory containing your `Dockerfile`, `docker-compose.yml`, and `config.json`:
```bash
docker compose up -d --build
```

### Step 3: Initial Login (Scan QR Code)
Since this is the first time running the bot, you need to link the WhatsApp account. The bot will print a QR code to the container's logs.
1.  **View Logs:** Access the live logs to find the QR code:
    ```bash
    docker compose logs -f
    ```
2.  **Scan:** Use the WhatsApp account you wish to dedicate to the bot to **Link a Device** (WhatsApp on your phone -> Settings/Menu -> Linked Devices).
3.  **Wait:** Once the QR code is scanned and the connection is successful, the log messages will show a connection status. You can stop viewing the logs with `Ctrl+C`. The bot will continue running in the background.

## Development Commands
If you prefer to build and run the application without Docker, you can use the standard Go commands (provided in a Makefile).

### Build (Linux/macOS)
```bash
make build
```
### Run (Local)
```bash
make run
```

## Management
| Command                               | Description                                                                                                                            |
| :------------------------------------ | :------------------------------------------------------------------------------------------------------------------------------------- |
| `docker compose stop`                 | Gracefully stops the running bot container.                                                                                            |
| `docker compose start`                | Starts the bot container again without rebuilding.                                                                                     |
| `docker compose down`                 | Stops the container AND removes the container instance (but preserves the `remy_data` volume containing your session/DB).              |
| `docker volume rm remy-bot_remy_data` | **DELETES ALL BOT DATA and session information.** Use this only if you want to reset the bot entirely and link a new WhatsApp account. |
