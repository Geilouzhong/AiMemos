## Docker

Build the local image:

```bash
docker build -f scripts/Dockerfile -t aimemos:local .
```

Run the container directly with Docker:

```bash
docker run -d \
  --restart unless-stopped \
  --name aimemos \
  -p 5230:5230 \
  -v ~/.memos:/var/opt/memos \
  -e MEMOS_ENABLE_MCP=true \
  -e MEMOS_MCP_PATH=/mcp \
  -e TZ=Asia/Shanghai \
  aimemos:local
```

Or use the provided Compose file:

```bash
docker compose -f scripts/compose.yaml up -d --build
```
