# Logging Configuration

The Avalanche Remote Signer supports comprehensive logging with multiple outputs and formats.

## Features

- **Console Output**: Real-time logs to stdout/stderr for immediate visibility
- **File Output**: Persistent logs to disk for auditing and troubleshooting
- **Multiple Formats**: Console (human-readable) or JSON (structured logging)
- **Configurable Levels**: debug, info, warn, error
- **Docker Integration**: Seamless log access from container to local filesystem

## Configuration

### Local Development

The `config.example.yaml` includes file logging:

```yaml
logging:
  level: "info"
  format: "console"
  file: "./logs/signer.log"
```

Logs will be written to `./logs/signer.log` on your local machine.

### Docker Deployment

The `config-docker.yaml` configures file logging to a mounted volume:

```yaml
logging:
  level: "info"
  format: "json"
  file: "/data/logs/signer.log"
```

The `docker-compose.yml` mounts the logs volume:

```yaml
volumes:
  - ./logs:/data/logs
```

This means:
- Container writes logs to `/data/logs/signer.log`
- Local filesystem can access them at `./logs/signer.log`

## Usage

### Starting the Signer

**Local (development):**
```bash
./avax-remote-signer --config config.yaml
```

Logs will appear in both:
- Console output (real-time)
- `./logs/signer.log` (persistent)

**Docker:**
```bash
docker-compose up -d
```

Logs will be written to `./logs/signer.log` automatically.

### Accessing Logs

**Local:**
```bash
# View logs in real-time
tail -f ./logs/signer.log

# View last 100 lines
tail -100 ./logs/signer.log

# Search for errors
grep "error" ./logs/signer.log

# Format JSON logs for readability (if using json format)
jq '.' ./logs/signer.log
```

**Docker (while running):**
```bash
# View current logs
docker-compose logs -f avax-remote-signer

# From local filesystem
tail -f ./logs/signer.log

# Docker exec into container
docker exec -it avax-remote-signer cat /data/logs/signer.log
```

### Cleaning Up Logs

```bash
# Clear log file
> ./logs/signer.log

# Or
rm ./logs/signer.log

# Docker will recreate it automatically
```

## Log Output Examples

### Console Format

```
2024-04-07T14:23:45.123Z	INFO	Starting Avalanche Remote Signer	{"backend": "file", "grpc_address": "0.0.0.0:9090", "http_address": "0.0.0.0:8080"}
2024-04-07T14:23:45.234Z	WARN	Using file-based backend - NOT RECOMMENDED FOR PRODUCTION
2024-04-07T14:23:45.345Z	INFO	Validator public key loaded	{"public_key": "b8cb2d0090..."}
2024-04-07T14:23:45.456Z	INFO	Starting gRPC server	{"address": "0.0.0.0:9090"}
```

### JSON Format

```json
{"level":"info","ts":1712502225.123,"caller":"main.go:60","msg":"Starting Avalanche Remote Signer","backend":"file","grpc_address":"0.0.0.0:9090","http_address":"0.0.0.0:8080"}
{"level":"warn","ts":1712502225.234,"caller":"main.go:67","msg":"Using file-based backend - NOT RECOMMENDED FOR PRODUCTION"}
{"level":"info","ts":1712502225.345,"caller":"main.go:75","msg":"Validator public key loaded","public_key":"b8cb2d0090..."}
```

## Log Levels

| Level | Usage | Example |
|-------|-------|---------|
| **debug** | Development, detailed troubleshooting | Key material info, internal state changes |
| **info** | Normal operation (default) | Server start, config loading|
| **warn** | Important notices | Production recommendations, deprecated usage |
| **error** | Operational failures | Connection failures, invalid requests |

## Best Practices

1. **Local Development**
   - Use `console` format for quick visibility
   - Use `debug` level for troubleshooting
   - Example: `config.yaml`

2. **Docker/Production**
   - Use `json` format for structured logging (easier parsing)
   - Use `info` level for normal operation
   - Mount logs volume to preserve logs after container stops
   - Example: `config-docker.yaml`

3. **Log Retention**
   ```bash
   # Compress old logs
   gzip ./logs/signer.log
   
   # Rotate logs (keep last 10 days)
   find ./logs -name "*.log" -mtime +10 -delete
   ```

4. **Monitoring**
   - Watch for ERROR level logs
   - Monitor log file size
   - Set up alerts for failed signing operations

## Environment Variable Overrides

Log level can be overridden via environment variable in docker-compose:

```yaml
environment:
  - AVAX_SIGNER_LOGGING_LEVEL=debug
  - AVAX_SIGNER_LOGGING_FILE=/data/logs/signer.log
```

## Troubleshooting

**Logs not appearing in file:**
- Verify log directory exists: `ls -la ./logs/`
- Check file permissions: `ls -la ./logs/signer.log`
- Verify config file path is correct

**Logs not appearing in Docker:**
- Verify volume mount: `docker inspect avax-remote-signer | grep -A 5 Mounts`
- Check container permissions: `docker exec avax-remote-signer ls -la /data/logs/`
- View Docker logs: `docker logs avax-remote-signer`

**File growing too large:**
- Reduce log level (change from debug to info)
- Use log rotation tool
- Archive old logs
