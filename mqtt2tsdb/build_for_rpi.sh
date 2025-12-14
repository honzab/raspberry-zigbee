#!/bin/bash

GOOS=linux GOARM=6 GOARCH=arm go build mqtt2tsdb.go && du -h mqtt2tsdb && file mqtt2tsdb
