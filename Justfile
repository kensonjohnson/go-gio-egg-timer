set quiet := true

MAIN_PACKAGE_PATH := "."
BINARY_NAME := "egg-timer"

[private]
help:
    just --list --unsorted

# Run dev build
run:
  go run .

# Fetch and organize dependencies
tidy:
  go mod tidy

# Build binary for current OS/Arch
build:
  go build -o=./build/{{ BINARY_NAME }} {{ MAIN_PACKAGE_PATH }}
