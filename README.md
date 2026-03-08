# 🗑️ PurgeBot 🗑️

Welcome to **PurgeBot**! This bot helps manage and clean up your Discord server by automatically purging old messages from channels based on your specified duration. Whether you need to clear out outdated messages or keep your channels tidy, PurgeBot has you covered.

## 🚀 Features

- **Purge Old Messages**: Automatically delete messages older than a specified duration.
- **Stop Purging**: Easily stop the purging task for a channel.
- **List Purge Tasks**: Get a list of all active purge tasks in your guild.
- **Add or Remove Users/Roles**: Grant or revoke permission for specific users or roles to manage purge tasks.

## Prerequisites

- **Docker and Docker Compose** (recommended), or Go installed locally
- **Discord Bot Token**: Create a bot on [Discord Developer Portal](https://discord.com/developers/applications) and get your token
- The bot uses SQLite for database storage (no separate install when using Docker)

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/keshon/purge-bot.git
cd purge-bot
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and set `DISCORD_KEY` to your Discord bot token.

### 3. Run

**With Docker (recommended):**

```bash
docker compose up -d
```

The database is stored in a Docker volume (`purgebot-data`). Logs: `docker compose logs -f purgebot`.

**Without Docker:**

```bash
go mod tidy
go run ./cmd/purgebot
```

You can set `DISCORD_KEY` in the environment instead of using `.env`, or use a `.env` file in the current directory (or `-env /path/to/.env`).

### Command-line options (binary only)

When running the binary directly (not Docker):

- `-env`: Path to `.env` file. If omitted, the bot loads `.env` from the current working directory or uses environment variables.
- `-db`: Path to database file. Default: `database.db`.
- `-log-level`: `debug`, `info`, `warn`, `error`. Default: `info`.
- `-log-format`: `text` or `json`. Default: `text`.
- `-log-file`: Optional log file path (size-based rotation).

Examples:

```bash
go run ./cmd/purgebot -env /etc/purgebot/.env
go run ./cmd/purgebot -db /var/lib/purgebot/purge.db
```

## Docker Compose

Run PurgeBot with Docker (same pattern as the root `docker-compose.yml`):

```yaml
services:
  purgebot:
    build: .
    restart: unless-stopped
    volumes:
      - purgebot-data:/data
    env_file: .env

volumes:
  purgebot-data:
```

The database is stored in the `purgebot-data` volume. Use `cp .env.example .env` and set `DISCORD_KEY` before `docker compose up -d`.

## 🧪 Testing

Run tests the same way as CI (so you catch the same failures locally):

- **Unix / macOS / Git Bash:**  
  `./test.sh`  
  (Make it executable once: `chmod +x test.sh`)

- **Windows (cmd):**  
  `test.bat`

- **Or run the CI command directly:**  
  `go test -v ./... -count=1`

**Note:** Tests use a pure-Go SQLite driver; no C compiler is required (Windows, Linux, macOS).

**Lint (optional):**  
`golangci-lint run ./...`

## 📜 Commands

### Purge Old Messages

Automatically purge old messages in the channel. You can use a bare duration or the `messages` subcommand.

- **Usage:** `@PurgeBot <duration>` or `@PurgeBot messages <duration>`
- **Example:** `@PurgeBot 3d` or `@PurgeBot messages 3d` (purges messages older than 3 days)
- **Stop messages only:** `@PurgeBot messages stop`

### Delete Old Threads

Delete threads under this channel that are older than the given duration (the thread itself is deleted, not its messages). Uses a separate duration from message purge (e.g. messages 3d, threads 6d).

- **Usage:** `@PurgeBot threads <duration>`
- **Example:** `@PurgeBot threads 6d` (deletes threads under this channel older than 6 days)
- **Stop threads only:** `@PurgeBot threads stop`

### Stop All Tasks

Stop both message purge and thread cleanup for this channel.

- **Usage:** `@PurgeBot stop`

### List Purge Tasks

Get a list of all channels with active purge tasks in the guild.

- **Usage:** `@PurgeBot list`

### Add User

Grant a user permission to manage purge tasks. You can use either username or user ID.

- **Usage:** `@PurgeBot adduser <username>` or `@PurgeBot adduserid <userID>`
- **Example:** `@PurgeBot adduser JohnDoe` or `@PurgeBot adduserid 339767128292982785`

### Remove User

Revoke a user's permission to manage purge tasks. You can use either username or user ID.

- **Usage:** `@PurgeBot removeuser <username>` or `@PurgeBot removeuserid <userID>`
- **Example:** `@PurgeBot removeuser JohnDoe` or `@PurgeBot removeuserid 339767128292982785`

### Add Role

Grant a role permission to manage purge tasks. You can use either role name or role ID.

- **Usage:** `@PurgeBot addrole <roleName>` or `@PurgeBot addroleid <roleID>`
- **Example:** `@PurgeBot addrole Admin` or `@PurgeBot addroleid 1274017921756172403`

### Remove Role

Revoke a role's permission to manage purge tasks. You can use either role name or role ID.

- **Usage:** `@PurgeBot removerole <roleName>` or `@PurgeBot removeroleid <roleID>`
- **Example:** `@PurgeBot removerole Admin` or `@PurgeBot removeroleid 1274017921756172403`

### List Permissions

Get a list of all users and roles registered to manage purge tasks, including their names.

- **Usage:** `@PurgeBot listpermissions`

### Help

Get detailed usage instructions and a list of available commands.

- **Usage:** `@PurgeBot help`

## ⚙️ Configuration

- **Purge Interval**: The interval at which the bot checks for messages to purge (default: 33 seconds).
- **Minimum Duration**: The minimum duration for purging tasks (default: 30 seconds).
- **Maximum Duration**: The maximum duration for purging tasks (default: 3333 days).

## 🗳️ Invite the Bot

To invite **PurgeBot** to your server, use the following invite link format:

`https://discord.com/oauth2/authorize?client_id=YOUR_APPLICATION_ID&scope=bot&permissions=75776`

**Required Permissions:**
- **Read Messages**
- **Send Messages**
- **Manage Messages** (for purging messages)
- **Read Message History**

Replace `YOUR_APPLICATION_ID` in the URL with your bot's actual application ID from the Discord Developer Portal.

## 📝 Example

Here's how you can use PurgeBot in your server:

1. **Start purging messages older than 1 hour:**

    ```markdown
    @PurgeBot 1h
    ```

2. **Stop purging in a channel:**

    ```markdown
    @PurgeBot stop
    ```

3. **Get a list of all purge tasks:**

    ```markdown
    @PurgeBot list
    ```

4. **Add a user to manage purge tasks:**

    ```markdown
    @PurgeBot adduser JohnDoe
    ```

5. **Remove a user from managing purge tasks:**

    ```markdown
    @PurgeBot removeuser JohnDoe
    ```

6. **Add a role to manage purge tasks:**

    ```markdown
    @PurgeBot addrole Admin
    ```

7. **Remove a role from managing purge tasks:**

    ```markdown
    @PurgeBot removerole Admin
    ```

8. **Get a list of all registered users and roles:**

    ```markdown
    @PurgeBot listpermissions
    ```

9. **Get help:**

    ```markdown
    @PurgeBot help
    ```

## 🙏 Acknowledgements

**PurgeBot** was inspired by the original [KMS Bot](https://github.com/internetisgone/kms-bot) project. The original bot, written in Python, provided the foundational concept for this Go implementation. A special thanks to the creator of that project!
