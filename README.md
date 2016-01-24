### Escarole

[![Build Status](https://travis-ci.org/albertrdixon/escarole.svg)](https://travis-ci.org/albertrdixon/escarole)

Run your git based projects in containers and keep 'em fresh!

This is *SUPER* dirty. Just threw this together to deal with updating things like [Sickrage](sickrage.github.io) and [Couchpotato](couchpota.to) in docker containers. It is a little cleaner after a rewrite, but still fairly limited.

## Usage

First create a config yaml. The config yaml only specifies the cammand use to run this project. Environment variables are expanded using the container's environment plus `APP_HOME` which is the path to the project clone:

```yaml
cmd: python ${APP_HOME}/my_script.py --flag ${SOME_VAR}
```

Put this in your container, default path is `/escarole.yml`

Now just run it. No big deal.

```
usage: escarole [<flags>] <project> [<name>]

Keeps your app leafy fresh!

Flags:
  --help            
        Show context-sensitive help

  -C, --config=/escarole.yml
        path to command config

  -b, --branch=BRANCH  
        branch to use

  -u, --update-interval=24h
        app update interval. Must be able to be parsed by time.ParseDuration

  --uid=0              
        app uid

  --gid=0              
        app gid

  -e, --env=key=value  
        app env vars. Note: if give, these will be the only environment variables available to the app.

  -l, --log-level={debug,info,warn,error,fatal}
        log level.

Args:
  <project>  
        github project. Format: Organization/Project, e.g. albertrdixon/escarole

  [<name>]  
        app name. If not given will use lowercase project name, e.g. Org/MyProject -> myproject
```
