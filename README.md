# sol-saver-svc

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Saver service (oracle) to observe Solana bridge program deposit events.

## Configuration

To run that service you need a configuration file `config.yaml` with the following structure:

```yaml
log:
   disable_sentry: true
   level: debug
listener:
   addr: :8000
rpc:
   url: "" # solana node address
ws:
   url: "" # solana node address

listen:
   chain: Solana
   from_tx: ""
   program_id: ""
broadcaster:
   addr: "ip:8000" # broadcaster service address
   sender_account: "" # account used in the broadcaster service
core:
   addr: tcp://ip:26657 # your rarimo node address
cosmos:
   addr: "ip:9090" # your rarimo node address
subscriber:
   min_retry_period: 1s
   max_retry_period: 10s
```

Also, some environment variables is required to run:
```yaml
- name: KV_VIPER_FILE
  value: /config/config.yaml is the path to your config file
```

## Run

* Vote mode:
```shell
sol-saver-svc run voter
```

* Saver mode
```shell
sol-saver-svc run saver
```

* Full mode (saver + voter)
```shell
sol-saver-svc run service
```

* Catchup old transactions from Solana
```shell
sol-saver-svc run saver-catchup
```