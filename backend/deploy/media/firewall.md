# Media server network checklist
Open to the internet: TCP 443 (signalling behind your TLS proxy, or 7880 directly), TCP 7881 (WebRTC over TCP), UDP 50000-60000
(media, one port per connection), UDP 3478 and TCP 5349 (TURN). Keep closed to the internet: 6789 (metrics), 9100 (node exporter),
and Redis. The API must be reachable from the media servers at `HUB_API_URL` for the signed notices.
Kernel: `net.core.rmem_max=25000000`, `net.core.wmem_max=25000000`, `net.ipv4.ip_local_port_range="1024 65535"`, and the
file-descriptor limit in the compose file. Use a 10 Gbps NIC for nodes that carry video.
