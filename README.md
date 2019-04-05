# BWLog

A minimalistic bandwidth logger.

```
Usage of bwlog:
  -d string
        database path (default "./bwlog.sqlite")
  -i string
        interfaces to monitor, comma separated (default "eth0")
  -l string
        port to listen on (default "0.0.0.0:8080")
  -s int
        save to database every X seconds (default 60)
```

## Example

`bwlog -i eth0,docker0 -d ~/bwlog.sqlite`


## Compiling

`make`
