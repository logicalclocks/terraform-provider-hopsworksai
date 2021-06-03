package hopsworksai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/helpers"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/test"
)

func init() {
	resource.AddTestSweepers("hopsworksai_cluster", &resource.Sweeper{
		Name: "hopsworksai_cluster",
		F: func(region string) error {
			client := sharedClient()

			ctx := context.Background()
			clusters, err := api.GetClusters(ctx, client, "")
			if err != nil {
				return fmt.Errorf("Error getting clusters %s", err)
			}

			for _, cluster := range clusters {
				for _, tag := range cluster.Tags {
					if strings.HasPrefix(cluster.Name, default_CLUSTER_NAME_PREFIX) || (tag.Name == default_CLUSTER_TAG_KEY && tag.Value == default_CLUSTER_TAG_VALUE) {
						if err := api.DeleteCluster(ctx, client, cluster.Id); err != nil {
							log.Printf("Error destroying %s during sweep: %s", cluster.Id, err)
						}
						break
					}
				}
			}
			return nil
		},
	})
}

func TestAccClusterAWS_basic(t *testing.T) {
	testAccCluster_basic(t, api.AWS)
}

func TestAccClusterAZURE_basic(t *testing.T) {
	testAccCluster_basic(t, api.AZURE)
}

func TestAccClusterAWS_workers(t *testing.T) {
	testAccCluster_workers(t, api.AWS)
}

func TestAccClusterAZURE_workers(t *testing.T) {
	testAccCluster_workers(t, api.AZURE)
}

func testAccCluster_basic(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:     testAccPreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccClusterCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(cloud, rName, suffix, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.ssh", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.kafka", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.feature_store", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.online_feature_store", "false"),

					resource.TestCheckResourceAttr(resourceName, strings.ToLower(cloud.String())+"_attributes.0.network.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccClusterConfig_basic(cloud, rName, suffix, `update_state = "start"`),
				ExpectError: regexp.MustCompile("cluster is already running"),
			},
			{
				Config: testAccClusterConfig_basic(cloud, rName, suffix, `
				open_ports{
					ssh = true
					kafka = true
				}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "open_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.ssh", "true"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.kafka", "true"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.feature_store", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.online_feature_store", "false"),
				),
			},
			{
				Config: testAccClusterConfig_basic(cloud, rName, suffix, `update_state = "stop"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Stopped.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Startable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "stop"),
				),
			},
			{
				Config: testAccClusterConfig_basic(cloud, rName, suffix, `
				open_ports{
					feature_store = true
					online_feature_store = true
				}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "open_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.ssh", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.kafka", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.feature_store", "true"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.online_feature_store", "true"),
				),
			},
			{
				Config:      testAccClusterConfig_basic(cloud, rName, suffix, `update_state = "stop"`),
				ExpectError: regexp.MustCompile("cluster is already stopped"),
			},
			{
				Config: testAccClusterConfig_basic(cloud, rName, suffix, `update_state = "start"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "start"),
				),
			},
		},
	})
}

func testAccCluster_workers(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:     testAccPreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccClusterCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_workers(cloud, rName, suffix, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 256
					count = 2
				}`, testWorkerInstanceType1(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "256",
						"count":         "2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_workers(cloud, rName, suffix, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 256
					count = 1
				}
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				`, testWorkerInstanceType1(cloud), testWorkerInstanceType1(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "256",
						"count":         "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
				),
			},
			{
				Config: testAccClusterConfig_workers(cloud, rName, suffix, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 256
					count = 1
				}
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				`, testWorkerInstanceType1(cloud), testWorkerInstanceType1(cloud), testWorkerInstanceType2(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "256",
						"count":         "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType2(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
				),
			},
			{
				Config: testAccClusterConfig_workers(cloud, rName, suffix, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				`, testWorkerInstanceType2(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType2(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
				),
			},
			{
				Config: testAccClusterConfig_workers(cloud, rName, suffix, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 2
				}
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				`, testWorkerInstanceType2(cloud), testWorkerInstanceType1(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType2(cloud),
						"disk_size":     "512",
						"count":         "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
				),
			},
			{
				Config: testAccClusterConfig_workers(cloud, rName, suffix, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				`, testWorkerInstanceType2(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType2(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
				),
			},
		},
	})
}

