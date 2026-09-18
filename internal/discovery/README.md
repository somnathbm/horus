# /internal/discovery

responsible for service discovery from - sources and gives one or more -> resources (Targets).

# Process

```
                 Manager
                    │
       ┌────────────┼────────────┐
       ▼            ▼            ▼
 LocalDiscoverer  AWS         Kubernetes
       │            │            │
 DiscoveryResult DiscoveryResult DiscoveryResult
       └────────────┼────────────┘
                    ▼
           combined DiscoveryResult
```

![discovery-process](./docs/discovery_process.png)