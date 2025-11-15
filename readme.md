# Snapfox – Simple Backup Scheduler

## 🚀 Installation (Linux)

Install Snapfox with a single command:

```sh
curl -fsSL https://raw.githubusercontent.com/Inspiractus01/snapfox/main/install.sh | sudo bash
```

This will:

- install Snapfox into `/usr/local/bin/snapfox`
- create the systemd service and timer (Linux only)
- enable automatic hourly backup checks
- generate the config directory `~/.snapfox`

---

## 🦊 What is Snapfox?

Snapfox is a lightweight and user-friendly backup scheduler powered by rsync.  
It allows you to easily create, edit, and run backup jobs — either manually or automatically in the background.

Snapfox works on:

- Linux (x86_64 & ARM64)
- macOS (Intel & Apple Silicon)
- Raspberry Pi

Configuration is stored inside one simple JSON file:

```
~/.snapfox/config.json
```

---

## ✨ Features

- Interactive terminal menu (TUI)
- Add, edit, and delete backup jobs
- Rsync-based efficient backups
- Automatic run scheduling via systemd timers (Linux only)
- Per-job retention settings (keep last X backups)
- Safe validation of source/destination paths
- Manual `run-all` and `run-due` commands
- Cross-platform binaries
- No dependencies except `rsync`

---

## 📘 Usage

### 🧭 Interactive Menu

Start Snapfox:

```sh
snapfox
```

Menu options:

- List backup jobs
- Add new job
- Edit existing job
- Delete job
- Run all backups
- Exit

---

### 🖥 CLI Commands

Run only due backups (used automatically by systemd timer):

```sh
snapfox run-due
```

Run all backups immediately:

```sh
snapfox run-all
```

Show the menu:

```sh
snapfox menu
```

---

## 🍏 macOS Installation

1. Download the correct binary from Releases:

- `snapfox-darwin-arm64` (Apple Silicon)
- `snapfox-darwin-amd64` (Intel)

2. Install manually:

```sh
chmod +x snapfox
sudo mv snapfox /usr/local/bin/
```

macOS does not use systemd — add this cron job if you want automation:

```sh
crontab -e
```

Example hourly run:

```
0 * * * * /usr/local/bin/snapfox run-due
```

---

## ⚙️ Config File

Config is stored here:

```
~/.snapfox/config.json
```

Example:

```json
{
  "jobs": [
    {
      "id": "job1",
      "name": "Documents Backup",
      "source": "/home/user/Documents",
      "destination": "/mnt/backup/documents",
      "interval_hours": 6,
      "retention": 5,
      "last_run": "2025-01-18T12:00:00Z"
    }
  ]
}
```

---

## 📁 Backup Output Structure

Each backup creates timestamped directories:

```
/mnt/backup/documents/
    2025-01-18_12-00-00/
    2025-01-18_06-00-00/
    2025-01-17_18-00-00/
```

Retention automatically removes older backups.

---

## 🧹 Uninstallation

### Linux

```sh
sudo systemctl disable --now snapfox.timer
sudo rm /etc/systemd/system/snapfox.service
sudo rm /etc/systemd/system/snapfox.timer
sudo rm /usr/local/bin/snapfox
rm -r ~/.snapfox
```

### macOS

```sh
sudo rm /usr/local/bin/snapfox
rm -r ~/.snapfox
```

---

## 🛣 Roadmap

- Remote backups (SSH)
- Status dashboard (web UI)
- Backup progress bar
- Encrypted backup archives
- Snapshot diff mode
- Notifications

---

## 📜 License

MIT License
