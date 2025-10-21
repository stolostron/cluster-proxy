# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Cluster Proxy is a Go-based Kubernetes addon for Open Cluster Management (OCM) that provides network connectivity between hub clusters and managed clusters via reverse proxy tunnels. It automatically installs `apiserver-network-proxy` components to enable hub network clients to access services in managed cluster networks, even across VPC boundaries.

## Common Development Commands

### Build and Development
```bash
make build                    # Build addon-manager and addon-agent binaries
make verify                   # Run fmt, vet, and golint
make fmt                      # Format Go code
make vet                      # Run go vet
make golint                   # Run golangci-lint (v1.64.8)
```

### Testing
```bash
make test                     # Run unit tests with coverage (./pkg/...)
make test-integration         # Run integration tests with envtest
make test-e2e                 # Run end-to-end tests with Kind clusters
```

### Code Generation
```bash
make generate                 # Generate DeepCopy methods
make manifests               # Generate CRDs and RBAC
make client-gen              # Generate API clients
```

### Container Images
```bash
make images                   # Build container images
make docker-build            # Build Docker image
make docker-push             # Push Docker image
```

## Architecture Overview

### Core Components
- **Addon-Manager** (`cmd/addon-manager/`): Hub cluster component that manages proxy servers (proxy ingress)
- **Addon-Agent** (`cmd/addon-agent/`): Managed cluster component that manages proxy agents

### Key Packages
- `pkg/apis/proxy/v1alpha1/`: Custom Resource Definitions (ManagedProxyConfiguration, ManagedProxyServiceResolver)
- `pkg/proxyserver/`: Hub-side proxy server logic and controllers
- `pkg/proxyagent/`: Managed cluster-side proxy agent logic
- `pkg/config/`: Configuration management utilities
- `pkg/generated/`: Generated client code for custom resources

### Design Patterns
- **Kubernetes Operator Pattern**: Uses controller-runtime v0.18.4
- **OCM Addon Framework**: Integrates with Open Cluster Management addon-framework
- **Hub-Spoke Architecture**: Central hub with multiple managed clusters
- **Reverse Proxy Pattern**: Managed clusters initiate connections to hub cluster

### Dependencies
- Kubernetes v0.30.2 (client-go, apimachinery, api)
- Open Cluster Management (addon-framework, api, sdk-go)
- apiserver-network-proxy (konnectivity-client)
- gRPC for tunnel communication

## Testing Framework

- **Unit Tests**: Standard Go testing with coverage reporting
- **Integration Tests**: Uses controller-runtime's envtest framework
- **E2E Tests**: Ginkgo v2 and Gomega with Kind clusters for BDD-style testing

## Special Development Considerations

1. **Certificate Management**: Automatic TLS certificate generation and rotation for secure proxy tunnels
2. **Multi-cluster Networking**: Designed for hub-spoke topology where managed clusters may be in isolated VPCs
3. **Code Generation**: Requires running `make generate` and `make manifests` after API changes
4. **DCO Required**: All commits must include Developer Certificate of Origin sign-off
5. **gRPC Communication**: Network tunnels use gRPC protocol between hub and managed clusters
6. **LoadBalancer Configuration**: Supports both LoadBalancer services and custom hostname entrypoints

## Deployment

The project includes Helm charts in `charts/cluster-proxy/` for production deployment:
```bash
helm install cluster-proxy ocm/cluster-proxy -n open-cluster-management-addon --create-namespace
```