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
	ClientSecret   string `json:"clientSecret"`
}

type Cluster struct {
	Name              string
	ResourceGroup     string
	ProvisioningState string
}

type NodePool struct {
	Name              string
	Cluster           string
	ResourceGroup     string
	ProvisioningState string
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
				clusters = append(clusters, Cluster{
					Name:              *cluster.Name,
					ResourceGroup:     resourceGroup,
					ProvisioningState: *cluster.Properties.ProvisioningState,
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
			nodePools = append(nodePools, NodePool{
				Name:              *nodePool.Name,
				Cluster:           clusterName,
				ResourceGroup:     resourceGroup,
				ProvisioningState: *nodePool.Properties.ProvisioningState,
			})
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

	return &client{
		subscriptionID:        config.Credentials.SubscriptionID,
		resourceGroups:        config.ResourceGroups,
		managedClustersClient: managedClustersClient,
		agentPoolsClient:      agentPoolsClient,
	}, nil
}
