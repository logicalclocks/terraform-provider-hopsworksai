package hopsworksai

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAzureUserAssignedIdentity_basic(t *testing.T) {
	dataSourceName := "data.hopsworksai_azure_user_assigned_identity_permissions.test"
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureUserAssignedIdentityConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "actions.#", "10"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.0", "Microsoft.Storage/storageAccounts/blobServices/containers/write"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.1", "Microsoft.Storage/storageAccounts/blobServices/containers/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.2", "Microsoft.Storage/storageAccounts/blobServices/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.3", "Microsoft.Storage/storageAccounts/blobServices/write"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.4", "Microsoft.Storage/storageAccounts/listKeys/action"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.5", "Microsoft.ContainerService/managedClusters/listClusterUserCredential/action"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.6", "Microsoft.ContainerService/managedClusters/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.7", "Microsoft.ContainerRegistry/registries/pull/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.8", "Microsoft.ContainerRegistry/registries/push/write"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.9", "Microsoft.ContainerRegistry/registries/artifacts/delete"),
					resource.TestCheckResourceAttr(dataSourceName, "not_actions.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "data_actions.#", "4"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "data_actions.*", "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/delete"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "data_actions.*", "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "data_actions.*", "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/move/action"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "data_actions.*", "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/write"),
					resource.TestCheckResourceAttr(dataSourceName, "not_data_actions.#", "0"),
				),
			},
		},
	})
}

func TestAccAzureUserAssignedIdentity_enableAKSandACROnly(t *testing.T) {
	dataSourceName := "data.hopsworksai_azure_user_assigned_identity_permissions.test"
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureUserAssignedIdentityConfig_enableAKSandACROnly(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "actions.#", "5"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.0", "Microsoft.ContainerService/managedClusters/listClusterUserCredential/action"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.1", "Microsoft.ContainerService/managedClusters/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.2", "Microsoft.ContainerRegistry/registries/pull/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.3", "Microsoft.ContainerRegistry/registries/push/write"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.4", "Microsoft.ContainerRegistry/registries/artifacts/delete"),
					resource.TestCheckResourceAttr(dataSourceName, "not_actions.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "data_actions.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "not_data_actions.#", "0"),
				),
			},
		},
	})
}

func TestAccAzureUserAssignedIdentity_enableACROnly(t *testing.T) {
	dataSourceName := "data.hopsworksai_azure_user_assigned_identity_permissions.test"
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureUserAssignedIdentityConfig_enableACROnly(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "actions.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.0", "Microsoft.ContainerRegistry/registries/pull/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.1", "Microsoft.ContainerRegistry/registries/push/write"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.2", "Microsoft.ContainerRegistry/registries/artifacts/delete"),
					resource.TestCheckResourceAttr(dataSourceName, "not_actions.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "data_actions.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "not_data_actions.#", "0"),
				),
			},
		},
	})
}

func TestAccAzureUserAssignedIdentity_disableBackup(t *testing.T) {
	dataSourceName := "data.hopsworksai_azure_user_assigned_identity_permissions.test"
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureUserAssignedIdentityConfig_disableBackup(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "actions.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.0", "Microsoft.Storage/storageAccounts/blobServices/containers/write"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.1", "Microsoft.Storage/storageAccounts/blobServices/containers/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.2", "Microsoft.Storage/storageAccounts/blobServices/read"),
					resource.TestCheckResourceAttr(dataSourceName, "not_actions.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "data_actions.#", "4"),
					resource.TestCheckResourceAttr(dataSourceName, "not_data_actions.#", "0"),
				),
			},
		},
	})
}

func TestAccAzureUserAssignedIdentity_disableAll(t *testing.T) {
	dataSourceName := "data.hopsworksai_azure_user_assigned_identity_permissions.test"
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureUserAssignedIdentityConfig_disableAll(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "actions.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "not_actions.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "data_actions.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "not_data_actions.#", "0"),
				),
			},
		},
	})
}

func testAccAzureUserAssignedIdentityConfig_basic() string {
	return `
	data "hopsworksai_azure_user_assigned_identity_permissions" "test" {
	}
	`
}

func testAccAzureUserAssignedIdentityConfig_enableAKSandACROnly() string {
	return `
	data "hopsworksai_azure_user_assigned_identity_permissions" "test" {
		enable_backup = false
		enable_storage = false
	}
	`
}

func testAccAzureUserAssignedIdentityConfig_disableBackup() string {
	return `
	data "hopsworksai_azure_user_assigned_identity_permissions" "test" {
		enable_backup = false
		enable_aks = false
		enable_acr = false
	}
	`
}

func testAccAzureUserAssignedIdentityConfig_disableAll() string {
	return `
	data "hopsworksai_azure_user_assigned_identity_permissions" "test" {
		enable_backup = false
		enable_storage = false
		enable_aks = false
		enable_acr = false
	}
	`
}

func testAccAzureUserAssignedIdentityConfig_enableACROnly() string {
	return `
	data "hopsworksai_azure_user_assigned_identity_permissions" "test" {
		enable_backup = false
		enable_storage = false
		enable_aks = false
	}
	`
}
