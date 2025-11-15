#!/usr/bin/env bash
set -e

BINARY_NAME="snapfox"
INSTALL_DIR="/usr/local/bin"
OS_NAME="$(uname -s)"

# User, under which snapfox will run (and where ~/.snapfox/config.json is)
SNAPFOX_USER="${SUDO_USER:-$USER}"

echo "=== Snapfox installer ==="
echo "OS:       $OS_NAME"
echo "Binary:   $BINARY_NAME"
echo "Install:  $INSTALL_DIR/$BINARY_NAME"
echo "User:     $SNAPFOX_USER"
echo ""

if [ "$EUID" -ne 0 ]; then
  echo "Please run this script as root, e.g.:"
  echo "  sudo ./install.sh"
  exit 1
fi

# check binary
if [ ! -f "./$BINARY_NAME" ]; then
  echo "Error: ./$BINARY_NAME not found."
  echo "Build it first:"
  echo "  go build -o snapfox"
  exit 1
fi

# install rsync only on Linux with apt (Debian/Ubuntu)
if [ "$OS_NAME" = "Linux" ]; then
  if ! command -v rsync >/dev/null 2>&1; then
    echo "rsync not found, installing via apt-get..."
    if command -v apt-get >/dev/null 2>&1; then
      apt-get update -y
      apt-get install -y rsync
    else
      echo "apt-get not found. Please install rsync manually."
    fi
  fi
fi

echo "Installing $BINARY_NAME to $INSTALL_DIR..."
install -m 755 "./$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

# ----- non-linux: just install binary and exit -----
if [ "$OS_NAME" != "Linux" ]; then
  echo ""
  echo "Non-Linux OS detected ($OS_NAME)."
  echo "Binary installed to: $INSTALL_DIR/$BINARY_NAME"
  echo ""
  echo "Automatic scheduling with systemd is only supported on Linux."
  echo "You can run Snapfox manually, e.g.:"
  echo "  snapfox ui"
  echo "  snapfox run-due"
  exit 0
fi

# ----- Linux + systemd timer -----
SERVICE_PATH="/etc/systemd/system/snapfox.service"
TIMER_PATH="/etc/systemd/system/snapfox.timer"

echo "Creating $SERVICE_PATH ..."
cat > "$SERVICE_PATH" <<EOF
[Unit]
Description=Snapfox backup runner
After=network-online.target

[Service]
Type=oneshot
User=$SNAPFOX_USER
WorkingDirectory=/home/$SNAPFOX_USER
ExecStart=$INSTALL_DIR/$BINARY_NAME run-due
Nice=10
EOF

echo "Creating $TIMER_PATH ..."
cat > "$TIMER_PATH" <<EOF
[Unit]
Description=Run Snapfox backups periodically

[Timer]
OnBootSec=10s
OnUnitActiveSec=10s
Unit=snapfox.service

[Install]
WantedBy=timers.target
EOF

echo "Reloading systemd..."
systemctl daemon-reload

echo "Enabling and starting snapfox.timer..."
systemctl enable --now snapfox.timer

echo ""
echo "=== Done (Linux + systemd) ==="
echo "Now, as user $SNAPFOX_USER, run:"
echo "  snapfox ui"
echo "to configure your backup jobs."
echo ""
echo "Check timer status:"
echo "  systemctl status snapfox.timer"
echo "Check run logs:"
echo "  journalctl -u snapfox.service -n 50"