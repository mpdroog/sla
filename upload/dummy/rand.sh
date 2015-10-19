#!/bin/bash
# 100MB
dd if=/dev/random of=100mb.bin bs=1000000 count=100
openssl rand -base64 32 > uuid.txt
