package hopsworksai

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGCPServiceAccountCustomRole_basic(t *testing.T) {
	dataSourceName := "data.hopsworksai_gcp_service_account_custom_role_permissions.test"
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGCPServiceAccountCustomRole_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "permissions.#", "17"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.0", "storage.buckets.get"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.1", "storage.multipartUploads.abort"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.2", "storage.multipartUploads.create"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.3", "storage.multipartUploads.list"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.4", "storage.multipartUploads.listParts"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.5", "storage.objects.create"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.6", "storage.objects.delete"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.7", "storage.objects.get"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.8", "storage.objects.list"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.9", "storage.objects.update"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.10", "storage.buckets.update"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.11", "artifactregistry.repositories.create"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.12", "artifactregistry.repositories.get"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.13", "artifactregistry.repositories.uploadArtifacts"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.14", "artifactregistry.repositories.downloadArtifacts"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.15", "artifactregistry.tags.list"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.16", "artifactregistry.tags.delete"),
				),
			},
		},
	})
}

func TestAccGCPServiceAccountCustomRole_noBackup(t *testing.T) {
	dataSourceName := "data.hopsworksai_gcp_service_account_custom_role_permissions.test"
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGCPServiceAccountCustomRole_noBackup(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "permissions.#", "16"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.0", "storage.buckets.get"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.1", "storage.multipartUploads.abort"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.2", "storage.multipartUploads.create"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.3", "storage.multipartUploads.list"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.4", "storage.multipartUploads.listParts"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.5", "storage.objects.create"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.6", "storage.objects.delete"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.7", "storage.objects.get"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.8", "storage.objects.list"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.9", "storage.objects.update"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.10", "artifactregistry.repositories.create"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.11", "artifactregistry.repositories.get"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.12", "artifactregistry.repositories.uploadArtifacts"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.13", "artifactregistry.repositories.downloadArtifacts"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.14", "artifactregistry.tags.list"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.15", "artifactregistry.tags.delete"),
				),
			},
		},
	})
}

func TestAccGCPServiceAccountCustomRole_noStorage(t *testing.T) {
	dataSourceName := "data.hopsworksai_gcp_service_account_custom_role_permissions.test"
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGCPServiceAccountCustomRole_noStorage(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "permissions.#", "6"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.0", "artifactregistry.repositories.create"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.1", "artifactregistry.repositories.get"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.2", "artifactregistry.repositories.uploadArtifacts"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.3", "artifactregistry.repositories.downloadArtifacts"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.4", "artifactregistry.tags.list"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions.5", "artifactregistry.tags.delete"),
				),
			},
		},
	})
}
func TestAccGCPServiceAccountCustomRole_noPerm(t *testing.T) {
	dataSourceName := "data.hopsworksai_gcp_service_account_custom_role_permissions.test"
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGCPServiceAccountCustomRole_noPerm(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "permissions.#", "0"),
				),
			},
		},
	})
}

func testAccGCPServiceAccountCustomRole_basic() string {
	return `
	data "hopsworksai_gcp_service_account_custom_role_permissions" "test" {
	}
	`
}

func testAccGCPServiceAccountCustomRole_noBackup() string {
	return `
	data "hopsworksai_gcp_service_account_custom_role_permissions" "test" {
		enable_backup = false
	}
	`
}

func testAccGCPServiceAccountCustomRole_noStorage() string {
	return `
	data "hopsworksai_gcp_service_account_custom_role_permissions" "test" {
		enable_backup = false
		enable_storage = false
	}
	`
}

func testAccGCPServiceAccountCustomRole_noPerm() string {
	return `
	data "hopsworksai_gcp_service_account_custom_role_permissions" "test" {
		enable_backup = false
		enable_storage = false
		enable_artifact_registry = false
	}
	`
}
