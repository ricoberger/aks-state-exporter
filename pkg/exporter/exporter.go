package exporter

import (
	"context"
	"log/slog"

	"github.com/ricoberger/aks-state-exporter/pkg/exporter/aks"

	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	AKS aks.Config `json:"aks"`
}

// StatsCollector collects the AKS state metrics and implements the
// `prometheus.Collector` interface so it can be used as follows:
type Exporter struct {
	aksClient                 aks.Client
	ClusterProvisioningState  *prometheus.Desc
	NodePoolProvisioningState *prometheus.Desc
}

// New returns a new `Exporter` which can be passed to the
// `prometheus.MustRegister` function to collect the AKS state metrics.
func New(config Config) (*Exporter, error) {
	aksClient, err := aks.NewClient(config.AKS)
	if err != nil {
		return nil, err
	}

	return &Exporter{
		aksClient:                 aksClient,
		ClusterProvisioningState:  prometheus.NewDesc("aks_cluster_provisioning_state", "The provisioning state of the cluster (0 - Unknown, 1 - Succeeded, 2 - Failed, 3 - Canceled, 4 - Creating, 5 - Updating, 6 - Deleting, 7 - Upgrading, 8 - UpgradingNodeImageVersion, 9 - ReconcilingClusterETCDCertificates)", []string{"name", "resource_group"}, nil),
		NodePoolProvisioningState: prometheus.NewDesc("aks_nodepool_provisioning_state", "The provisioning state of the node pool (0 - Unknown, 1 - Succeeded, 2 - Failed, 3 - Canceled, 4 - Creating, 5 - Updating, 6 - Deleting, 7 - Upgrading, 8 - UpgradingNodeImageVersion, 9 - ReconcilingClusterETCDCertificates)", []string{"name", "cluster", "resource_group"}, nil),
	}, nil
}

// Describe sends the super-set of all possible descriptors of metrics collected
// by the StatusCollector to the provided channel and returns once the last
// descriptor has been sent. The sent descriptors fulfill the consistency and
// uniqueness requirements described in the Desc documentation.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.ClusterProvisioningState
	ch <- e.NodePoolProvisioningState
}

// Collect is called by the Prometheus registry when collecting metrics. The
// implementation sends each collected metric via the provided channel and
// returns once the last metric has been sent. The descriptor of each sent
// metric is one of those returned by Describe. Returned metrics that share the
// same descriptor must differ in their variable label values.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	slog.Debug("Collecting metrics")

	ctx := context.Background()

	clusters, err := e.aksClient.GetClusters(ctx)
	if err != nil {
		slog.Error("Failed to get clusters", slog.String("error", err.Error()))
		return
	}

	slog.Debug("Collecting metrics for clusters", slog.Int("count", len(clusters)))

	for _, cluster := range clusters {
		slog.Debug("Collecting metrics for cluster", slog.String("name", cluster.Name), slog.String("resource_group", cluster.ResourceGroup), slog.String("provisioning_state", cluster.ProvisioningState))
		ch <- prometheus.MustNewConstMetric(e.ClusterProvisioningState, prometheus.GaugeValue, provisioningStateToFloat64(cluster.ProvisioningState), cluster.Name, cluster.ResourceGroup)

		nodePools, err := e.aksClient.GetNodePools(ctx, cluster.Name, cluster.ResourceGroup)
		if err != nil {
			slog.Error("Failed to get node pools", slog.String("error", err.Error()))
			return
		}

		slog.Debug("Collecting metrics for node pools", slog.String("cluster", cluster.Name), slog.String("resource_group", cluster.ResourceGroup), slog.Int("count", len(nodePools)))

		for _, nodePool := range nodePools {
			slog.Debug("Collecting metrics for node pool", slog.String("name", nodePool.Name), slog.String("cluster", nodePool.Cluster), slog.String("resource_group", nodePool.ResourceGroup), slog.String("provisioning_state", nodePool.ProvisioningState))
			ch <- prometheus.MustNewConstMetric(e.NodePoolProvisioningState, prometheus.GaugeValue, provisioningStateToFloat64(nodePool.ProvisioningState), nodePool.Name, nodePool.Cluster, nodePool.ResourceGroup)
		}
	}
}

func provisioningStateToFloat64(state string) float64 {
	switch state {
	case "Succeeded":
		return 1
	case "Failed":
		return 2
	case "Canceled":
		return 3
	case "Creating":
		return 4
	case "Updating":
		return 5
	case "Deleting":
		return 6
	case "Upgrading":
		return 7
	case "UpgradingNodeImageVersion":
		return 8
	case "ReconcilingClusterETCDCertificates":
		return 9
	default:
		return 0
	}
}
