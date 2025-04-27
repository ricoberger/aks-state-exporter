# AKS State Exporter

The AKS state exporter can be used to scrape the provisioning state of AKS
clusters and their node pools and export them as Prometheus metrics.

## Building and Running

To build and run the AKS state exporter the following commands can be used:

```sh
git clone https://github.com/ricoberger/aks-state-exporter.git
cd aks-state-exporter
make build

./bin/aks-state-exporter
```

Via Docker the following commands can be used to build the image and run the
exporter:

```sh
docker build -f ./Dockerfile -t ghcr.io/ricoberger/aks-state-exporter:latest .
docker run --rm -it --name aks-state-exporter -p 8080:8080 -v $(pwd)/tmp:/aks-state-exporter/config ghcr.io/ricoberger/aks-state-exporter:latest --config=/aks-state-exporter/config/config.yaml
```

The exporter can also be deployed on Kubernetes via Helm:

```sh
helm upgrade --install aks-state-exporter oci://ghcr.io/ricoberger/charts/aks-state-exporter --version <VERSION>
```

## Metrics

```txt
# HELP aks_cluster_provisioning_state The provisioning state of the cluster (0 - Unknown, 1 - Succeeded, 2 - Failed, 3 - Canceled, 4 - Creating, 5 - Updating, 6 - Deleting, 7 - Upgrading, 8 - UpgradingNodeImageVersion, 9 - ReconcilingClusterETCDCertificates)
# TYPE aks_cluster_provisioning_state gauge
aks_cluster_provisioning_state{name="dev-de1",resource_group="dev-de1"} 1
aks_cluster_provisioning_state{name="prod-de1",resource_group="prod-de1"} 1
aks_cluster_provisioning_state{name="stage-de1",resource_group="stage-de1"} 1
# HELP aks_nodepool_provisioning_state The provisioning state of the node pool (0 - Unknown, 1 - Succeeded, 2 - Failed, 3 - Canceled, 4 - Creating, 5 - Updating, 6 - Deleting, 7 - Upgrading, 8 - UpgradingNodeImageVersion, 9 - ReconcilingClusterETCDCertificates)
# TYPE aks_nodepool_provisioning_state gauge
aks_nodepool_provisioning_state{cluster="dev-de1",name="system",resource_group="dev-de1"} 1
aks_nodepool_provisioning_state{cluster="dev-de1",name="zone1",resource_group="dev-de1"} 1
aks_nodepool_provisioning_state{cluster="dev-de1",name="zone2",resource_group="dev-de1"} 1
aks_nodepool_provisioning_state{cluster="dev-de1",name="zone3",resource_group="dev-de1"} 1
aks_nodepool_provisioning_state{cluster="prod-de1",name="system",resource_group="prod-de1"} 1
aks_nodepool_provisioning_state{cluster="prod-de1",name="zone1",resource_group="prod-de1"} 1
aks_nodepool_provisioning_state{cluster="prod-de1",name="zone2",resource_group="prod-de1"} 1
aks_nodepool_provisioning_state{cluster="prod-de1",name="zone3",resource_group="prod-de1"} 1
aks_nodepool_provisioning_state{cluster="stage-de1",name="system",resource_group="stage-de1"} 1
aks_nodepool_provisioning_state{cluster="stage-de1",name="zone1",resource_group="stage-de1"} 1
aks_nodepool_provisioning_state{cluster="stage-de1",name="zone2",resource_group="stage-de1"} 1
aks_nodepool_provisioning_state{cluster="stage-de1",name="zone3",resource_group="stage-de1"} 1
```
