package aks

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"

	"github.com/stretchr/testify/require"
)

func TestMapCluster(t *testing.T) {
	t.Run("should map a fully populated cluster", func(t *testing.T) {
		cluster := &armcontainerservice.ManagedCluster{
			Name:     to.Ptr("dev-de1"),
			Location: to.Ptr("germanywestcentral"),
			SKU: &armcontainerservice.ManagedClusterSKU{
				Tier: to.Ptr(armcontainerservice.ManagedClusterSKUTierPaid),
			},
			Properties: &armcontainerservice.ManagedClusterProperties{
				ProvisioningState:        to.Ptr("Succeeded"),
				KubernetesVersion:        to.Ptr("1.35.5"),
				CurrentKubernetesVersion: to.Ptr("1.35.4"),
				PowerState: &armcontainerservice.PowerState{
					Code: to.Ptr(armcontainerservice.CodeRunning),
				},
			},
		}

		mapped, ok := mapCluster(cluster, "dev-de1")
		require.True(t, ok)
		require.Equal(t, Cluster{
			Name:                     "dev-de1",
			ResourceGroup:            "dev-de1",
			ProvisioningState:        "Succeeded",
			Location:                 "germanywestcentral",
			KubernetesVersion:        "1.35.5",
			CurrentKubernetesVersion: "1.35.4",
			SKUTier:                  "Paid",
			PowerState:               "Running",
		}, mapped)
	})

	t.Run("should map nil pointers to empty strings", func(t *testing.T) {
		cluster := &armcontainerservice.ManagedCluster{
			Name: to.Ptr("dev-de1"),
		}

		mapped, ok := mapCluster(cluster, "dev-de1")
		require.True(t, ok)
		require.Equal(t, Cluster{
			Name:          "dev-de1",
			ResourceGroup: "dev-de1",
		}, mapped)
	})

	t.Run("should skip a nil cluster", func(t *testing.T) {
		_, ok := mapCluster(nil, "dev-de1")
		require.False(t, ok)
	})

	t.Run("should skip a cluster without a name", func(t *testing.T) {
		_, ok := mapCluster(&armcontainerservice.ManagedCluster{}, "dev-de1")
		require.False(t, ok)
	})
}

func TestMapNodePool(t *testing.T) {
	t.Run("should map a fully populated node pool", func(t *testing.T) {
		nodePool := &armcontainerservice.AgentPool{
			Name: to.Ptr("system"),
			Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
				ProvisioningState:          to.Ptr("Succeeded"),
				Count:                      to.Ptr(int32(3)),
				MinCount:                   to.Ptr(int32(1)),
				MaxCount:                   to.Ptr(int32(5)),
				VMSize:                     to.Ptr("Standard_D16ds_v5"),
				OSType:                     to.Ptr(armcontainerservice.OSTypeLinux),
				OSSKU:                      to.Ptr(armcontainerservice.OSSKUUbuntu),
				Mode:                       to.Ptr(armcontainerservice.AgentPoolModeSystem),
				OrchestratorVersion:        to.Ptr("1.35.5"),
				CurrentOrchestratorVersion: to.Ptr("1.35.4"),
				NodeImageVersion:           to.Ptr("AKSUbuntu-2404gen2containerd-202607.29.0"),
				ScaleSetPriority:           to.Ptr(armcontainerservice.ScaleSetPriorityRegular),
				EnableAutoScaling:          to.Ptr(true),
			},
		}

		mapped, ok := mapNodePool(nodePool, "dev-de1", "dev-de1")
		require.True(t, ok)
		require.Equal(t, NodePool{
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
			CurrentOrchestratorVersion: "1.35.4",
			NodeImageVersion:           "AKSUbuntu-2404gen2containerd-202607.29.0",
			ScaleSetPriority:           "Regular",
			AutoScalingEnabled:         true,
		}, mapped)
	})

	t.Run("should map nil pointers to empty strings and zero counts", func(t *testing.T) {
		nodePool := &armcontainerservice.AgentPool{
			Name:       to.Ptr("system"),
			Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{},
		}

		mapped, ok := mapNodePool(nodePool, "dev-de1", "dev-de1")
		require.True(t, ok)
		require.Equal(t, NodePool{
			Name:          "system",
			Cluster:       "dev-de1",
			ResourceGroup: "dev-de1",
		}, mapped)
	})

	t.Run("should skip a nil node pool", func(t *testing.T) {
		_, ok := mapNodePool(nil, "dev-de1", "dev-de1")
		require.False(t, ok)
	})

	t.Run("should skip a node pool without properties", func(t *testing.T) {
		_, ok := mapNodePool(&armcontainerservice.AgentPool{Name: to.Ptr("system")}, "dev-de1", "dev-de1")
		require.False(t, ok)
	})

	t.Run("should skip a node pool without a name", func(t *testing.T) {
		nodePool := &armcontainerservice.AgentPool{
			Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{},
		}

		_, ok := mapNodePool(nodePool, "dev-de1", "dev-de1")
		require.False(t, ok)
	})
}
