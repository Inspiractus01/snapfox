#!/usr/bin/env bash
set -e

REPO_OWNER="Inspiractus01"
REPO_NAME="snapfox"
BINARY_NAME="snapfox"
INSTALL_DIR="/usr/local/bin"
OS_NAME="$(uname -s)"
ARCH_NAME="$(uname -m)"

# User under which snapfox will run (~/.snapfox/config.json)
SNAPFOX_USER="${SUDO_USER:-$USER}"

echo "=== Snapfox installer ==="
echo "OS:       $OS_NAME"
echo "Arch:     $ARCH_NAME"
echo "Binary:   $BINARY_NAME"
echo "Install:  $INSTALL_DIR/$BINARY_NAME"
echo "User:     $SNAPFOX_USER"
echo ""

if [ "$(id -u)" -ne 0 ]; then
  echo "Please run this script as root, e.g.:"
  echo "  curl -fsSL https://raw.githubusercontent.com/$REPO_OWNER/$REPO_NAME/main/install.sh | sudo sh"
  exit 1
fi

# ---- select asset name based on OS + arch ----
ASSET=""

case "$OS_NAME" in
  Linux)
    case "$ARCH_NAME" in
      x86_64|amd64)
        ASSET="snapfox-linux-amd64"
        ;;
      aarch64|arm64)
        ASSET="snapfox-linux-arm64"
        ;;
      *)
        echo "Unsupported Linux arch: $ARCH_NAME"
        exit 1
        ;;
    esac
    ;;
  Darwin)
    case "$ARCH_NAME" in
      arm64)
        ASSET="snapfox-darwin-arm64"
        ;;
      x86_64)
        ASSET="snapfox-darwin-amd64"
        ;;
      *)
        echo "Unsupported macOS arch: $ARCH_NAME"
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Unsupported OS: $OS_NAME"
    exit 1
    ;;
esac

DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/latest/download/$ASSET"

echo "Downloading $ASSET from:"
echo "  $DOWNLOAD_URL"
echo ""

TMP_BIN="/tmp/$ASSET"
curl -fsSL -o "$TMP_BIN" "$DOWNLOAD_URL"

chmod +x "$TMP_BIN"

echo "Installing $BINARY_NAME to $INSTALL_DIR..."
install -m 755 "$TMP_BIN" "$INSTALL_DIR/$BINARY_NAME"

# ---------- Non-Linux: just install binary ----------
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

# ---------- Linux: rsync + systemd ----------
if ! command -v rsync >/dev/null 2>&1; then
  echo "rsync not found, installing via apt-get (if available)..."
  if command -v apt-get >/dev/null 2>&1; then
    apt-get update -y
    apt-get install -y rsync
  else
    echo "apt-get not found. Please install rsync manually."
  fi
fi

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