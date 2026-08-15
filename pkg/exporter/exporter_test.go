package exporter

import (
	"context"
	"strings"
	"testing"

	"github.com/ricoberger/aks-state-exporter/pkg/exporter/aks"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

type fakeClient struct {
	clusters     []aks.Cluster
	clustersErr  error
	nodePools    map[string][]aks.NodePool
	nodePoolsErr error
}

func (c *fakeClient) GetClusters(_ context.Context) ([]aks.Cluster, error) {
	return c.clusters, c.clustersErr
}

func (c *fakeClient) GetNodePools(_ context.Context, clusterName string, _ string) ([]aks.NodePool, error) {
	if c.nodePoolsErr != nil {
		return nil, c.nodePoolsErr
	}

	return c.nodePools[clusterName], nil
}

func TestCollect(t *testing.T) {
	t.Run("should collect all metrics", func(t *testing.T) {
		client := &fakeClient{
			clusters: []aks.Cluster{
				{
					Name:                     "dev-de1",
					ResourceGroup:            "dev-de1",
					ProvisioningState:        "Succeeded",
					Location:                 "germanywestcentral",
					KubernetesVersion:        "1.35.5",
					CurrentKubernetesVersion: "1.35.5",
					SKUTier:                  "Standard",
					PowerState:               "Running",
				},
			},
			nodePools: map[string][]aks.NodePool{
				"dev-de1": {
					{
						Name:                       "system",
						Cluster:                    "dev-de1",
						ResourceGroup:              "dev-de1",
						ProvisioningState:          "Succeeded",
						Count:                      3,
						MinCount:                   1,
						MaxCount:                   5,
						VMSize:                     "Standard_D16ds_v5",
						OSType:                     "Linux",
						OSSKU:                      "Ubuntu",
						Mode:                       "System",
						OrchestratorVersion:        "1.35.5",
						CurrentOrchestratorVersion: "1.35.5",
						NodeImageVersion:           "AKSUbuntu-2404gen2containerd-202607.29.0",
						ScaleSetPriority:           "Spot",
					},
					{
						Name:          "empty",
						Cluster:       "dev-de1",
						ResourceGroup: "dev-de1",
					},
				},
			},
		}

		expected := `
# HELP aks_cluster_info Information about the cluster as labels, the value is always 1
# TYPE aks_cluster_info gauge
aks_cluster_info{current_kubernetes_version="1.35.5",kubernetes_version="1.35.5",location="germanywestcentral",name="dev-de1",power_state="Running",provisioning_state="Succeeded",resource_group="dev-de1",sku_tier="Standard"} 1
# HELP aks_cluster_provisioning_state The provisioning state of the cluster (0 - Unknown, 1 - Succeeded, 2 - Failed, 3 - Canceled, 4 - Creating, 5 - Updating, 6 - Deleting, 7 - Upgrading, 8 - UpgradingNodeImageVersion, 9 - ReconcilingClusterETCDCertificates)
# TYPE aks_cluster_provisioning_state gauge
aks_cluster_provisioning_state{name="dev-de1",resource_group="dev-de1"} 1
# HELP aks_nodepool_count The number of nodes in the node pool
# TYPE aks_nodepool_count gauge
aks_nodepool_count{cluster="dev-de1",name="empty",resource_group="dev-de1"} 0
aks_nodepool_count{cluster="dev-de1",name="system",resource_group="dev-de1"} 3
# HELP aks_nodepool_info Information about the node pool as labels, the value is always 1
# TYPE aks_nodepool_info gauge
aks_nodepool_info{cluster="dev-de1",current_orchestrator_version="1.35.5",mode="System",name="system",node_image_version="AKSUbuntu-2404gen2containerd-202607.29.0",orchestrator_version="1.35.5",os_sku="Ubuntu",os_type="Linux",provisioning_state="Succeeded",resource_group="dev-de1",scale_set_priority="Spot",vm_size="Standard_D16ds_v5"} 1
aks_nodepool_info{cluster="dev-de1",current_orchestrator_version="",mode="",name="empty",node_image_version="",orchestrator_version="",os_sku="",os_type="",provisioning_state="",resource_group="dev-de1",scale_set_priority="",vm_size=""} 1
# HELP aks_nodepool_max_count The maximum number of nodes in the node pool
# TYPE aks_nodepool_max_count gauge
aks_nodepool_max_count{cluster="dev-de1",name="empty",resource_group="dev-de1"} 0
aks_nodepool_max_count{cluster="dev-de1",name="system",resource_group="dev-de1"} 5
# HELP aks_nodepool_min_count The minimum number of nodes in the node pool
# TYPE aks_nodepool_min_count gauge
aks_nodepool_min_count{cluster="dev-de1",name="empty",resource_group="dev-de1"} 0
aks_nodepool_min_count{cluster="dev-de1",name="system",resource_group="dev-de1"} 1
# HELP aks_nodepool_provisioning_state The provisioning state of the node pool (0 - Unknown, 1 - Succeeded, 2 - Failed, 3 - Canceled, 4 - Creating, 5 - Updating, 6 - Deleting, 7 - Upgrading, 8 - UpgradingNodeImageVersion, 9 - ReconcilingClusterETCDCertificates)
# TYPE aks_nodepool_provisioning_state gauge
aks_nodepool_provisioning_state{cluster="dev-de1",name="empty",resource_group="dev-de1"} 0
aks_nodepool_provisioning_state{cluster="dev-de1",name="system",resource_group="dev-de1"} 1
`

		exporter := newExporter(client)
		err := testutil.CollectAndCompare(exporter, strings.NewReader(expected))
		require.NoError(t, err)
	})

	t.Run("should collect no metrics when there are no clusters", func(t *testing.T) {
		exporter := newExporter(&fakeClient{})
		require.Equal(t, 0, testutil.CollectAndCount(exporter))
	})
}
