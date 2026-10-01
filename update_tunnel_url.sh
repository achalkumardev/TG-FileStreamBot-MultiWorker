#!/bin/bash
while true; do
  for port in 20242 20241 20243; do
    url=$(curl -s http://127.0.0.1:$port/metrics | grep -o 'https://[-a-z0-9]*\.trycloudflare\.com' | tail -n 1)
    if [ -n "$url" ]; then
      echo "$url" > /home/achal/TG-FileStreamBot/logs/current_url.txt
      break
    fi
  done
  sleep 3
done
