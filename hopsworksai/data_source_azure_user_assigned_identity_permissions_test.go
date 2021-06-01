package hopsworksai

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAzureUserAssignedIdentity_basic(t *testing.T) {
	dataSourceName := "data.hopsworksai_azure_user_assigned_identity_permissions.test"
	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureUserAssignedIdentityConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "actions.#", "4"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.0", "Microsoft.Storage/storageAccounts/blobServices/containers/write"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.1", "Microsoft.Storage/storageAccounts/blobServices/containers/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.2", "Microsoft.Storage/storageAccounts/blobServices/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.3", "Microsoft.Storage/storageAccounts/blobServices/write"),
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

func TestAccAzureUserAssignedIdentity_enableUpgradeOnly(t *testing.T) {
	dataSourceName := "data.hopsworksai_azure_user_assigned_identity_permissions.test"
	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureUserAssignedIdentityConfig_enableUpgradeOnly(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "actions.#", "4"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.0", "Microsoft.Compute/virtualMachines/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.1", "Microsoft.Compute/virtualMachines/write"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.2", "Microsoft.Compute/disks/read"),
					resource.TestCheckResourceAttr(dataSourceName, "actions.3", "Microsoft.Compute/disks/write"),
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
		Providers: testAccProviders,
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
		Providers: testAccProviders,
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

func testAccAzureUserAssignedIdentityConfig_enableUpgradeOnly() string {
	return `
	data "hopsworksai_azure_user_assigned_identity_permissions" "test" {
		enable_upgrade = true
		enable_backup = false
		enable_storage = false
	}
	`
}

func testAccAzureUserAssignedIdentityConfig_disableBackup() string {
	return `
	data "hopsworksai_azure_user_assigned_identity_permissions" "test" {
		enable_backup = false
	}
	`
}

func testAccAzureUserAssignedIdentityConfig_disableAll() string {
	return `
	data "hopsworksai_azure_user_assigned_identity_permissions" "test" {
		enable_upgrade = false
		enable_backup = false
		enable_storage = false
	}
	`
}
