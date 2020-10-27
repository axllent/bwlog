# BWLog - Lightweight bandwidth logger for *nix

A lightweight bandwidth logger written in Go. The tool logs the incoming and outgoing network
traffic from each of the specified network interfaces, and provides a web frontend to view
both a live graph and statistics history for each interface.

![BWLog Screenshot](screenshot.png "BWLog Screenshot")


## Usage options

```shell
BWLog: A lightweight bandwidth logger

Usage:
  bwlog -i eth0 -d ~/bwlog/ [flags]
  bwlog [command]

Available Commands:
  update      Update bwlog to the latest version
  version     Display the app version & update information

Flags:
  -d, --database string     Database directory to save CSV files (default "./")
  -i, --interfaces string   Interfaces to monitor, comma separated eg: eth0,eth1
  -l, --listen string       Interface & port to listen on (default "0.0.0.0:8080")
  -p, --password string     Auth password file (must contain a single "<user> <pass>")
  -s, --save string         How often to save the database to disk. Examples: 30s, 5m, 1h (default "60s")
      --sslcert string      SSL certificate (must be used together with --sslkey)
      --sslkey string       SSL key (must be used together with --sslcert)
```

## Installing

Download a suitable binary for your system from the [releases](https://github.com/axllent/bwlog/releases) page.


## Running BWLog

```shell
bwlog -i eth0 -d ~/bwlog/
```

See `bwlog -h` for options.

Unless you have specified different listening options, you should be able to connect to `127.0.0.1:8080`
with your web browser.


## Basic auth

If you want to use basic auth, simply create a file with two words in it, your username and password, eg:
```
MyUser MySecretPass
```
Then just add `-p <password_file>` to your startup flags. BWLog does not handle multiple users/passwords.


## HTTPS

To enable HTTPS you must use both the `--sslcert` and `--sslkey` options to specify the respective certificate files.


## Compiling from source

Ensure you have `go` and `make` installed, then just:

```shell
make
```


## Integrate with systemd

BWLog does not have a background daemon. If you want bwlog to run automatically in the background then you can
easily integrate it with systemd.

Create a file `/etc/systemd/system/bwlog.service`, ensuring sure you modify the  `ExecStart` to your requirements.

```
[Unit]
Description=BWLog

[Service]
ExecStart=/usr/local/bin/bwlog -d /opt/bwlog/ -i eth0,eth1
Restart=always
RestartSec=10
# Output to syslog
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=bwlog

[Install]
WantedBy=multi-user.target
```

Enable the service: `systemctl enable bwlog.service`

Start the service: `systemctl start bwlog.service`

If you make changes to `/etc/systemd/system/bwlog.service` you will need to `systemctl daemon-reload`
before restarting the service.


## TODOs

There are some other things I'd like to do at some stage if I ever get inspired and have some time:

- Switch to vue.js
