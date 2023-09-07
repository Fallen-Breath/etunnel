# etunnel - encrypted tunnel

## Usages

```bash
# server
./etunnel -m server -c AES-128-GCM -k my_secret_key -l :12000
# client
./etunnel -m client -c AES-128-GCM -k my_secret_key -s 192.168.1.1:12000 -t tcp://:8080/10.10.10.1:8080 -t tcp://127.0.0.1:2222/127.0.0.1:22
```

Available protocols: `tcp`, `udp`, `unix`, `unixgram`
