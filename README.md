![banner](.github/banner.png)

## Overview

OP-FW Socket Storage provides the storage manager for the [op-fw socket](https://github.com/coalaura/admin-panel-socket). It is responsible for storing and retrieving player location data. It communicates via TCP, encrypted using AES, which allows it to run on a different server than the socket server itself.

## Requirements
- At least 2GB of memory (or swap)
- A decent bit of space depending on how many servers (for historic data)

## Configuration
Copy the `example.config.json` to `config.json` in same directory the storage server is running in. (If the config does not exist on startup, the storage server will create it with the default values).

```json
{
	"root": "./storage",
	"hostname": "0.0.0.0",
	"port": 4994,
	"allowed_ips": ["*"]
}
```

| Key           | Description                                                                | Default     |
| ------------- | -------------------------------------------------------------------------- | ----------- |
| `root`        | The storage root, this is where all data is stored.                        | `./storage` |
| `hostname`    | The hostname to bind to.                                                   | `0.0.0.0`   |
| `port`        | The port to bind to.                                                       | `4994`      |
| `allowed_ips` | A list of IP addresses that are allowed to connect. `*` will allow any IP. | `["*"]`     |

## Service
You can use the `panel_storage.service` file to start the storage server as a systemd service. Adjust the paths to match your environment.