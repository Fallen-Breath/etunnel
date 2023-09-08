# etunnel

A secure encrypted tunnel

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
./etunnel --conf etunnel_server.yml
./etunnel --conf etunnel_client.yml
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
