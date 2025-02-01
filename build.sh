#!/bin/bash
echo "------------start build-------------"
GOOS=linux GOARCH=amd64 go build -o awsAutoBackup
echo "------------end build-------------"