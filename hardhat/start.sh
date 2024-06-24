#!/bin/sh

# This script prepares and runs a standalone Hardhat node on your local machine, handling all dependency installation.

# Print a failure message to stderr and exit
fail() {
    MESSAGE=$1
    >&2 echo "\n**ERROR**\n$MESSAGE\n"
    exit 1
}

# Check if NVM is installed
if [ ! -d "${HOME}/.nvm/.git" ]; then
    echo "nvm not installed, running installer"
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash || fail "Error installing nvm"
fi

# Load NVM
. ~/.nvm/nvm.sh || fail "Error loading nvm"

# Check if Node v20 is installed
nvm which 20 > /dev/null 2>&1
if [ ! $? = 0 ]; then
    echo "Node 20 is not installed, installing it"
    nvm install 20 || fail "Error installing node v20"
fi

# Use Node v20
nvm use 20 || fail "Error setting node version to v20"

# Install dependencies and the hardhat standalone runtime
npm ci || fail "Error provisioning Hardhat"

# Run Hardhat
npx hardhat node --port 8545  &
HARDHAT_NODE_PID=$!
sleep 5 # Give the Hardhat node some time to start

# Deploy contracts
npx hardhat run scripts/deploy.js --network localhost || fail "Error deploying contracts"

# Wait for the Hardhat node process to end (optional)
wait $HARDHAT_NODE_PID