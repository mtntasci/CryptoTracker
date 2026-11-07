#!/usr/bin/env zsh
set -euo pipefail

DOMAIN="https://seekmycrypto.com"
OUT_DIR="downloads"
APP_NAME="cryptotracker"
BUNDLE_NAME="CryptoTracker"
IDENTIFIER="com.seekmycrypto.cryptotracker"
VERSION="1.0.0"

rm -rf "${OUT_DIR}"
mkdir -p "${OUT_DIR}"

echo "========================================================"
echo "  Building SeekMyCrypto (CryptoTracker) Release Suite"
echo "  Target URL: ${DOMAIN}/${OUT_DIR}/"
echo "========================================================"

# ------------------------------------------------------------------------------
# 1. Generate 1024x1024 macOS App Icon & Compile AppIcon.icns
# ------------------------------------------------------------------------------
echo "[1/5] Generating macOS AppIcon.icns..."
WORK_DIR="$(mktemp -d)"
ICONSET_DIR="${WORK_DIR}/AppIcon.iconset"
mkdir -p "${ICONSET_DIR}"

cat << 'EOF' > "${WORK_DIR}/gen_icon.go"
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func main() {
	const size = 1024
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	// Draw macOS Squircle with dark cyber gradient + neon red/purple border
	cx, cy := float64(size)/2.0, float64(size)/2.0
	radius := 440.0

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := math.Abs(float64(x)-cx) / radius
			dy := math.Abs(float64(y)-cy) / radius
			// Superellipse (squircle) exponent n=4.5
			dist := math.Pow(math.Pow(dx, 4.5)+math.Pow(dy, 4.5), 1.0/4.5)

			if dist > 1.0 {
				continue
			}

			// Background gradient: Dark Obsidian (#0B0E17) to Deep Cyber Purple (#1A0B2E)
			t := float64(y) / float64(size)
			r := uint8(11 + t*18)
			g := uint8(14 - t*4)
			b := uint8(23 + t*28)

			// Outer neon border ring (Red #FF1744 -> Purple #AA00FF)
			if dist > 0.92 && dist <= 0.97 {
				r = uint8(255 - t*85)
				g = uint8(23*(1-t))
				b = uint8(68 + t*187)
			}

			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// Helper to draw thick anti-aliased glowing line
	drawLine := func(x0, y0, x1, y1, thickness float64, col color.NRGBA) {
		minX := int(math.Max(0, math.Min(x0, x1)-thickness*2))
		maxX := int(math.Min(size-1, math.Max(x0, x1)+thickness*2))
		minY := int(math.Max(0, math.Min(y0, y1)-thickness*2))
		maxY := int(math.Min(size-1, math.Max(y0, y1)+thickness*2))

		vx, vy := x1-x0, y1-y0
		lenSq := vx*vx + vy*vy

		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				t := 0.0
				if lenSq > 0 {
					t = ((float64(x)-x0)*vx + (float64(y)-y0)*vy) / lenSq
					if t < 0 {
						t = 0
					} else if t > 1 {
						t = 1
					}
				}
				px, py := x0+t*vx, y0+t*vy
				d := math.Hypot(float64(x)-px, float64(y)-py)
				if d <= thickness {
					img.SetNRGBA(x, y, col)
				}
			}
		}
	}

	// Helper to draw filled circle node
	drawNode := func(nx, ny, rOuter, rInner float64, outerCol, innerCol color.NRGBA) {
		for y := int(ny - rOuter - 2); y <= int(ny + rOuter + 2); y++ {
			for x := int(nx - rOuter - 2); x <= int(nx + rOuter + 2); x++ {
				d := math.Hypot(float64(x)-nx, float64(y)-ny)
				if d <= rInner {
					img.SetNRGBA(x, y, innerCol)
				} else if d <= rOuter {
					img.SetNRGBA(x, y, outerCol)
				}
			}
		}
	}

	red := color.NRGBA{R: 255, G: 23, B: 68, A: 255}
	purple := color.NRGBA{R: 170, G: 0, B: 255, A: 255}
	matrix := color.NRGBA{R: 0, G: 255, B: 65, A: 255}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	dark := color.NRGBA{R: 14, G: 17, B: 26, A: 255}

	// Draw giant stylized "C" (Crimson Red) on the left
	drawLine(250, 310, 450, 310, 34, red)
	drawLine(250, 310, 250, 610, 34, red)
	drawLine(250, 610, 450, 610, 34, red)

	// Draw giant stylized "T" (Crimson Red -> Matrix Green accent) on the right
	drawLine(540, 310, 790, 310, 34, red)
	drawLine(665, 310, 665, 610, 34, red)

	// Draw bottom multi-stage trace graph bar (Stage 1 Red -> Stage 2 Purple -> Stage 3 Matrix Green)
	drawLine(270, 740, 512, 740, 14, purple)
	drawLine(512, 740, 754, 740, 14, matrix)

	drawNode(270, 740, 36, 18, red, white)
	drawNode(512, 740, 36, 18, purple, dark)
	drawNode(754, 740, 42, 22, matrix, white)

	f, err := os.Create(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
EOF

MASTER_PNG="${WORK_DIR}/icon_1024.png"
go run "${WORK_DIR}/gen_icon.go" "${MASTER_PNG}"

for sz in 16 32 128 256 512; do
    sips -z ${sz} ${sz} "${MASTER_PNG}" --out "${ICONSET_DIR}/icon_${sz}x${sz}.png" >/dev/null
    sz2=$((sz * 2))
    sips -z ${sz2} ${sz2} "${MASTER_PNG}" --out "${ICONSET_DIR}/icon_${sz}x${sz}@2x.png" >/dev/null
done

ICNS_PATH="${WORK_DIR}/AppIcon.icns"
iconutil -c icns "${ICONSET_DIR}" -o "${ICNS_PATH}"
cp "${MASTER_PNG}" "${OUT_DIR}/cryptotracker-icon.png"

# ------------------------------------------------------------------------------
# 2. Build macOS Binaries (arm64, amd64, and Single Universal Binary)
# ------------------------------------------------------------------------------
echo "[2/5] Building macOS Universal Binary (Apple Silicon + Intel in one file)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${OUT_DIR}/${APP_NAME}-darwin-arm64" ./cmd/crypto-tracker
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${OUT_DIR}/${APP_NAME}-darwin-amd64" ./cmd/crypto-tracker

# Create Universal Binary so macOS users don't need architecture detection in curl
lipo -create -output "${OUT_DIR}/${APP_NAME}-mac" \
    "${OUT_DIR}/${APP_NAME}-darwin-arm64" \
    "${OUT_DIR}/${APP_NAME}-darwin-amd64"
chmod +x "${OUT_DIR}/${APP_NAME}-mac"

# ------------------------------------------------------------------------------
# 3. Build macOS CryptoTracker.app Bundle + CryptoTracker-Installer.pkg
# ------------------------------------------------------------------------------
echo "[3/5] Building macOS Application Bundle (CryptoTracker.app) & Installer (.pkg)..."
PKG_ROOT="${WORK_DIR}/pkgroot"
APP_BUNDLE="${PKG_ROOT}/Applications/${BUNDLE_NAME}.app"
mkdir -p "${APP_BUNDLE}/Contents/MacOS"
mkdir -p "${APP_BUNDLE}/Contents/Resources"
mkdir -p "${PKG_ROOT}/usr/local/bin"

# Install CLI binary into /usr/local/bin/cryptotracker
cp "${OUT_DIR}/${APP_NAME}-mac" "${PKG_ROOT}/usr/local/bin/${APP_NAME}"
chmod 755 "${PKG_ROOT}/usr/local/bin/${APP_NAME}"

# Copy Universal binary and Icon into CryptoTracker.app
cp "${OUT_DIR}/${APP_NAME}-mac" "${APP_BUNDLE}/Contents/MacOS/${APP_NAME}-bin"
chmod 755 "${APP_BUNDLE}/Contents/MacOS/${APP_NAME}-bin"
cp "${ICNS_PATH}" "${APP_BUNDLE}/Contents/Resources/AppIcon.icns"

# Create macOS App Launcher that opens Terminal and runs cryptotracker when clicked in Applications/Launchpad
cat << 'EOF' > "${APP_BUNDLE}/Contents/MacOS/CryptoTracker"
#!/bin/sh
BIN_DIR="$(cd "$(dirname "$0")" && pwd)"
osascript -e "tell application \"Terminal\"
    activate
    do script \"exec '${BIN_DIR}/cryptotracker-bin'\"
end tell"
EOF
chmod 755 "${APP_BUNDLE}/Contents/MacOS/CryptoTracker"

cat << EOF > "${APP_BUNDLE}/Contents/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleDisplayName</key>
    <string>CryptoTracker</string>
    <key>CFBundleExecutable</key>
    <string>CryptoTracker</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>${IDENTIFIER}</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>CryptoTracker</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>LSMinimumSystemVersion</key>
    <string>12.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

pkgbuild \
    --root "${PKG_ROOT}" \
    --identifier "${IDENTIFIER}" \
    --version "${VERSION}" \
    --install-location "/" \
    "${OUT_DIR}/CryptoTracker-Installer.pkg" >/dev/null

rm -rf "${WORK_DIR}"

# ------------------------------------------------------------------------------
# 4. Build Linux and Windows Binaries
# ------------------------------------------------------------------------------
echo "[4/5] Building Linux and Windows Binaries..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${OUT_DIR}/${APP_NAME}-linux" ./cmd/crypto-tracker
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o "${OUT_DIR}/${APP_NAME}-linux-arm64" ./cmd/crypto-tracker
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "${OUT_DIR}/${APP_NAME}-windows.exe" ./cmd/crypto-tracker
CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o "${OUT_DIR}/${APP_NAME}-windows-arm64.exe" ./cmd/crypto-tracker

# ------------------------------------------------------------------------------
# 5. Create One-Command Remote Install Scripts for seekmycrypto.com/downloads/
# ------------------------------------------------------------------------------
echo "[5/5] Creating remote installer scripts (install.sh & install.ps1) and checksums..."

cat << 'EOF' > "${OUT_DIR}/install.sh"
#!/bin/sh
set -e

BASE_URL="https://seekmycrypto.com/downloads"
OS="$(uname -s)"

echo "Installing CryptoTracker from ${BASE_URL}..."

if [ "${OS}" = "Darwin" ]; then
    echo "-> Downloading macOS Installer Package (CryptoTracker.app + cryptotracker CLI)..."
    TMP_PKG="/tmp/CryptoTracker-Installer.pkg"
    curl -fsSL "${BASE_URL}/CryptoTracker-Installer.pkg" -o "${TMP_PKG}"
    sudo installer -pkg "${TMP_PKG}" -target /
    rm -f "${TMP_PKG}"
    echo "✔ Installed /Applications/CryptoTracker.app and /usr/local/bin/cryptotracker"
    echo "Run 'cryptotracker' in Terminal or open CryptoTracker from Applications!"
elif [ "${OS}" = "Linux" ]; then
    ARCH="$(uname -m)"
    BIN_FILE="cryptotracker-linux"
    if [ "${ARCH}" = "aarch64" ] || [ "${ARCH}" = "arm64" ]; then
        BIN_FILE="cryptotracker-linux-arm64"
    fi
    sudo curl -fsSL "${BASE_URL}/${BIN_FILE}" -o /usr/local/bin/cryptotracker
    sudo chmod +x /usr/local/bin/cryptotracker
    echo "✔ Installed /usr/local/bin/cryptotracker"
    echo "Run 'cryptotracker' from any terminal!"
else
    echo "Unsupported OS: ${OS}"
    exit 1
fi
EOF
chmod +x "${OUT_DIR}/install.sh"

cat << 'EOF' > "${OUT_DIR}/install.ps1"
$ErrorActionPreference = "Stop"
$BaseUrl = "https://seekmycrypto.com/downloads"
$InstallDir = "$env:LOCALAPPDATA\Programs\CryptoTracker"

Write-Host "Installing CryptoTracker from $BaseUrl..." -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Invoke-WebRequest -Uri "$BaseUrl/cryptotracker-windows.exe" -OutFile "$InstallDir\cryptotracker.exe"

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
}

Write-Host "✔ CryptoTracker installed successfully!" -ForegroundColor Green
Write-Host "Type 'cryptotracker' in PowerShell or Command Prompt to launch." -ForegroundColor White
EOF

(cd "${OUT_DIR}" && shasum -a 256 CryptoTracker-Installer.pkg ${APP_NAME}-* > SHA256SUMS.txt)

echo ""
echo "✔ All files generated inside ./${OUT_DIR}/:"
ls -lh "${OUT_DIR}"