func testWorkerInstanceType1(cloud api.CloudProvider) string {
	return testWorkerInstanceType(cloud, true)
}

func testWorkerInstanceType2(cloud api.CloudProvider) string {
	return testWorkerInstanceType(cloud, false)
}

func testWorkerInstanceType(cloud api.CloudProvider, alternative bool) string {
	if cloud == api.AWS {
		if alternative {
			return "t3a.medium"
		} else {
			return "t3a.large"
		}
	} else if cloud == api.AZURE {
		if alternative {
			return "Standard_D4_v3"
		} else {
			return "Standard_D8_v3"
		}
	}
	return ""
}

func testAccClusterCheckDestroy() func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*api.HopsworksAIClient)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "hopsworksai_cluster" {
				continue
			}
			cluster, err := api.GetCluster(context.Background(), client, rs.Primary.ID)
			if err != nil {
				return err
			}

			if cluster != nil {
				return fmt.Errorf("found unterminated cluster %s", rs.Primary.ID)
			}
		}
		return nil
	}
}

func testAccClusterConfig_basic(cloud api.CloudProvider, rName string, suffix string, extraConfig string) string {
	return testAccClusterConfig(cloud, rName, suffix, extraConfig, 0)
}

func testAccClusterConfig_workers(cloud api.CloudProvider, rName string, suffix string, extraConfig string) string {
	return testAccClusterConfig(cloud, rName, suffix, extraConfig, 1)
}

func testAccClusterConfig(cloud api.CloudProvider, rName string, suffix string, extraConfig string, bucketIndex int) string {
	return fmt.Sprintf(`
	resource "hopsworksai_cluster" "%s" {
		name    = "%s%s%s"
		ssh_key = "%s"	  
		head {
		}
		
		%s
		
		%s 

		tags = {
		  "%s" = "%s"
		}
	  }
	`,
		rName,
		default_CLUSTER_NAME_PREFIX,
		strings.ToLower(cloud.String()),
		suffix,
		testAccClusterCloudSSHKeyAttribute(cloud),
		testAccClusterCloudConfigAttributes(cloud, bucketIndex),
		extraConfig,
		default_CLUSTER_TAG_KEY,
		default_CLUSTER_TAG_VALUE,
	)
}

// Unit tests

