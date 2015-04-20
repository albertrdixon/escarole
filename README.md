### Escarole

[![Build Status](https://travis-ci.org/albertrdixon/escarole.svg)](https://travis-ci.org/albertrdixon/escarole)

This is *SUPER* dirty. Just threw this together to deal with updating things like Sickrage in docker containers.

## Usage

Create a config yaml (by default this is `/etc/escarole.yaml`):

```yaml
---
    name: <process_name>
    directory: /path/to/cloned/repo
    command: <restart_command> # e.g. supervisorctl restart sickrage
```

Run it! Help is below:

```
usage: escarole [<flags>]

Flags:
  --help              Show help.
  -d, --debug         Enable debug output.
  -i, --interval=30m  Set update interval. Must be parseable by time.ParseTime (e.g. 20m, 2h, etc.).
  -C, --conf=/etc/escarole.yaml
                      Escarole config. Must be a real file.
  --version           Show application version.
```
