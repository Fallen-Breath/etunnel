# etunnel

[![Docker](https://img.shields.io/docker/v/fallenbreath/etunnel/latest)](https://hub.docker.com/r/fallenbreath/etunnel)

A secure encrypted tunnel. Encrypt and transfer your secret data on public network safely

Thanks [go-shadowsocks2](https://github.com/shadowsocks/go-shadowsocks2) for providing the tunnel encryption logic.
Notes that this tool only provides the tunneling functionality, it cannot be used as a proxy

## Usages

### CLI Mode

```bash
# server
./etunnel server -c AES-128-GCM -k my_secret_key -l :12000
# client
./etunnel client -c AES-128-GCM -k my_secret_key -s 192.168.1.1:12000 -t tcp://:8080/10.10.10.1:8080 -t tcp://127.0.0.1:2222/127.0.0.1:22
```

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

## TODO

- [ ] hot-reload client
- [ ] `udp` / `unixgram` support for tunnel
- [ ] `kcp` / `quic` support for client - server communication
- [ ] tcp proxy protocol support
