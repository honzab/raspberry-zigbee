#!/bin/bash

GOOS=linux GOARM=6 GOARCH=arm go build inkywhat.go && du -h inkywhat && file inkywhat
