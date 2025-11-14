#!/bin/bash

# Script to run the Bybit trading bot in the background

# Build the bot
echo "Building the trading bot..."
go build -o bot cmd/bot/main.go

# Check if build was successful
if [ $? -ne 0 ]; then
    echo "Build failed. Please check for errors."
    exit 1
fi

# Run the bot in the background
echo "Starting the trading bot..."
nohup ./bot > bot.log 2>&1 &

echo "Trading bot started successfully!"
echo "Logs are being written to bot.log"
echo "Bot PID: $!"