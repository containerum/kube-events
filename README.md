# Kube-events for Containerum
Kube-events is a component that sources events from Kubernetes clusters and writes them in the database for [Containerum](https://github.com/containerum/containerum).

## Prerequisites
* Kubernetes

## Installation

### Using Helm

```
  helm repo add containerum https://charts.containerum.io
  helm repo update
  helm install containerum/kube-events
```

## Contributions
Please submit all contributions concerning Kube-events component to this repository. Contributing guidelines are available [here](https://github.com/containerum/containerum/blob/master/CONTRIBUTING.md).

## License
Kube-events project is licensed under the terms of the Apache License Version 2.0. Please see LICENSE in this repository for more details.
