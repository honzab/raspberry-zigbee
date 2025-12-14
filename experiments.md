# Experiments

## Pairing the Parasoll

Open the web interface of Zigbee2MQTT and click the `Permit join` button. 
Then press the pairing button on the Parasoll four times. The LED should start blinking.
For me it took a few tries to get it to pairing mode, not sure why.

## Debounce

For some reason, the messages are repeated on the topic. I got two per every state change.
I've added a debounce to the Parasoll in the `configuration.yaml` file:

```yaml
devices:
  'YOURDEVICEHASH':
    friendly_name: maindoor
    debounce: 1
    homeassistant: {}
    optimistic: true
```

## Quick bash script to announce the state of a door

```bash
mosquitto_sub -v -h 192.168.1.102 -t zigbee2mqtt/maindoor | 
  while read payload; do 
      opp=$(echo ${payload} | sed 's/zigbee2mqtt\/maindoor//' | jq .contact); 
      if [[ ${opp} = "true" ]]; then 
        say -v Klara "stängd"; else 
        say -v Klara "öppet"; 
      fi; 
  done
```

## Quick'n'dirty solution to push to Victoria-metrics

```bash
mosquitto_sub -v -h 192.168.1.102 -t zigbee2mqtt/maindoor | 
  while read payload; do 
      closed=$(echo ${payload} | sed 's/zigbee2mqtt\/maindoor//' | jq .contact)
      battery=$(echo ${payload} | sed 's/zigbee2mqtt\/maindoor//' | jq .battery)
      name="maindoor"
      curl -H 'Content-Type: application/json' -d "{\"metric\":\"door_closed\",\"value\": \"${closed}\",\"tags\":{\"sensor\": \"${name}\"}}' http://192.168.1.102:4242/api/put
  done
```

## Where do I store all of it?

```bash
sudo apt install victoria-metrics
```

Activate the Open TSDB HTTP api in the service file `/lib/systemd/system/victoria-metrics.service`:
```
ExecStart=/usr/bin/victoria-metrics -opentsdbHTTPListenAddr=:4242 $ARGS`
```

Test from another computer

```bash
curl -H 'Content-Type: application/json' -d '{"metric":"x.y.z","value":45.34,"tags":{"t1":"v1","t2":"v2"}}' \
  http://192.168.1.102:4242/api/put
```
