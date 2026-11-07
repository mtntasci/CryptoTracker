#!/usr/bin/env zsh
set -euo pipefail

APP_NAME="cryptotracker"
PKG_NAME="CryptoTracker-Installer.pkg"
IDENTIFIER="com.cryptotracker.cli"
VERSION="1.0.0"

echo "========================================================"
echo "  CRYPTO TRACKER - macOS Installer & Package Builder"
echo "========================================================"

# 1. Build optimized native binary
echo "[1/3] Compiling optimized ${APP_NAME} binary for $(uname -m)..."
go build -ldflags="-s -w" -o "${APP_NAME}" ./cmd/crypto-tracker
chmod +x "${APP_NAME}"

# 2. Build macOS .pkg installer if pkgbuild is available
if command -v pkgbuild >/dev/null 2>&1; then
    echo "[2/3] Generating native macOS installer (${PKG_NAME})..."
    STAGE_DIR="$(mktemp -d)"
    mkdir -p "${STAGE_DIR}/usr/local/bin"
    cp "${APP_NAME}" "${STAGE_DIR}/usr/local/bin/${APP_NAME}"
    chmod 755 "${STAGE_DIR}/usr/local/bin/${APP_NAME}"

    pkgbuild \
        --root "${STAGE_DIR}" \
        --identifier "${IDENTIFIER}" \
        --version "${VERSION}" \
        --install-location "/" \
        "${PKG_NAME}" >/dev/null

    rm -rf "${STAGE_DIR}"
    echo "      ✔ Created macOS GUI Installer: ./${PKG_NAME}"
fi

# 3. Install directly into system PATH so 'cryptotracker' works from anywhere
echo "[3/3] Installing '${APP_NAME}' into global PATH..."
INSTALLED_PATH=""

if [ -d "/opt/homebrew/bin" ] && [ -w "/opt/homebrew/bin" ]; then
    cp "${APP_NAME}" "/opt/homebrew/bin/${APP_NAME}"
    chmod 755 "/opt/homebrew/bin/${APP_NAME}"
    INSTALLED_PATH="/opt/homebrew/bin/${APP_NAME}"
elif [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
    cp "${APP_NAME}" "/usr/local/bin/${APP_NAME}"
    chmod 755 "/usr/local/bin/${APP_NAME}"
    INSTALLED_PATH="/usr/local/bin/${APP_NAME}"
elif [ -d "${HOME}/.local/bin" ] && [ -w "${HOME}/.local/bin" ]; then
    cp "${APP_NAME}" "${HOME}/.local/bin/${APP_NAME}"
    chmod 755 "${HOME}/.local/bin/${APP_NAME}"
    INSTALLED_PATH="${HOME}/.local/bin/${APP_NAME}"
else
    echo "      Requesting administrator privileges to install to /usr/local/bin..."
    sudo mkdir -p /usr/local/bin
    sudo cp "${APP_NAME}" "/usr/local/bin/${APP_NAME}"
    sudo chmod 755 "/usr/local/bin/${APP_NAME}"
    INSTALLED_PATH="/usr/local/bin/${APP_NAME}"
fi

echo ""
echo "✔ Installation Complete!"
echo "  Binary Installed : ${INSTALLED_PATH}"
echo "  macOS Installer  : $(pwd)/${PKG_NAME}"
echo ""
echo "You can now run '${APP_NAME}' from any terminal directory:"
echo "  $ ${APP_NAME}"
echo "========================================================"