func TestClusterCreate_AWS(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"id" : "new-cluster-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/new-cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id" : "new-cluster-id-1",
							"state": "running"
						}
					}
				}`,
			},
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/new-cluster-id-1/ports",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name":    "cluster-1",
			"version": "2.0",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			},
			"ssh_key": "ssh-key-1",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"region":               "region-1",
					"bucket_name":          "bucket-1",
					"instance_profile_arn": "profile-1",
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  true,
					"kafka":                true,
					"feature_store":        true,
					"online_feature_store": true,
				},
			},
		},
		ExpectId: "new-cluster-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AWSSetNetwork(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"id" : "new-cluster-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/new-cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id" : "new-cluster-id-1",
							"state": "running"
						}
					}
				}`,
			},
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/new-cluster-id-1/ports",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name":    "cluster-1",
			"version": "2.0",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			},
			"ssh_key": "ssh-key-1",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"region":               "region-1",
					"bucket_name":          "bucket-1",
					"instance_profile_arn": "profile-1",
					"network": []interface{}{
						map[string]interface{}{
							"vpc_id":            "vpc-id-1",
							"subnet_id":         "subnet-id-1",
							"security_group_id": "security-group-id-1",
						},
					},
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  true,
					"kafka":                true,
					"feature_store":        true,
					"online_feature_store": true,
				},
			},
		},
		ExpectId: "new-cluster-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AWS_errorOpenPorts(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"id" : "new-cluster-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/new-cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id" : "new-cluster-id-1",
							"state": "running"
						}
					}
				}`,
			},
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/new-cluster-id-1/ports",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 400,
					"message": "could not open ports"
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name":    "cluster-1",
			"version": "2.0",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"region":               "region-1",
					"bucket_name":          "bucket-1",
					"instance_profile_arn": "profile-1",
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  true,
					"kafka":                true,
					"feature_store":        true,
					"online_feature_store": true,
				},
			},
		},
		ExpectError: "failed to open ports on cluster, error: could not open ports",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AWSInvalidName(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"id" : "new-cluster-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/new-cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id" : "new-cluster-id-1",
							"state": "running"
						}
					}
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name":    "cluster-1#",
			"version": "2.0",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			},
			"ssh_key": "ssh-key-1",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"region":               "region-1",
					"bucket_name":          "bucket-1",
					"instance_profile_arn": "profile-1",
				},
			},
		},
		ExpectError: "invalid value for name, cluster name can only include a-z, A-Z, 0-9, _, - and a maximum of 20 characters",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AWS_errorWaiting(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"id" : "new-cluster-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/new-cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 400,
					"message": "failed while waiting"
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name":    "cluster-1",
			"version": "2.0",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"region":               "region-1",
					"bucket_name":          "bucket-1",
					"instance_profile_arn": "profile-1",
				},
			},
		},
		ExpectError: "failed while waiting",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AZURE(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"id" : "new-cluster-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/new-cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id" : "new-cluster-id-1",
							"state": "running"
						}
					}
				}`,
			},
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/new-cluster-id-1/ports",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name":    "cluster",
			"version": "2.0",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			},
			"ssh_key": "ssh-key-1",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
			"azure_attributes": []interface{}{
				map[string]interface{}{
					"location":                       "location-1",
					"resource_group":                 "resource-group-1",
					"storage_account":                "storage-account-1",
					"user_assigned_managed_identity": "user-identity-1",
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  true,
					"kafka":                true,
					"feature_store":        true,
					"online_feature_store": true,
				},
			},
		},
		ExpectId: "new-cluster-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AZURESetNetwork(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"id" : "new-cluster-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/new-cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id" : "new-cluster-id-1",
							"state": "running"
						}
					}
				}`,
			},
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/new-cluster-id-1/ports",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name":    "cluster",
			"version": "2.0",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			},
			"ssh_key": "ssh-key-1",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
			"azure_attributes": []interface{}{
				map[string]interface{}{
					"location":                       "location-1",
					"resource_group":                 "resource-group-1",
					"storage_account":                "storage-account-1",
					"user_assigned_managed_identity": "user-identity-1",
					"network": []interface{}{
						map[string]interface{}{
							"virtual_network_name": "virtual-network-name-1",
							"subnet_name":          "subnet-name-1",
							"security_group_name":  "security-group-name-1",
						},
					},
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  true,
					"kafka":                true,
					"feature_store":        true,
					"online_feature_store": true,
				},
			},
		},
		ExpectId: "new-cluster-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AZUREInvalidName(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"id" : "new-cluster-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/new-cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id" : "new-cluster-id-1",
							"state": "running"
						}
					}
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name":    "cluster-1",
			"version": "2.0",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			},
			"ssh_key": "ssh-key-1",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
			"azure_attributes": []interface{}{
				map[string]interface{}{
					"location":                       "location-1",
					"resource_group":                 "resource-group-1",
					"storage_account":                "storage-account-1",
					"user_assigned_managed_identity": "user-identity-1",
				},
			},
		},
		ExpectError: "invalid value for name, cluster name can only include a-z, 0-9 and a maximum of 20 characters",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_updateState(t *testing.T) {
	r := test.ResourceFixture{
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"update_state": "start",
			"name":         "cluster-1",
			"version":      "2.0",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			},
			"ssh_key": "ssh-key-1",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"region":               "region-1",
					"bucket_name":          "bucket-1",
					"instance_profile_arn": "profile-1",
				},
			},
		},
		ExpectError: "you cannot update cluster state during creation",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_noCloudConfiguration(t *testing.T) {
	r := test.ResourceFixture{
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name":    "cluster-1",
			"version": "2.0",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			},
			"ssh_key": "ssh-key-1",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
		},
		ExpectError: "no request to create cluster",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_error(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 400,
					"message": "cannot create cluster"
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name":    "cluster",
			"version": "2.0",
			"head": []interface{}{
				map[string]interface{}{
					"disk_size": 512,
				},
			},
			"azure_attributes": []interface{}{
				map[string]interface{}{
					"location":                       "location-1",
					"resource_group":                 "resource-group-1",
					"storage_account":                "storage-account-1",
					"user_assigned_managed_identity": "user-identity-1",
				},
			},
		},
		ExpectError: "failed to create cluster, error: cannot create cluster",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AWSDefaultInstanceType(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 400,
					"message": "skip"
				}`,
				CheckRequestBody: func(reqBody io.Reader) error {
					var req api.NewAWSClusterRequest
					if err := json.NewDecoder(reqBody).Decode(&req); err != nil {
						return err
					}
					headInstanceType := req.CreateRequest.ClusterConfiguration.Head.InstanceType
					workerInstanceType := req.CreateRequest.ClusterConfiguration.Workers[0].InstanceType
					if headInstanceType != awsDefaultInstanceType {
						return fmt.Errorf("expected default head instance type %s but got %s", awsDefaultInstanceType, headInstanceType)
					}
					if workerInstanceType != awsDefaultInstanceType {
						return fmt.Errorf("expected default worker instance type %s but got %s", awsDefaultInstanceType, workerInstanceType)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name": "cluster",
			"head": []interface{}{
				map[string]interface{}{
					"disk_size": 512,
				},
			},
			"workers": []interface{}{
				map[string]interface{}{
					"disk_size": 256,
					"count":     2,
				},
			},
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"region":               "region-1",
					"bucket_name":          "bucket-1",
					"instance_profile_arn": "profile-1",
				},
			},
		},
		ExpectError: "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AzureDefaultInstanceType(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 400,
					"message": "skip"
				}`,
				CheckRequestBody: func(reqBody io.Reader) error {
					var req api.NewAzureClusterRequest
					if err := json.NewDecoder(reqBody).Decode(&req); err != nil {
						return err
					}
					headInstanceType := req.CreateRequest.ClusterConfiguration.Head.InstanceType
					workerInstanceType := req.CreateRequest.ClusterConfiguration.Workers[0].InstanceType
					if headInstanceType != azureDefaultInstanceType {
						return fmt.Errorf("expected default head instance type %s but got %s", azureDefaultInstanceType, headInstanceType)
					}
					if workerInstanceType != azureDefaultInstanceType {
						return fmt.Errorf("expected default worker instance type %s but got %s", azureDefaultInstanceType, workerInstanceType)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name": "cluster",
			"head": []interface{}{
				map[string]interface{}{
					"disk_size": 512,
				},
			},
			"workers": []interface{}{
				map[string]interface{}{
					"disk_size": 256,
					"count":     2,
				},
			},
			"azure_attributes": []interface{}{
				map[string]interface{}{
					"location":                       "location-1",
					"resource_group":                 "resource-group-1",
					"storage_account":                "storage-account-1",
					"user_assigned_managed_identity": "user-identity-1",
				},
			},
		},
		ExpectError: "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AWS_defaultECRAccountId(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 400,
					"message": "skip"
				}`,
				CheckRequestBody: func(reqBody io.Reader) error {
					var req api.NewAWSClusterRequest
					if err := json.NewDecoder(reqBody).Decode(&req); err != nil {
						return err
					}
					ecr := req.CreateRequest.EcrRegistryAccountId
					if ecr != "000011111333" {
						return fmt.Errorf("expected ecr account id 000011111333 but got %s", ecr)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name": "cluster",
			"head": []interface{}{
				map[string]interface{}{
					"disk_size": 512,
				},
			},
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"region":               "region-1",
					"bucket_name":          "bucket-1",
					"instance_profile_arn": "arn:aws:iam::000011111333:instance-profile/my-instance-profile",
					"eks_cluster_name":     "my-cluster",
				},
			},
		},
		ExpectError: "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AWS_setECRAccountId(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 400,
					"message": "skip"
				}`,
				CheckRequestBody: func(reqBody io.Reader) error {
					var req api.NewAWSClusterRequest
					if err := json.NewDecoder(reqBody).Decode(&req); err != nil {
						return err
					}
					ecr := req.CreateRequest.EcrRegistryAccountId
					if ecr != "000011111444" {
						return fmt.Errorf("expected ecr account id 000011111444 but got %s", ecr)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
			"name": "cluster",
			"head": []interface{}{
				map[string]interface{}{
					"disk_size": 512,
				},
			},
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"region":                  "region-1",
					"bucket_name":             "bucket-1",
					"instance_profile_arn":    "arn:aws:iam::000011111333:instance-profile/my-instance-profile",
					"eks_cluster_name":        "my-cluster",
					"ecr_registry_account_id": "000011111444",
				},
			},
		},
		ExpectError: "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}
func TestClusterRead_AWS(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"statue": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id": "cluster-id-1",
							"name": "cluster-name-1",
							"state" : "running", 
							"activationState": "stoppable", 
							"initializationStage": "running", 
							"createdOn": 123, 
							"startedOn" : 123,
							"version": "version-1",
							"url": "https://cluster-url",
							"provider": "AWS",
							"tags": [
								{
									"name": "tag1",
									"value": "tag1-value1"
								}
							],
							"sshKeyName": "ssh-key-1",
							"clusterConfiguration": {
								"head": {
									"instanceType": "node-type-1",
									"diskSize": 512
								},
								"workers": [
									{
										"instanceType": "node-type-2",
										"diskSize": 256,
										"count": 2
									}
								]
							},
							"publicIPAttached": true,
							"letsEncryptIssued": true,
							"managedUsers": true,
							"backupRetentionPeriod": 10,
							"aws": {
								"region": "region-1",
								"instanceProfileArn": "profile-1",
								"bucketName": "bucket-1",
								"vpcId": "vpc-1",
								"subnetId": "subnet-1",
								"securityGroupId": "security-group-1"
							}
						}
					}
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().ReadContext,
		Id:                   "cluster-id-1",
		ExpectState: map[string]interface{}{
			"cluster_id":       "cluster-id-1",
			"state":            "running",
			"activation_state": "stoppable",
			"creation_date":    time.Unix(123, 0).Format(time.RFC3339),
			"start_date":       time.Unix(123, 0).Format(time.RFC3339),
			"version":          "version-1",
			"url":              "https://cluster-url",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
			"ssh_key": "ssh-key-1",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": schema.NewSet(helpers.WorkerSetHash, []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			}),
			"attach_public_ip":               true,
			"issue_lets_encrypt_certificate": true,
			"managed_users":                  true,
			"backup_retention_period":        10,
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"region":               "region-1",
					"bucket_name":          "bucket-1",
					"instance_profile_arn": "profile-1",
					"network": []interface{}{
						map[string]interface{}{
							"vpc_id":            "vpc-1",
							"subnet_id":         "subnet-1",
							"security_group_id": "security-group-1",
						},
					},
					"eks_cluster_name":        "",
					"ecr_registry_account_id": "",
				},
			},
			"azure_attributes": []interface{}{},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClusterRead_AZURE(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"statue": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id": "cluster-id-1",
							"name": "cluster-name-1",
							"state" : "running", 
							"activationState": "stoppable", 
							"initializationStage": "running", 
							"createdOn": 123, 
							"startedOn" : 123,
							"version": "version-1",
							"url": "https://cluster-url",
							"provider": "AZURE",
							"tags": [
								{
									"name": "tag1",
									"value": "tag1-value1"
								}
							],
							"sshKeyName": "ssh-key-1",
							"clusterConfiguration": {
								"head": {
									"instanceType": "node-type-1",
									"diskSize": 512
								},
								"workers": [
									{
										"instanceType": "node-type-2",
										"diskSize": 256,
										"count": 2
									}
								]
							},
							"publicIPAttached": true,
							"letsEncryptIssued": true,
							"managedUsers": true,
							"backupRetentionPeriod": 10,
							"azure": {
								"location": "location-1",
								"resourceGroup": "resource-group-1",
								"managedIdentity": "profile-1",
								"blobContainerName": "container-1",
								"storageAccount": "account-1",
								"virtualNetworkName": "network-name-1",
								"subnetName": "subnet-name-1",
								"securityGroupName": "security-group-name-1"
							}
						}
					}
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().ReadContext,
		Id:                   "cluster-id-1",
		ExpectState: map[string]interface{}{
			"cluster_id":       "cluster-id-1",
			"state":            "running",
			"activation_state": "stoppable",
			"creation_date":    time.Unix(123, 0).Format(time.RFC3339),
			"start_date":       time.Unix(123, 0).Format(time.RFC3339),
			"version":          "version-1",
			"url":              "https://cluster-url",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
			"ssh_key": "ssh-key-1",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": schema.NewSet(helpers.WorkerSetHash, []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			}),
			"attach_public_ip":               true,
			"issue_lets_encrypt_certificate": true,
			"managed_users":                  true,
			"backup_retention_period":        10,
			"azure_attributes": []interface{}{
				map[string]interface{}{
					"location":                       "location-1",
					"resource_group":                 "resource-group-1",
					"storage_account":                "account-1",
					"user_assigned_managed_identity": "profile-1",
					"storage_container_name":         "container-1",
					"network": []interface{}{
						map[string]interface{}{
							"virtual_network_name": "network-name-1",
							"subnet_name":          "subnet-name-1",
							"security_group_name":  "security-group-name-1",
						},
					},
					"aks_cluster_name":  "",
					"acr_registry_name": "",
				},
			},
			"aws_attributes": []interface{}{},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClusterRead_error(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"statue": "ok",
					"code": 400,
					"message": "cannot read cluster"
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().ReadContext,
		Id:                   "cluster-id-1",
		ExpectError:          "failed to obtain cluster state: cannot read cluster",
	}
	r.Apply(t, context.TODO())
}

func TestClusterDelete(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodDelete,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 404
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().DeleteContext,
		Id:                   "cluster-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/cluster-id-1/workers",
				ExpectRequestBody: `{
					"workers":[
						{
							"instanceType": "node-type-2",
							"diskSize": 512,
							"count": 1
						}
					]
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
			{
				Method: http.MethodDelete,
				Path:   "/api/clusters/cluster-id-1/workers",
				ExpectRequestBody: `{
					"workers":[
						{
							"instanceType": "node-type-2",
							"diskSize": 256,
							"count": 1
						}
					]
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/cluster-id-1/ports",
				ExpectRequestBody: `{
					"ports":{
						"featureStore": false,
						"onlineFeatureStore": false,
						"kafka": false,
						"ssh": false
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id": "cluster-id-1",
							"name": "cluster-name-1",
							"state" : "running",
							"provider": "AZURE",
							"clusterConfiguration": {
								"head": {
									"instanceType": "node-type-1",
									"diskSize": 512
								},
								"workers": [
									{
										"instanceType": "node-type-2",
										"diskSize": 256,
										"count": 2
									}
								]
							}
						}
					}
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"workers": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     512,
					"count":         1,
				},
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         1,
				},
			},
		},
	}
	r.Apply(t, context.TODO())
}
