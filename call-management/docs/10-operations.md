# 10. Operations: deploy, size, monitor, recover

## 10.1 Where things run

| Component | Where | Notes |
|---|---|---|
| Yeastar PBX | **On-premises appliance/software edition in our office or data centre**, or **Yeastar Cloud edition** | On-prem keeps recordings in Kenya and lowers latency to local carriers; cloud removes hardware care. Decide in 11. |
| Call service | Container, close to the PBX (same data centre/region) | Low latency to the PBX event stream matters more than closeness to Hub. Two replicas. |
| Postgres, Redis | Managed or self-hosted next to the service | Separate from Hub's database. |
| Object storage | S3-compatible, in-country if required | Recordings, voicemails. |
| TURN server (coturn) | Public IP, next to the PBX | For agents behind strict networks. |

Hub's current hosting (Render) can run the call service container, but the PBX needs a reachable, stable network path; if the PBX is in our office, the service connects over a **site-to-site VPN** or a fixed public IP with an allow-list.

## 10.2 Sizing (example: 60 agents, 5,000 calls a day)

| Item | Estimate | How derived |
|---|---|---|
| Peak concurrent calls | ~45 | 60 agents × 75% busy at peak |
| Trunk channels | 60 (two trunks of 30 on different carriers) | Peak calls + waiting callers in queue + headroom |
| Bandwidth at PBX | ~10 Mbps each way | 45 calls × ~100 kbps × 2 for headroom |
| Transcoding (Opus↔G.711) | 45 sessions | Check PBX model's limit |
| Recording storage | ~35 GB/month | 5,000 calls × 4 min avg × ~1 MB/min (MP3/compressed); ~130 MB/min as WAV, so store compressed |
| Database growth | ~3 GB/month | calls + events + audit |
| PBX model | One that supports ≥ 100 users and ≥ 60 concurrent calls with recording | Check Yeastar model capacity sheet |

## 10.3 Monitoring

| Signal | Alert when |
|---|---|
| PBX event socket connected | Disconnected > 30 s |
| Event lag (PBX time → processed) | > 2 s for 1 min |
| LMS lookup p95 latency / error rate | > 500 ms / > 5% |
| Registered agents vs scheduled | Sudden drop (network issue) |
| Trunk channels in use | > 80% |
| Calls failing (SIP 5xx from carrier) | > 5% in 10 min |
| Recording archive backlog | > 50 files or oldest > 1 h |
| Audit chain verification | Any break |
| Disk on PBX | > 70% |

Metrics via Prometheus, logs in JSON with the call ID on every line (so one call can be traced end to end), dashboards per queue and per carrier.

## 10.4 What fails and what happens

| Failure | Effect on customers | Effect on agents | Recovery |
|---|---|---|---|
| Call service down | **None**: calls still route and ring | No caller names, no history updates; softphone still works if already registered | Restart; reconcile from PBX CDR |
| LMS down | None | Number only, no customer card | Retry lookups; nothing lost |
| PBX down | **Calls fail** | No calls | PBX high availability (Yeastar HA pair) or carrier failover to a backup number/mobile group |
| One carrier down | Calls on that trunk fail | — | Outbound routes fail over to the second trunk; ask carrier to forward DIDs to the backup |
| Office internet down (on-prem PBX) | Inbound fails | — | Carrier-side "failover destination" (forward DIDs to mobile numbers of on-call staff) |
| Agent internet poor | Bad audio | Poor-line warning | Agent switches network; supervisor reassigns |

## 10.5 Backups and disaster recovery
- Postgres: continuous WAL archiving + nightly full backup, kept 35 days; restore tested monthly.
- Recordings: object storage with versioning and a second-region copy.
- PBX configuration: nightly automatic backup exported off the PBX.
- Targets: service back within 1 hour (RTO), at most 5 minutes of data lost (RPO). Calls themselves keep working while the service is restored.

## 10.6 Environments
| Env | PBX | Customers |
|---|---|---|
| Development | Yeastar trial/test PBX (or the Hub simulation) | Fake LMS with test data |
| Staging | Test PBX with one test trunk and test DIDs | LMS staging copy, masked |
| Production | Production PBX | Real LMS |

No real customer data in development. Test calls in staging go only to staff phones on an allow-list.
