#!/bin/sh
set -e
rm -rf manpages
mkdir manpages
go run ./cmd/dinnerclub/ man | gzip -c -9 >manpages/dinnerclub.1.gz
