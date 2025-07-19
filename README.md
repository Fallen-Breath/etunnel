# etunnel

**Still in early development, expect breaking changes**

[![Docker](https://img.shields.io/docker/v/fallenbreath/etunnel/latest)](https://hub.docker.com/r/fallenbreath/etunnel)

A secure encrypted tunnel. Encrypt and transfer your secret data on public network safely

Thanks [go-shadowsocks2](https://github.com/shadowsocks/go-shadowsocks2) for providing the tunnel encryption logic.
Notes that this tool only provides the tunneling functionality, it cannot be used as a proxy

## Usages

### CLI Mode

Helps

```bash
./etunnel --help
./etunnel client -h
./etunnel server --help
```

Quick Run using default values

```bash
# inbound -> (etunnel client) 0.0.0.0:8000 --[etunnel]-> (etunnel server) 1.2.3.4:12000 -> (target) 127.0.0.1:8888 
./etunnel server -l :12000 -t tcp://127.0.0.1:8888  # listen on :12000, tunnel outbound to :8888
./etunnel client -s 1.2.3.4:12000 -t tcp://:8000  # serve at 127.0.0.1:12000, tunnel inbound from :8000
```

More detailed run examples

```bash
# server
./etunnel server \
    -c AES-256-GCM -k my_secret_key  \
    -l :12000 \
    -t website:tcp://127.0.0.1:8000 \
    -t ssh:tcp://192.168.1.1:22
# client
./etunnel client \
    -c AES-256-GCM -k my_secret_key \
    -s 192.168.1.1:12000 \
    -t website:tcp://127.0.0.1:8888 \
    -t ssh:tcp://:2222
```

Using unix socket

```bash
./etunnel server -l 127.0.0.1:12000 -t unix:///var/run/docker.sock

# 127.0.0.1:2375 -> /var/run/docker.sock
./etunnel client -s 127.0.0.1:12000 -t tcp://127.0.0.1:2375
docker -H tcp://127.0.0.1:2375 image ls

# ./etunnel.sock -> /var/run/docker.sock
./etunnel client -s 127.0.0.1:12000 -t unix://etunnel.sock
docker -H unix://etunnel.sock image ls
```

Tunnel definition format: `[id:]protocol://address`, where:

- `id`: the identifier of the tunnel. 
  - Case-sensitive string, 1-255 characters long, composed of letters, numbers, hyphens, and underscores
  - Should be consistent between client and server. 
  - If not set, a default name based on the argument order will be used
- `protocol`: the protocol of payload to be carried by the tunnel. Can be `tcp`, `udp`, `unix`, `unixgram`
- `address`: a valid address depends on the mode and protocol
  - For client, it's the address to listen on
  - For server, it's the address to target on. It can be an address/hostname + port e.g. `127.0.0.1:22`, or just the port e.g. `:8000` and the host will be `127.0.0.1`
  - Address format:
    - `tcp`, `udp` tunnels: An address + port e.g. `127.0.0.1:2222`, or just the port e.g. `:2222`
    - `unix`, `unixgram` tunnels: Path to the unix socket file

### Config Mode

```bash
./etunnel -c etunnel_server.yml
./etunnel -c etunnel_client.yml
```

```yaml
# etunnel_server.yml
mode: server
listen: :12000
crypt: AES-256-GCM
key: my_secret_key
tunnels:
  website:
    protocol: tcp
    target: 127.0.0.1:8888
  ssh:
    protocol: tcp
    target: 192.168.1.1:22
  docker:
    protocol: unix
    target: /var/run/docker.sock
```

```yaml
# etunnel_client.yml
mode: client
server: 192.168.1.1:12000
crypt: AES-256-GCM
key: my_secret_key
tunnels:
  website:
    protocol: tcp
    listen: 127.0.0.1:8000
  ssh:
    protocol: tcp
    listen: 0.0.0.0:2222
  docker:
    protocol: tcp
    listen: 127.0.0.1:2375
```

You can append a `--export config.yml` to CLI mode commands to export the arguments into a config file

### Full help message view

TODO, just enter the following command

```bash
$ ./etunnel -h
$ ./etunnel server -h
$ ./etunnel client -h
$ ./etunnel tool -h
```

## Docker

Image available at [DockerHub](https://hub.docker.com/r/fallenbreath/etunnel)

`docker run fallenbreath/etunnel:master xxx` works just like `./etunnel xxx`

## Security

etunnel uses Authenticated Encryption with Associated Data (AEAD) encryption methods to encrypt the whole tunneled data

Supported encryption methods:

- `AES-128-GCM`
- `AES-256-GCM` (default)
- `CHACHA20-IETF-POLY1305`

## TODO

- [ ] hot-reload client / server
- [x] `unix` support for tunnel
- [ ] `udp` / `unixgram` support for tunnel
- [ ] `kcp` / `quic` support for client - server communication
- [ ] tcp proxy protocol support
