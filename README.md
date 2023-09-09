# etunnel

**Still in early development, expect breaking changes**

[![Docker](https://img.shields.io/docker/v/fallenbreath/etunnel/latest)](https://hub.docker.com/r/fallenbreath/etunnel)

A secure encrypted tunnel. Encrypt and transfer your secret data on public network safely

Thanks [go-shadowsocks2](https://github.com/shadowsocks/go-shadowsocks2) for providing the tunnel encryption logic.
Notes that this tool only provides the tunneling functionality, it cannot be used as a proxy

## Usages

### CLI Mode

```bash
# helps
./etunnel --help
./etunnel server -h
# server
./etunnel server -c AES-128-GCM -k my_secret_key -l :12000
# client
./etunnel client -c AES-128-GCM -k my_secret_key -s 192.168.1.1:12000 \
    -t tcp://127.0.0.1:2222/127.0.0.1:22 \
    -t tcp://:8000/10.10.10.1:8080
# tool
./etunnel tool -p 5132 --reload
```

Tunnel definition format: `protocol://listen/target`, where:

- `protocol` can be `tcp` (TODO: or `udp`)
- `listen` can be an address + port e.g. `127.0.0.1:2222`, or just the port e.g. `:2222` and the host will be `0.0.0.0`
- `target` can be an address/hostname + port e.g. `127.0.0.1:22`, or just the port e.g. `:8000` and the host will be `127.0.0.1`

### Config Mode

```bash
./etunnel -c etunnel_server.yml
./etunnel -c etunnel_client.yml
```

```yaml
# etunnel_server.yml
mode: server
listen: :12000
crypt: AES-128-GCM
key: my_secret_key
```

```yaml
# etunnel_client.yml
mode: client
server: 192.168.1.1:12000
crypt: AES-128-GCM
key: my_secret_key
tunnels:
  - name: website
    listen: tcp://:8080
    target: tcp://10.10.10.1:8080
  - name: ssh
    listen: tcp://127.0.0.1:2222
    target: tcp://127.0.0.1:22
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

`docker run fallenbreath/etunnel xxx` works just like `./etunnel xxx`

## Security

etunnel uses Authenticated Encryption with Associated Data (AEAD) encryption methods to encrypt the whole tunneled data

Supported encryption methods:

- `AES-128-GCM` (default)
- `AES-256-GCM`
- `CHACHA20-IETF-POLY1305`

## TODO

- [ ] hot-reload client
- [ ] `udp` / `unixgram` support for tunnel
- [ ] `kcp` / `quic` support for client - server communication
- [ ] tcp proxy protocol support
