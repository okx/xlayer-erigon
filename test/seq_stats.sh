#!/bin/bash

while true; do
  echo "$(date)" >> seq_stats.txt
  docker stats xlayer-seq --no-stream >> seq_stats.txt
  sleep 5  # 每5秒记录一次
done
