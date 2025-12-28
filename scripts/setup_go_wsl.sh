#!/bin/bash
# FitStack Payments - Go 1.22+ Installation Script for WSL
# Run this script in WSL (Ubuntu) to setup Go environment

set -e

GO_VERSION="1.22.5"
GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="https://go.dev/dl/${GO_TAR}"

echo "=== FitStack Payments - Go Setup for WSL ==="
echo ""

# Check if Go is already installed
if command -v go &> /dev/null; then
    CURRENT_VERSION=$(go version | awk '{print $3}')
    echo "Go is already installed: $CURRENT_VERSION"
    read -p "Do you want to reinstall/update? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Keeping existing Go installation."
        exit 0
    fi
fi

echo "1. Updating system packages..."
sudo apt-get update -qq

echo "2. Installing dependencies..."
sudo apt-get install -y -qq wget curl git

echo "3. Downloading Go ${GO_VERSION}..."
cd /tmp
wget -q "${GO_URL}" -O "${GO_TAR}"

echo "4. Removing old Go installation (if exists)..."
sudo rm -rf /usr/local/go

echo "5. Extracting Go to /usr/local..."
sudo tar -C /usr/local -xzf "${GO_TAR}"

echo "6. Cleaning up..."
rm -f "${GO_TAR}"

echo "7. Configuring environment variables..."

# Add Go to PATH in .bashrc if not already present
BASHRC="$HOME/.bashrc"
GO_PATH_EXPORT='export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin'
GO_PATH_LINE='export GOPATH=$HOME/go'

if ! grep -q "/usr/local/go/bin" "$BASHRC"; then
    echo "" >> "$BASHRC"
    echo "# Go environment variables (added by fitstack-payments setup)" >> "$BASHRC"
    echo "$GO_PATH_LINE" >> "$BASHRC"
    echo "$GO_PATH_EXPORT" >> "$BASHRC"
    echo "Added Go environment variables to $BASHRC"
else
    echo "Go PATH already configured in $BASHRC"
fi

# Create GOPATH directory
mkdir -p "$HOME/go/bin"
mkdir -p "$HOME/go/src"
mkdir -p "$HOME/go/pkg"

echo ""
echo "=== Installation Complete ==="
echo ""
echo "Go version installed:"
/usr/local/go/bin/go version
echo ""
echo "IMPORTANT: Run the following command to apply changes:"
echo "  source ~/.bashrc"
echo ""
echo "Or open a new terminal window."
echo ""
echo "To verify installation, run:"
echo "  go version"
echo ""
echo "Next steps:"
echo "  cd /mnt/h/fitstack_payments"
echo "  go mod tidy"
echo "  go build ./..."
