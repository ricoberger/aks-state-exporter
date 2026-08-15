package aks

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
)

// Config is the structure of the configuration for a single GitHub instance.
type Config struct {
	Credentials    Credentials `json:"credentials"`
	ResourceGroups []string    `json:"resourceGroups"`
}

type Credentials struct {
	SubscriptionID string `json:"subscriptionID"`
	TenantID       string `json:"tenantID"`
	ClientID       string `json:"clientID"`
	//nolint:gosec
	ClientSecret string `json:"clientSecret"`
}

type Cluster struct {
	Name                     string
	ResourceGroup            string
	ProvisioningState        string
	Location                 string
	KubernetesVersion        string
	CurrentKubernetesVersion string
	SKUTier                  string
	PowerState               string
}

type NodePool struct {
	Name                       string
	Cluster                    string
	ResourceGroup              string
	ProvisioningState          string
	Count                      int32
	MinCount                   int32
	MaxCount                   int32
	VMSize                     string
	OSType                     string
	OSSKU                      string
	Mode                       string
	OrchestratorVersion        string
	CurrentOrchestratorVersion string
	NodeImageVersion           string
	ScaleSetPriority           string
}

type Client interface {
	GetClusters(ctx context.Context) ([]Cluster, error)
	GetNodePools(ctx context.Context, clusterName string, resourceGroup string) ([]NodePool, error)
}

type client struct {
	subscriptionID        string
	resourceGroups        []string
	managedClustersClient *armcontainerservice.ManagedClustersClient
	agentPoolsClient      *armcontainerservice.AgentPoolsClient
}

func (c *client) GetClusters(ctx context.Context) ([]Cluster, error) {
	var clusters []Cluster

	for _, resourceGroup := range c.resourceGroups {
		pager := c.managedClustersClient.NewListByResourceGroupPager(resourceGroup, &armcontainerservice.ManagedClustersClientListByResourceGroupOptions{})

		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, cluster := range page.Value {
				if cluster == nil || cluster.Name == nil {
					continue
				}

				provisioningState := ""
				kubernetesVersion := ""
				currentKubernetesVersion := ""
				powerState := ""
				if cluster.Properties != nil {
					if cluster.Properties.ProvisioningState != nil {
						provisioningState = *cluster.Properties.ProvisioningState
					}
					if cluster.Properties.KubernetesVersion != nil {
						kubernetesVersion = *cluster.Properties.KubernetesVersion
					}
					if cluster.Properties.CurrentKubernetesVersion != nil {
						currentKubernetesVersion = *cluster.Properties.CurrentKubernetesVersion
					}
					if cluster.Properties.PowerState != nil && cluster.Properties.PowerState.Code != nil {
						powerState = string(*cluster.Properties.PowerState.Code)
					}
				}

				location := ""
				if cluster.Location != nil {
					location = *cluster.Location
				}

				skuTier := ""
				if cluster.SKU != nil && cluster.SKU.Tier != nil {
					skuTier = string(*cluster.SKU.Tier)
				}

				clusters = append(clusters, Cluster{
					Name:                     *cluster.Name,
					ResourceGroup:            resourceGroup,
					ProvisioningState:        provisioningState,
					Location:                 location,
					KubernetesVersion:        kubernetesVersion,
					CurrentKubernetesVersion: currentKubernetesVersion,
					SKUTier:                  skuTier,
					PowerState:               powerState,
				})
			}
		}
	}

	return clusters, nil
}

func (c *client) GetNodePools(ctx context.Context, clusterName string, resourceGroup string) ([]NodePool, error) {
	var nodePools []NodePool

	pager := c.agentPoolsClient.NewListPager(resourceGroup, clusterName, &armcontainerservice.AgentPoolsClientListOptions{})

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, nodePool := range page.Value {
			if nodePool != nil && nodePool.Properties != nil && nodePool.Name != nil {
				provisioningState := ""
				if nodePool.Properties.ProvisioningState != nil {
					provisioningState = *nodePool.Properties.ProvisioningState
				}

				count := int32(0)
				if nodePool.Properties.Count != nil {
					count = *nodePool.Properties.Count
				}

				minCount := int32(0)
				if nodePool.Properties.MinCount != nil {
					minCount = *nodePool.Properties.MinCount
				}

				maxCount := int32(0)
				if nodePool.Properties.MaxCount != nil {
					maxCount = *nodePool.Properties.MaxCount
				}

				vmSize := ""
				if nodePool.Properties.VMSize != nil {
					vmSize = *nodePool.Properties.VMSize
				}

				osType := ""
				if nodePool.Properties.OSType != nil {
					osType = string(*nodePool.Properties.OSType)
				}

				osSKU := ""
				if nodePool.Properties.OSSKU != nil {
					osSKU = string(*nodePool.Properties.OSSKU)
				}

				mode := ""
				if nodePool.Properties.Mode != nil {
					mode = string(*nodePool.Properties.Mode)
				}

				orchestratorVersion := ""
				if nodePool.Properties.OrchestratorVersion != nil {
					orchestratorVersion = *nodePool.Properties.OrchestratorVersion
				}

				currentOrchestratorVersion := ""
				if nodePool.Properties.CurrentOrchestratorVersion != nil {
					currentOrchestratorVersion = *nodePool.Properties.CurrentOrchestratorVersion
				}

				nodeImageVersion := ""
				if nodePool.Properties.NodeImageVersion != nil {
					nodeImageVersion = *nodePool.Properties.NodeImageVersion
				}

				scaleSetPriority := ""
				if nodePool.Properties.ScaleSetPriority != nil {
					scaleSetPriority = string(*nodePool.Properties.ScaleSetPriority)
				}

				nodePools = append(nodePools, NodePool{
					Name:                       *nodePool.Name,
					Cluster:                    clusterName,
					ResourceGroup:              resourceGroup,
					ProvisioningState:          provisioningState,
					Count:                      count,
					MinCount:                   minCount,
					MaxCount:                   maxCount,
					VMSize:                     vmSize,
					OSType:                     osType,
					OSSKU:                      osSKU,
					Mode:                       mode,
					OrchestratorVersion:        orchestratorVersion,
					CurrentOrchestratorVersion: currentOrchestratorVersion,
					NodeImageVersion:           nodeImageVersion,
					ScaleSetPriority:           scaleSetPriority,
				})
			}
		}
	}

	return nodePools, nil
}

func NewClient(config Config) (Client, error) {
	credentials, err := azidentity.NewClientSecretCredential(config.Credentials.TenantID, config.Credentials.ClientID, config.Credentials.ClientSecret, nil)
	if err != nil {
		return nil, err
	}

	managedClustersClient, err := armcontainerservice.NewManagedClustersClient(config.Credentials.SubscriptionID, credentials, &arm.ClientOptions{})
	if err != nil {
		return nil, err
	}

	agentPoolsClient, err := armcontainerservice.NewAgentPoolsClient(config.Credentials.SubscriptionID, credentials, &arm.ClientOptions{})
	if err != nil {
		return nil, err
	}

	return &client{
		subscriptionID:        config.Credentials.SubscriptionID,
		resourceGroups:        config.ResourceGroups,
		managedClustersClient: managedClustersClient,
		agentPoolsClient:      agentPoolsClient,
	}, nil
}
