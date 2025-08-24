#!/bin/sh

~/go/bin/dlv debug --listen=localhost:5005 --headless=true ./main.go
