# fileserver

A simple file server to share files on a local network or on the internet using
[ngrok](https://ngrok.com/).

## Install

```
go install github.com/uri/fileserver@latest
```

## Usage

Running the program will share the current directory:

```
fileserver # host current directory on local network
fileserver -ngrok <NGROK_AUTHTOKEN> # hosts through ngrok
```

Use `-prefix` to set the network prefix (if your local network is different than
the default.)

```
Usage of fileserver:
  -log string
        log level: info|debug
  -ngrok string
        Ngrok API token
  -port string
        local port to serve (default "8080")
  -prefix string
        subnet to use for local server (default "192.168.0.0/16")
```
