# mqtt2tsdb

Bridges MQTT messages to a time series database.

## Running at startup (systemd)

Create `/etc/systemd/system/mqtt2tsdb.service`:

```ini
[Unit]
Description=mqtt2tsdb bridge
After=network.target mosquitto.service

[Service]
ExecStart=/home/janbrucek/mqtt2tsdb/mqtt2tsdb
Environment=CLOUDMQTT_URL=mqtt://localhost:1883/zigbee2mqtt/maindoor
Restart=always
RestartSec=5
User=janbrucek
WorkingDirectory=/home/janbrucek/mqtt2tsdb

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable mqtt2tsdb
sudo systemctl start mqtt2tsdb
```

## Logs

```bash
journalctl -u mqtt2tsdb        # all logs
journalctl -u mqtt2tsdb -f     # follow (tail)
journalctl -u mqtt2tsdb -n 50  # last 50 lines
```

To persist logs across reboots:

```bash
sudo mkdir -p /var/log/journal
sudo systemctl restart systemd-journald
```
