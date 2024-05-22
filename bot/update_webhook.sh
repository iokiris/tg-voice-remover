#!/bin/bash
 
cd /app

# Получение публичного URL ngrok
NGROK_API="http://ngrok:4040/api/tunnels"
NGROK_TUNNEL_URL=$(curl -s ${NGROK_API} | jq -r '.tunnels[0].public_url')
export NGROK_URL=${NGROK_TUNNEL_URL}

exec CompileDaemon -polling=true -log-prefix=true -build="go build -o ./tmp/myapp ./" -command="./tmp/myapp"
