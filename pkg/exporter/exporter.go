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
	aksClient                  aks.Client
	ClusterProvisioningState   *prometheus.Desc
	ClusterInfo                *prometheus.Desc
	NodePoolProvisioningState  *prometheus.Desc
	NodePoolInfo               *prometheus.Desc
	NodePoolCount              *prometheus.Desc
	NodePoolMinCount           *prometheus.Desc
	NodePoolMaxCount           *prometheus.Desc
	NodePoolAutoScalingEnabled *prometheus.Desc
}

// New returns a new `Exporter` which can be passed to the
// `prometheus.MustRegister` function to collect the AKS state metrics.
func New(config Config) (*Exporter, error) {
	aksClient, err := aks.NewClient(config.AKS)
	if err != nil {
		return nil, err
	}

	return newExporter(aksClient), nil
}

// newExporter creates an `Exporter` for the provided `aks.Client`. It is
// separated from `New` so the exporter can be constructed with a fake client in
// tests without requiring valid Azure credentials.
func newExporter(aksClient aks.Client) *Exporter {
	return &Exporter{
		aksClient:                  aksClient,
		ClusterProvisioningState:   prometheus.NewDesc("aks_cluster_provisioning_state", "The provisioning state of the cluster (0 - Unknown, 1 - Succeeded, 2 - Failed, 3 - Canceled, 4 - Creating, 5 - Updating, 6 - Deleting, 7 - Upgrading, 8 - UpgradingNodeImageVersion, 9 - ReconcilingClusterETCDCertificates)", []string{"name", "resource_group"}, nil),
		ClusterInfo:                prometheus.NewDesc("aks_cluster_info", "Information about the cluster as labels, the value is always 1", []string{"name", "resource_group", "location", "kubernetes_version", "current_kubernetes_version", "sku_tier", "power_state", "provisioning_state"}, nil),
		NodePoolProvisioningState:  prometheus.NewDesc("aks_nodepool_provisioning_state", "The provisioning state of the node pool (0 - Unknown, 1 - Succeeded, 2 - Failed, 3 - Canceled, 4 - Creating, 5 - Updating, 6 - Deleting, 7 - Upgrading, 8 - UpgradingNodeImageVersion, 9 - ReconcilingClusterETCDCertificates)", []string{"name", "cluster", "resource_group"}, nil),
		NodePoolInfo:               prometheus.NewDesc("aks_nodepool_info", "Information about the node pool as labels, the value is always 1", []string{"name", "cluster", "resource_group", "vm_size", "os_type", "os_sku", "mode", "orchestrator_version", "current_orchestrator_version", "node_image_version", "scale_set_priority", "provisioning_state"}, nil),
		NodePoolCount:              prometheus.NewDesc("aks_nodepool_count", "The number of nodes in the node pool", []string{"name", "cluster", "resource_group"}, nil),
		NodePoolMinCount:           prometheus.NewDesc("aks_nodepool_min_count", "The minimum number of nodes in the node pool", []string{"name", "cluster", "resource_group"}, nil),
		NodePoolMaxCount:           prometheus.NewDesc("aks_nodepool_max_count", "The maximum number of nodes in the node pool", []string{"name", "cluster", "resource_group"}, nil),
		NodePoolAutoScalingEnabled: prometheus.NewDesc("aks_nodepool_autoscaling_enabled", "Whether autoscaling is enabled for the node pool (0 - disabled, 1 - enabled)", []string{"name", "cluster", "resource_group"}, nil),
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected
// by the StatusCollector to the provided channel and returns once the last
// descriptor has been sent. The sent descriptors fulfill the consistency and
// uniqueness requirements described in the Desc documentation.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.ClusterProvisioningState
	ch <- e.ClusterInfo
	ch <- e.NodePoolProvisioningState
	ch <- e.NodePoolInfo
	ch <- e.NodePoolCount
	ch <- e.NodePoolMinCount
	ch <- e.NodePoolMaxCount
	ch <- e.NodePoolAutoScalingEnabled
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
		ch <- prometheus.MustNewConstMetric(e.ClusterInfo, prometheus.GaugeValue, 1, cluster.Name, cluster.ResourceGroup, cluster.Location, cluster.KubernetesVersion, cluster.CurrentKubernetesVersion, cluster.SKUTier, cluster.PowerState, cluster.ProvisioningState)

		nodePools, err := e.aksClient.GetNodePools(ctx, cluster.Name, cluster.ResourceGroup)
		if err != nil {
			slog.Error("Failed to get node pools", slog.String("error", err.Error()))
			return
		}

		slog.Debug("Collecting metrics for node pools", slog.String("cluster", cluster.Name), slog.String("resource_group", cluster.ResourceGroup), slog.Int("count", len(nodePools)))

		for _, nodePool := range nodePools {
			slog.Debug("Collecting metrics for node pool", slog.String("name", nodePool.Name), slog.String("cluster", nodePool.Cluster), slog.String("resource_group", nodePool.ResourceGroup), slog.String("provisioning_state", nodePool.ProvisioningState), slog.Int("count", int(nodePool.Count)), slog.Int("min_count", int(nodePool.MinCount)), slog.Int("max_count", int(nodePool.MaxCount)))
			ch <- prometheus.MustNewConstMetric(e.NodePoolProvisioningState, prometheus.GaugeValue, provisioningStateToFloat64(nodePool.ProvisioningState), nodePool.Name, nodePool.Cluster, nodePool.ResourceGroup)
			ch <- prometheus.MustNewConstMetric(e.NodePoolInfo, prometheus.GaugeValue, 1, nodePool.Name, nodePool.Cluster, nodePool.ResourceGroup, nodePool.VMSize, nodePool.OSType, nodePool.OSSKU, nodePool.Mode, nodePool.OrchestratorVersion, nodePool.CurrentOrchestratorVersion, nodePool.NodeImageVersion, nodePool.ScaleSetPriority, nodePool.ProvisioningState)
			ch <- prometheus.MustNewConstMetric(e.NodePoolCount, prometheus.GaugeValue, float64(nodePool.Count), nodePool.Name, nodePool.Cluster, nodePool.ResourceGroup)
			ch <- prometheus.MustNewConstMetric(e.NodePoolMinCount, prometheus.GaugeValue, float64(nodePool.MinCount), nodePool.Name, nodePool.Cluster, nodePool.ResourceGroup)
			ch <- prometheus.MustNewConstMetric(e.NodePoolMaxCount, prometheus.GaugeValue, float64(nodePool.MaxCount), nodePool.Name, nodePool.Cluster, nodePool.ResourceGroup)
			ch <- prometheus.MustNewConstMetric(e.NodePoolAutoScalingEnabled, prometheus.GaugeValue, boolToFloat64(nodePool.AutoScalingEnabled), nodePool.Name, nodePool.Cluster, nodePool.ResourceGroup)
		}
	}
}

func boolToFloat64(value bool) float64 {
	if value {
		return 1
	}

	return 0
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
