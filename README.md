promprobe
=========

*Prometheus Probe*

```
$ ./promprobe --help
Prometheus Prober

Usage:
  promprobe [flags]
  promprobe [command]

Available Commands:
  cpu
  help        Help about any command
  memory

Flags:
      --config string   config file
  -h, --help            help for promprobe

Use "promprobe [command] --help" for more information about a command.
```

Compiling:

```
$ git clone https://github.com/honza/promprobe
$ cd promprobe
$ go build -o promprobe main.go
$ # Create a config.yaml
$ ./promprobe memory --config config.yaml
```

config.yaml:

``` yaml
---
token: sha256~...
host: https://console-openshift-console.apps.ostest.test.metalkube.org
pod: some-pod
containers:
  - some-container-1
  - some-container-2
...
```

Sample output:

```
|           CONTAINER           |     VALUE     |   MB    |
|-------------------------------|---------------|---------|
| some-container-1              |       4395008 |    4.19 |
| some-container-2              |       4546560 |    4.34 |
```
