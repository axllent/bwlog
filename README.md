# BWLog - Lightweight bandwidth logger for *nix

A lightweight bandwidth logger written in Go. The tool logs the incoming and outgoing network traffic from each
of the set network interfaces to a sqlite database, and provides a web frontend to view both a live graph and
statistics history for each monitored interface.

![BWLog Screenshot](screenshot.png "BWLog Screenshot")


## Usage options

```shell
Options:
  -d string
    	database path (default "./bwlog.sqlite")
  -i string
    	interfaces to monitor, comma separated (eg: "eth0")
  -l string
    	port to listen on (default "0.0.0.0:8080")
  -s int
    	save to database every X seconds (default 60)
  -u	update to latest release
  -v	show version number
```


## Running BWLog

```shell
bwlog -i eth0,docker0 -d ~/bwlog.sqlite
```

See `bwlog -h` for options.

If you wish to just run the code without building a binary, the wrapper `run.sh` can make this easy.
Note that you need `golang` & `gcc` installed for this.


```shell
./run.sh -i eth0,docker0 -d ~/bwlog.sqlite
```

Unless you have specified different listening options, you should be able to connect to `<server-ip>:8080`
with your web browser.


## Compiling

Ensure you have `golang`, `gcc` (for go-sqlite3) and `make` installed, then just:

```shell
make
```


### Cross compiling

I haven't had much luck cross-compiling as `mattn/go-sqlite3` is a CGO enabled package, so requires a valid `gcc`
compiler for that required platform/architecture installed. I'm sure there are ways of doing it, but I gave up.


## Integrate with systemd

BWLog does not have a background daemon. If you want bwlog to run automatically in the background then you can
easily integrate this with systemd.

Create a file `/etc/systemd/system/bwlog.service`, ensuring sure you modify the  `ExecStart` to your requirements.

```
[Unit]
Description=BWLog

[Service]
ExecStart=/usr/local/bin/bwlog -d /opt/bwlog/bwlog.sqlite -i eth0,eth1
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
- HTTP compression (gzip), possibly minify css/js
- Optional basic auth
