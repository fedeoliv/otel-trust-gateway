# Kubernetes Deployment Guide

This directory contains Kubernetes manifests for deploying the custom OpenTelemetry Collector in Azure Kubernetes Service (AKS) with high availability and resilience.

## Architecture Overview

```
Mobile Apps (Multiple Clients)
    ↓
Azure Load Balancer / Ingress
    ↓
Kubernetes Service (Load Balancing)
    ↓
┌─────────────────────────────────────────┐
│  Pod 1        Pod 2        Pod 3        │
│  (Zone 1)     (Zone 2)     (Zone 3)     │
│                                          │
│  Each pod:                               │
│  - 256Mi-512Mi memory                    │
│  - Health checks                         │
│  - Auto-restart on failure               │
└─────────────────────────────────────────┘
    ↓
Horizontal Pod Autoscaler (3-10 pods)
    ↓
External Telemetry Backend
```

## Resilience Features

### 1. **High Availability (HA)**
- ✅ **3 replicas** by default (can handle 1 pod failure)
- ✅ **Anti-affinity rules** - Pods spread across different nodes
- ✅ **Topology spread** - Pods distributed across availability zones
- ✅ **Pod Disruption Budget** - Ensures minimum 2 pods always running

### 2. **Automatic Failover**
- ✅ **Load balancing** - Kubernetes Service distributes traffic across healthy pods
- ✅ **Readiness probes** - Unhealthy pods removed from load balancer automatically
- ✅ **Liveness probes** - Failed pods restarted automatically
- ✅ **Rolling updates** - Zero-downtime deployments

### 3. **Auto-Scaling**
- ✅ **Horizontal Pod Autoscaler (HPA)** - Scales from 3 to 10 pods based on:
  - Memory usage > 70%
  - CPU usage > 70%
- ✅ **Fast scale-up** - Adds pods immediately when needed
- ✅ **Slow scale-down** - Waits 5 minutes to avoid flapping

### 4. **Memory Management**
- ✅ **Memory limiter at 410 MiB** (80% of 512Mi pod limit)
- ✅ **Spike limit at 128 MiB** - Allows temporary bursts
- ✅ **Pod restarts** if memory limit exceeded
- ✅ **Traffic shifted** to healthy pods automatically

### 5. **Data Loss Prevention**
- ✅ **Client retries** - Mobile apps can retry failed requests to different pods
- ✅ **Graceful shutdown** (30s) - Processes in-flight data before stopping
- ✅ **Batching** - Reduces data loss by batching before export

## What Happens When Memory Limit is Reached?

### Scenario: Pod 2 reaches memory limit

```
Time    Event                           Result
────────────────────────────────────────────────────────────────
T+0s    Pod 2 reaches 80% memory       Pod 2 refuses new data
        (328 MiB / 410 MiB)            Returns HTTP 429/503

T+0s    Load balancer detects          Traffic redirected to Pod 1 & 3
        Pod 2 unhealthy                ✅ No data loss for new requests

T+5s    Readiness probe fails          Pod 2 removed from Service
                                       ✅ Clients only see healthy pods

T+10s   Pod 2 reaches 100% memory      Kubernetes kills and restarts Pod 2
        (512 MiB)                      

T+15s   Pod 2 restarts clean           Pod 2 rejoins load balancer
                                       ✅ System back to full capacity

T+30s   HPA detects high memory        Spawns Pod 4 for extra capacity
        across pods (avg > 70%)        ✅ Prevents future issues
```

### Key Points:
- ❌ **Data in Pod 2 is lost** when memory limit reached
- ✅ **New requests go to healthy pods** (Pod 1, 3) - No client impact
- ✅ **Pod 2 auto-restarts** and rejoins cluster
- ✅ **HPA scales out** to prevent recurrence
- ✅ **Mobile apps see minimal errors** due to load balancing

## Deployment

### Prerequisites

```bash
# Create namespace
kubectl create namespace observability

# Build and push Docker image to Azure Container Registry
az acr login --name <your-acr-name>
docker build -t <your-acr-name>.azurecr.io/otelcol-custom:latest .
docker push <your-acr-name>.azurecr.io/otelcol-custom:latest
```

### Deploy

```bash
# Update image in deployment.yaml
# Then apply all manifests
kubectl apply -f k8s/deployment.yaml

# Verify deployment
kubectl get pods -n observability
kubectl get svc -n observability
kubectl get hpa -n observability
```

### Verify High Availability

```bash
# Check pod distribution across nodes
kubectl get pods -n observability -o wide

# Check HPA status
kubectl get hpa -n observability -w

# Simulate pod failure (test resilience)
kubectl delete pod -n observability -l app=otelcol-custom --wait=false
kubectl get pods -n observability -w  # Watch pods recover
```

### Monitor

```bash
# Check pod health
kubectl get pods -n observability
kubectl describe pod -n observability <pod-name>

# View logs from all pods
kubectl logs -n observability -l app=otelcol-custom --tail=100 -f

# Check service endpoints
kubectl get endpoints -n observability otelcol-custom

# Check HPA metrics
kubectl get hpa -n observability otelcol-custom
kubectl describe hpa -n observability otelcol-custom
```

## Configuration Tuning

### For Higher Traffic (Production)

```yaml
# In deployment.yaml
spec:
  replicas: 5  # Increase baseline

  resources:
    requests:
      memory: 512Mi
    limits:
      memory: 1Gi

# In config.yaml
memory_limiter:
  limit_mib: 820  # 80% of 1Gi
```

### For Lower Costs (Development)

```yaml
# In deployment.yaml
spec:
  replicas: 1  # Single replica

# Remove or disable HPA
```

## Troubleshooting

### Pods keep restarting
```bash
# Check memory usage
kubectl top pods -n observability

# Check logs before crash
kubectl logs -n observability <pod-name> --previous

# Increase memory limits in deployment.yaml
```

### High latency
```bash
# Check if HPA is scaling
kubectl get hpa -n observability

# Check pod distribution
kubectl get pods -n observability -o wide

# Manually scale up
kubectl scale deployment otelcol-custom -n observability --replicas=5
```

### Data loss issues
```bash
# Increase batch timeout and size for more buffering
# Add persistent volume for queuing (advanced)
# Configure retry policies in mobile app
```

## Best Practices

1. **Always run at least 3 replicas** for production
2. **Set memory_limiter to 80% of pod memory limit**
3. **Monitor HPA metrics** to tune scaling thresholds
4. **Use PodDisruptionBudget** to prevent cluster-wide outages
5. **Configure client retries** in mobile apps (exponential backoff)
6. **Use Application Insights** or Prometheus to monitor collector health
7. **Test failover regularly** by simulating pod failures

## Cost Optimization

```yaml
# Use Azure Spot instances for collector pods (70% cost reduction)
nodeSelector:
  kubernetes.azure.com/scalesetpriority: spot

# Or use lower-cost node pools
nodeSelector:
  agentpool: standard
```

## Next Steps

- [ ] Configure external exporter (Azure Monitor, Application Insights)
- [ ] Set up alerting on collector health metrics
- [ ] Implement persistent queue for critical telemetry
- [ ] Add network policies for security
- [ ] Configure Ingress for external access
