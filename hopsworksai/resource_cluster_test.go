package hopsworksai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
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

func TestAccClusterAWS_RonDB(t *testing.T) {
	testAccCluster_RonDB(t, api.AWS)
}

func TestAccClusterAZURE_RonDB(t *testing.T) {
	testAccCluster_RonDB(t, api.AZURE)
}

func TestAccClusterAWS_Autoscale(t *testing.T) {
	testAccCluster_Autoscale(t, api.AWS)
}

func TestAccClusterAZURE_Autoscale(t *testing.T) {
	testAccCluster_Autoscale(t, api.AZURE)
}

func TestAccClusterAWS_Autoscale_Update(t *testing.T) {
	testAccCluster_Autoscale_update(t, api.AWS)
}

func TestAccClusterAZURE_Autoscale_Update(t *testing.T) {
	testAccCluster_Autoscale_update(t, api.AZURE)
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
			{
				Config: testAccClusterConfig_workers(cloud, rName, suffix, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
					spot_config {
						max_price_percent = 10
					}
				}
				`, testWorkerInstanceType2(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type":                     testWorkerInstanceType2(cloud),
						"disk_size":                         "512",
						"count":                             "1",
						"spot_config.0.max_price_percent":   "10",
						"spot_config.0.fall_back_on_demand": "true",
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

func testAccCluster_RonDB(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	defaultRonDBConfig := defaultRonDBConfiguration(cloud)
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:     testAccPreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccClusterCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_RonDB(cloud, rName, suffix, `
				rondb {

				}
				`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rondb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.configuration.0.ndbd_default.0.replication_factor", strconv.Itoa(defaultRonDBConfig.Configuration.NdbdDefault.ReplicationFactor)),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.configuration.0.general.0.benchmark.0.grant_user_privileges", strconv.FormatBool(defaultRonDBConfig.Configuration.General.Benchmark.GrantUserPrivileges)),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.management_nodes.0.instance_type", defaultRonDBConfig.ManagementNodes.InstanceType),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.management_nodes.0.disk_size", strconv.Itoa(defaultRonDBConfig.ManagementNodes.DiskSize)),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.management_nodes.0.count", strconv.Itoa(defaultRonDBConfig.ManagementNodes.Count)),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.data_nodes.0.instance_type", defaultRonDBConfig.DataNodes.InstanceType),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.data_nodes.0.disk_size", strconv.Itoa(defaultRonDBConfig.DataNodes.DiskSize)),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.data_nodes.0.count", strconv.Itoa(defaultRonDBConfig.DataNodes.Count)),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.mysql_nodes.0.instance_type", defaultRonDBConfig.MYSQLNodes.InstanceType),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.mysql_nodes.0.disk_size", strconv.Itoa(defaultRonDBConfig.MYSQLNodes.DiskSize)),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.mysql_nodes.0.count", strconv.Itoa(defaultRonDBConfig.MYSQLNodes.Count)),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.api_nodes.0.instance_type", defaultRonDBConfig.APINodes.InstanceType),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.api_nodes.0.disk_size", strconv.Itoa(defaultRonDBConfig.APINodes.DiskSize)),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.api_nodes.0.count", strconv.Itoa(defaultRonDBConfig.APINodes.Count)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCluster_Autoscale(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	defaultAutoscaleConfig := defaultAutoscaleConfiguration()
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:     testAccPreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccClusterCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Autoscale(cloud, rName, suffix, fmt.Sprintf(`
				autoscale {
					non_gpu_workers {
						instance_type = "%s"
					}
				}
				`, testWorkerInstanceType1(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.instance_type", testWorkerInstanceType1(cloud)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.disk_size", strconv.Itoa(defaultAutoscaleConfig.DiskSize)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.min_workers", strconv.Itoa(defaultAutoscaleConfig.MinWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.max_workers", strconv.Itoa(defaultAutoscaleConfig.MaxWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.standby_workers", fmt.Sprint(defaultAutoscaleConfig.StandbyWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.downscale_wait_time", strconv.Itoa(defaultAutoscaleConfig.DownscaleWaitTime)),
				),
			},
			{
				Config: testAccClusterConfig_Autoscale(cloud, rName, suffix, fmt.Sprintf(`
				autoscale {
					non_gpu_workers {
						instance_type = "%s"
						spot_config {

						}
					}
				}
				`, testWorkerInstanceType1(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.instance_type", testWorkerInstanceType1(cloud)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.disk_size", strconv.Itoa(defaultAutoscaleConfig.DiskSize)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.min_workers", strconv.Itoa(defaultAutoscaleConfig.MinWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.max_workers", strconv.Itoa(defaultAutoscaleConfig.MaxWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.standby_workers", fmt.Sprint(defaultAutoscaleConfig.StandbyWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.downscale_wait_time", strconv.Itoa(defaultAutoscaleConfig.DownscaleWaitTime)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.spot_config.0.max_price_percent", strconv.Itoa(defaultSpotConfig().MaxPrice)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.spot_config.0.fall_back_on_demand", strconv.FormatBool(defaultSpotConfig().FallBackOnDemand)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_Autoscale(cloud, rName, suffix, fmt.Sprintf(`
				autoscale {
					non_gpu_workers {
						instance_type = "%s"
					}

					gpu_workers {
						instance_type = "%s"
					}
				}
				`, testWorkerInstanceType1(cloud), testWorkerInstanceTypeWithGPU(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.instance_type", testWorkerInstanceType1(cloud)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.disk_size", strconv.Itoa(defaultAutoscaleConfig.DiskSize)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.min_workers", strconv.Itoa(defaultAutoscaleConfig.MinWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.max_workers", strconv.Itoa(defaultAutoscaleConfig.MaxWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.standby_workers", fmt.Sprint(defaultAutoscaleConfig.StandbyWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.downscale_wait_time", strconv.Itoa(defaultAutoscaleConfig.DownscaleWaitTime)),

					resource.TestCheckResourceAttr(resourceName, "autoscale.0.gpu_workers.0.instance_type", testWorkerInstanceTypeWithGPU(cloud)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.gpu_workers.0.disk_size", strconv.Itoa(defaultAutoscaleConfig.DiskSize)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.gpu_workers.0.min_workers", strconv.Itoa(defaultAutoscaleConfig.MinWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.gpu_workers.0.max_workers", strconv.Itoa(defaultAutoscaleConfig.MaxWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.gpu_workers.0.standby_workers", fmt.Sprint(defaultAutoscaleConfig.StandbyWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.gpu_workers.0.downscale_wait_time", strconv.Itoa(defaultAutoscaleConfig.DownscaleWaitTime)),
				),
			},
		},
	})
}

func testAccCluster_Autoscale_update(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	defaultAutoscaleConfig := defaultAutoscaleConfiguration()
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:     testAccPreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccClusterCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Autoscale_Update(cloud, rName, suffix, fmt.Sprintf(`
				workers {
					instance_type = "%s"
				}
				`, testWorkerInstanceType1(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "autoscale.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_Autoscale_Update(cloud, rName, suffix, fmt.Sprintf(`
				autoscale {
					non_gpu_workers {
						instance_type = "%s"
					}
				}
				`, testWorkerInstanceType1(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.instance_type", testWorkerInstanceType1(cloud)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.disk_size", strconv.Itoa(defaultAutoscaleConfig.DiskSize)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.min_workers", strconv.Itoa(defaultAutoscaleConfig.MinWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.max_workers", strconv.Itoa(defaultAutoscaleConfig.MaxWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.standby_workers", fmt.Sprint(defaultAutoscaleConfig.StandbyWorkers)),
					resource.TestCheckResourceAttr(resourceName, "autoscale.0.non_gpu_workers.0.downscale_wait_time", strconv.Itoa(defaultAutoscaleConfig.DownscaleWaitTime)),
				),
			},
			{
				Config: testAccClusterConfig_Autoscale_Update(cloud, rName, suffix, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "autoscale.#", "0"),
				),
			},
			{
				Config: testAccClusterConfig_Autoscale_Update(cloud, rName, suffix, fmt.Sprintf(`
				workers {
					instance_type = "%s"
				}
				`, testWorkerInstanceType1(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "autoscale.#", "0"),
				),
			},
		},
	})
}

func TestAccClusterAWS_Head_upscale(t *testing.T) {
	testAccCluster_Head_upscale(t, api.AWS, "m5.2xlarge", "m5.4xlarge")
}

func TestAccClusterAZURE_Head_upscale(t *testing.T) {
	testAccCluster_Head_upscale(t, api.AZURE, "Standard_D8_v3", "Standard_D16_v3")
}

func testAccCluster_Head_upscale(t *testing.T, cloud api.CloudProvider, currentInstanceType string, newInstanceType string) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:     testAccPreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccClusterCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Head_upscale(cloud, rName, suffix, currentInstanceType, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "head.0.instance_type", currentInstanceType),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_Head_upscale(cloud, rName, suffix, currentInstanceType, `update_state = "stop"`),
				Check:  resource.TestCheckResourceAttr(resourceName, "state", api.Stopped.String()),
			},
			{
				Config: testAccClusterConfig_Head_upscale(cloud, rName, suffix, newInstanceType, ""),
				Check:  resource.TestCheckResourceAttr(resourceName, "head.0.instance_type", newInstanceType),
			},
		},
	})
}

func TestAccClusterAWS_RonDB_upscale(t *testing.T) {
	testAccCluster_RonDB_upscale(t, api.AWS, "t3a.xlarge", "t3a.2xlarge", "t3a.medium", "t3a.large", "t3a.medium", "t3a.large")
}

func TestAccClusterAZURE_RonDB_upscale(t *testing.T) {
	testAccCluster_RonDB_upscale(t, api.AZURE, "Standard_D4s_v4", "Standard_D8s_v4", "Standard_D2s_v4", "Standard_D4s_v4", "Standard_D2s_v4", "Standard_D4s_v4")
}

func testAccCluster_RonDB_upscale(t *testing.T, cloud api.CloudProvider, currentDataNodeType string, newDataNodeType string, currentMySQLNodeType string, newMySQLNodeType string, currentAPINodeType string, newAPINodeType string) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:     testAccPreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccClusterCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_RonDB_upscale(cloud, rName, suffix, fmt.Sprintf(`
				rondb {
					data_nodes {
						instance_type = "%s"
					}

					mysql_nodes {
						instance_type = "%s"
					}

					api_nodes {
						instance_type = "%s"
						count = 1
					}
				}
				`, currentDataNodeType, currentMySQLNodeType, currentAPINodeType)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.data_nodes.0.instance_type", currentDataNodeType),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.mysql_nodes.0.instance_type", currentMySQLNodeType),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.api_nodes.0.instance_type", currentAPINodeType),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.api_nodes.0.count", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_RonDB_upscale(cloud, rName, suffix, fmt.Sprintf(`
				rondb {
					data_nodes {
						instance_type = "%s"
					}

					mysql_nodes {
						instance_type = "%s"
					}

					api_nodes {
						instance_type = "%s"
						count = 1
					}
				}

				update_state = "stop"
				`, currentDataNodeType, currentMySQLNodeType, currentAPINodeType)),
				Check: resource.TestCheckResourceAttr(resourceName, "state", api.Stopped.String()),
			},
			{
				Config: testAccClusterConfig_RonDB_upscale(cloud, rName, suffix, fmt.Sprintf(`
				rondb {
					data_nodes {
						instance_type = "%s"
					}

					mysql_nodes {
						instance_type = "%s"
					}

					api_nodes {
						instance_type = "%s"
						count = 1
					}
				}
				`, newDataNodeType, newMySQLNodeType, newAPINodeType)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rondb.0.data_nodes.0.instance_type", newDataNodeType),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.mysql_nodes.0.instance_type", newMySQLNodeType),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.api_nodes.0.instance_type", newAPINodeType),
					resource.TestCheckResourceAttr(resourceName, "rondb.0.api_nodes.0.count", "1"),
				),
			},
		},
	})
}

func testWorkerInstanceTypeWithGPU(cloud api.CloudProvider) string {
	if cloud == api.AWS {
		return "g3s.xlarge"
	} else if cloud == api.AZURE {
		return "Standard_NC6"
	}
	return ""
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
			return "m5.xlarge"
		} else {
			return "m5.2xlarge"
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
	return testAccClusterConfig(cloud, rName, suffix, extraConfig, 2)
}

func testAccClusterConfig_workers(cloud api.CloudProvider, rName string, suffix string, extraConfig string) string {
	return testAccClusterConfig(cloud, rName, suffix, extraConfig, 3)
}

func testAccClusterConfig_RonDB(cloud api.CloudProvider, rName string, suffix string, extraConfig string) string {
	return testAccClusterConfig(cloud, rName, suffix, extraConfig, 4)
}

func testAccClusterConfig_Autoscale(cloud api.CloudProvider, rName string, suffix string, extraConfig string) string {
	return testAccClusterConfig(cloud, rName, suffix, extraConfig, 5)
}

func testAccClusterConfig_Autoscale_Update(cloud api.CloudProvider, rName string, suffix string, extraConfig string) string {
	return testAccClusterConfig(cloud, rName, suffix, extraConfig, 6)
}

func testAccClusterConfig_Head_upscale(cloud api.CloudProvider, rName string, suffix string, instanceType string, extraConfig string) string {
	return fmt.Sprintf(`
	resource "hopsworksai_cluster" "%s" {
		name    = "%s%s%s"
		ssh_key = "%s"	  
		head {
			instance_type = "%s"
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
		instanceType,
		testAccClusterCloudConfigAttributes(cloud, 11, false),
		extraConfig,
		default_CLUSTER_TAG_KEY,
		default_CLUSTER_TAG_VALUE,
	)
}

func testAccClusterConfig_RonDB_upscale(cloud api.CloudProvider, rName string, suffix string, extraConfig string) string {
	return testAccClusterConfig(cloud, rName, suffix, extraConfig, 12)
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
		testAccClusterCloudConfigAttributes(cloud, bucketIndex, false),
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
					if headInstanceType != awsDefaultInstanceType {
						return fmt.Errorf("expected default head instance type %s but got %s", awsDefaultInstanceType, headInstanceType)
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
					if headInstanceType != azureDefaultInstanceType {
						return fmt.Errorf("expected default head instance type %s but got %s", azureDefaultInstanceType, headInstanceType)
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
					"spot_config":   []interface{}{},
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
								"networkResourceGroup": "network-resource-group-1",
								"virtualNetworkName": "network-name-1",
								"subnetName": "subnet-name-1",
								"securityGroupName": "security-group-name-1",
								"searchDomain": "internal.cloudapp.net"
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
					"spot_config":   []interface{}{},
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
							"resource_group":       "network-resource-group-1",
							"virtual_network_name": "network-name-1",
							"subnet_name":          "subnet-name-1",
							"security_group_name":  "security-group-name-1",
							"search_domain":        "internal.cloudapp.net",
						},
					},
					"aks_cluster_name":  "",
					"acr_registry_name": "",
					"search_domain":     "internal.cloudapp.net",
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
							"version": "v1",
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
			"version": "v1",
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

func TestClusterCreate_AWS_RonDB(t *testing.T) {
	testClusterCreate_RonDB(t, api.AWS)
}

func TestClusterCreate_AZURE_RonDB(t *testing.T) {
	testClusterCreate_RonDB(t, api.AZURE)
}

func TestClusterCreate_AWS_RonDB_default(t *testing.T) {
	testClusterCreate_RonDB_default(t, api.AWS)
}

func TestClusterCreate_AZURE_RonDB_default(t *testing.T) {
	testClusterCreate_RonDB_default(t, api.AZURE)
}

func TestClusterCreate_AWS_RonDB_default2(t *testing.T) {
	testClusterCreate_RonDB_defaultEmptyBlocks(t, api.AWS)
}

func TestClusterCreate_AZURE_RonDB_default2(t *testing.T) {
	testClusterCreate_RonDB_defaultEmptyBlocks(t, api.AZURE)
}

func TestClusterCreate_RonDB_invalidReplicationFactor(t *testing.T) {
	testClusterCreate_RonDB_invalidReplicationFactor(t, api.AWS)
	testClusterCreate_RonDB_invalidReplicationFactor(t, api.AZURE)
}

func testClusterCreate_RonDB(t *testing.T, cloud api.CloudProvider) {
	state := map[string]interface{}{
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
		"rondb": []interface{}{
			map[string]interface{}{
				"configuration": []interface{}{
					map[string]interface{}{
						"ndbd_default": []interface{}{
							map[string]interface{}{
								"replication_factor": 2,
							},
						},
						"general": []interface{}{
							map[string]interface{}{
								"benchmark": []interface{}{
									map[string]interface{}{
										"grant_user_privileges": false,
									},
								},
							},
						},
					},
				},
				"management_nodes": []interface{}{
					map[string]interface{}{
						"instance_type": "mgm-node-1",
						"disk_size":     30,
						"count":         1,
					},
				},
				"data_nodes": []interface{}{
					map[string]interface{}{
						"instance_type": "data-node-1",
						"disk_size":     512,
						"count":         2,
					},
				},
				"mysql_nodes": []interface{}{
					map[string]interface{}{
						"instance_type": "mysqld-node-1",
						"disk_size":     100,
						"count":         1,
					},
				},
				"api_nodes": []interface{}{
					map[string]interface{}{
						"instance_type": "api-node-1",
						"disk_size":     50,
						"count":         1,
					},
				},
			},
		},
	}

	if cloud == api.AWS {
		state["aws_attributes"] = []interface{}{
			map[string]interface{}{
				"region":               "region-1",
				"bucket_name":          "bucket-1",
				"instance_profile_arn": "profile-1",
			},
		}
	} else if cloud == api.AZURE {
		state["azure_attributes"] = []interface{}{
			map[string]interface{}{
				"location":                       "location-1",
				"resource_group":                 "resource-group-1",
				"storage_account":                "storage-account-1",
				"user_assigned_managed_identity": "user-identity-1",
			},
		}
	}

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
					output, err := testGetRonDBConfig(reqBody, cloud)
					if err != nil {
						return err
					}

					expected := api.RonDBConfiguration{
						Configuration: api.RonDBBaseConfiguration{
							NdbdDefault: api.RonDBNdbdDefaultConfiguration{
								ReplicationFactor: 2,
							},
							General: api.RonDBGeneralConfiguration{
								Benchmark: api.RonDBBenchmarkConfiguration{
									GrantUserPrivileges: false,
								},
							},
						},
						ManagementNodes: api.WorkerConfiguration{
							NodeConfiguration: api.NodeConfiguration{
								InstanceType: "mgm-node-1",
								DiskSize:     30,
							},
							Count: 1,
						},
						DataNodes: api.WorkerConfiguration{
							NodeConfiguration: api.NodeConfiguration{
								InstanceType: "data-node-1",
								DiskSize:     512,
							},
							Count: 2,
						},
						MYSQLNodes: api.WorkerConfiguration{
							NodeConfiguration: api.NodeConfiguration{
								InstanceType: "mysqld-node-1",
								DiskSize:     100,
							},
							Count: 1,
						},
						APINodes: api.WorkerConfiguration{
							NodeConfiguration: api.NodeConfiguration{
								InstanceType: "api-node-1",
								DiskSize:     50,
							},
							Count: 1,
						},
					}

					if !reflect.DeepEqual(&expected, output) {
						return fmt.Errorf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State:                state,
		ExpectError:          "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}

func testGetRonDBConfig(reqBody io.Reader, cloud api.CloudProvider) (*api.RonDBConfiguration, error) {
	var output *api.RonDBConfiguration
	if cloud == api.AZURE {
		var req api.NewAzureClusterRequest
		if err := json.NewDecoder(reqBody).Decode(&req); err != nil {
			return nil, err
		}
		output = req.CreateRequest.RonDB
	} else if cloud == api.AWS {
		var req api.NewAWSClusterRequest
		if err := json.NewDecoder(reqBody).Decode(&req); err != nil {
			return nil, err
		}
		output = req.CreateRequest.RonDB
	}
	return output, nil
}

func testClusterCreate_RonDB_default(t *testing.T, cloud api.CloudProvider) {
	state := map[string]interface{}{
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
		"rondb": []interface{}{
			map[string]interface{}{},
		},
	}

	if cloud == api.AWS {
		state["aws_attributes"] = []interface{}{
			map[string]interface{}{
				"region":               "region-1",
				"bucket_name":          "bucket-1",
				"instance_profile_arn": "profile-1",
			},
		}
	} else if cloud == api.AZURE {
		state["azure_attributes"] = []interface{}{
			map[string]interface{}{
				"location":                       "location-1",
				"resource_group":                 "resource-group-1",
				"storage_account":                "storage-account-1",
				"user_assigned_managed_identity": "user-identity-1",
			},
		}
	}

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
					output, err := testGetRonDBConfig(reqBody, cloud)
					if err != nil {
						return err
					}
					expected := defaultRonDBConfiguration(cloud)
					if !reflect.DeepEqual(&expected, output) {
						return fmt.Errorf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State:                state,
		ExpectError:          "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}

func testClusterCreate_RonDB_defaultEmptyBlocks(t *testing.T, cloud api.CloudProvider) {
	state := map[string]interface{}{
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
		"rondb": []interface{}{
			map[string]interface{}{
				"configuration": []interface{}{
					map[string]interface{}{
						"ndbd_default": []interface{}{
							map[string]interface{}{},
						},
						"general": []interface{}{
							map[string]interface{}{
								"benchmark": []interface{}{
									map[string]interface{}{},
								},
							},
						},
					},
				},
				"management_nodes": []interface{}{
					map[string]interface{}{},
				},
				"data_nodes": []interface{}{
					map[string]interface{}{},
				},
				"mysql_nodes": []interface{}{
					map[string]interface{}{},
				},
				"api_nodes": []interface{}{
					map[string]interface{}{},
				},
			},
		},
	}

	if cloud == api.AWS {
		state["aws_attributes"] = []interface{}{
			map[string]interface{}{
				"region":               "region-1",
				"bucket_name":          "bucket-1",
				"instance_profile_arn": "profile-1",
			},
		}
	} else if cloud == api.AZURE {
		state["azure_attributes"] = []interface{}{
			map[string]interface{}{
				"location":                       "location-1",
				"resource_group":                 "resource-group-1",
				"storage_account":                "storage-account-1",
				"user_assigned_managed_identity": "user-identity-1",
			},
		}
	}

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
					output, err := testGetRonDBConfig(reqBody, cloud)
					if err != nil {
						return err
					}
					expected := defaultRonDBConfiguration(cloud)
					if !reflect.DeepEqual(&expected, output) {
						return fmt.Errorf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State:                state,
		ExpectError:          "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}

func testClusterCreate_RonDB_invalidReplicationFactor(t *testing.T, cloud api.CloudProvider) {
	state := map[string]interface{}{
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
		"rondb": []interface{}{
			map[string]interface{}{
				"configuration": []interface{}{
					map[string]interface{}{
						"ndbd_default": []interface{}{
							map[string]interface{}{
								"replication_factor": 2,
							},
						},
					},
				},
				"data_nodes": []interface{}{
					map[string]interface{}{
						"count": 3,
					},
				},
			},
		},
	}

	if cloud == api.AWS {
		state["aws_attributes"] = []interface{}{
			map[string]interface{}{
				"region":               "region-1",
				"bucket_name":          "bucket-1",
				"instance_profile_arn": "profile-1",
			},
		}
	} else if cloud == api.AZURE {
		state["azure_attributes"] = []interface{}{
			map[string]interface{}{
				"location":                       "location-1",
				"resource_group":                 "resource-group-1",
				"storage_account":                "storage-account-1",
				"user_assigned_managed_identity": "user-identity-1",
			},
		}
	}

	r := test.ResourceFixture{
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State:                state,
		ExpectError:          "number of RonDB data nodes must be multiples of RonDB replication factor",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_Autoscale(t *testing.T) {
	testClusterCreate_Autoscale(t, api.AWS, true)
	testClusterCreate_Autoscale(t, api.AZURE, true)
	testClusterCreate_Autoscale(t, api.AWS, false)
	testClusterCreate_Autoscale(t, api.AZURE, false)
}

func testClusterCreate_Autoscale(t *testing.T, cloud api.CloudProvider, withGpu bool) {
	state := map[string]interface{}{
		"name": "cluster",
		"head": []interface{}{
			map[string]interface{}{
				"disk_size": 512,
			},
		},
		"autoscale": []interface{}{
			map[string]interface{}{
				"non_gpu_workers": []interface{}{
					map[string]interface{}{
						"instance_type":       "non-gpu-node",
						"disk_size":           100,
						"min_workers":         0,
						"max_workers":         10,
						"standby_workers":     0.5,
						"downscale_wait_time": 200,
						"spot_config": []interface{}{
							map[string]interface{}{
								"max_price_percent": 10,
							},
						},
					},
				},
			},
		},
	}

	if withGpu {
		state["autoscale"].([]interface{})[0].(map[string]interface{})["gpu_workers"] = []interface{}{
			map[string]interface{}{
				"instance_type":       "gpu-node",
				"disk_size":           200,
				"min_workers":         1,
				"max_workers":         5,
				"standby_workers":     0.4,
				"downscale_wait_time": 100,
			},
		}
	}
	if cloud == api.AWS {
		state["aws_attributes"] = []interface{}{
			map[string]interface{}{
				"region":               "region-1",
				"bucket_name":          "bucket-1",
				"instance_profile_arn": "profile-1",
			},
		}
	} else if cloud == api.AZURE {
		state["azure_attributes"] = []interface{}{
			map[string]interface{}{
				"location":                       "location-1",
				"resource_group":                 "resource-group-1",
				"storage_account":                "storage-account-1",
				"user_assigned_managed_identity": "user-identity-1",
			},
		}
	}

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
					output, err := testGetAutoscaleConfig(reqBody, cloud)
					if err != nil {
						return err
					}

					expected := api.AutoscaleConfiguration{
						NonGPU: &api.AutoscaleConfigurationBase{
							InstanceType:      "non-gpu-node",
							DiskSize:          100,
							MinWorkers:        0,
							MaxWorkers:        10,
							StandbyWorkers:    0.5,
							DownscaleWaitTime: 200,
							SpotInfo: &api.SpotConfiguration{
								MaxPrice:         10,
								FallBackOnDemand: true,
							},
						},
					}
					if withGpu {
						expected.GPU = &api.AutoscaleConfigurationBase{
							InstanceType:      "gpu-node",
							DiskSize:          200,
							MinWorkers:        1,
							MaxWorkers:        5,
							StandbyWorkers:    0.4,
							DownscaleWaitTime: 100,
						}
					}
					if !reflect.DeepEqual(&expected, output) {
						return fmt.Errorf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State:                state,
		ExpectError:          "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}

func testGetAutoscaleConfig(reqBody io.Reader, cloud api.CloudProvider) (*api.AutoscaleConfiguration, error) {
	var output *api.AutoscaleConfiguration
	if cloud == api.AZURE {
		var req api.NewAzureClusterRequest
		if err := json.NewDecoder(reqBody).Decode(&req); err != nil {
			return nil, err
		}
		output = req.CreateRequest.Autoscale
	} else if cloud == api.AWS {
		var req api.NewAWSClusterRequest
		if err := json.NewDecoder(reqBody).Decode(&req); err != nil {
			return nil, err
		}
		output = req.CreateRequest.Autoscale
	}
	return output, nil
}

func TestClusterUpdate_upgrade(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"version": "v1",
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
			},
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/cluster-id-1/upgrade",
				ExpectRequestBody: `{
					"version": "v2"
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v2",
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate_upgrade_pipeline(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"state" : "pending",
							"provider": "AZURE",
							"version": "v1",
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
				RunOnlyOnce: true,
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
							"version": "v2",
							"upgradeInProgress": {
								"from": "v1",
								"to": "v2"
							},
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
				RunOnlyOnce: true,
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
							"version": "v2",
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
				RunOnlyOnce: true,
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
							"version": "v2",
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
				RunOnlyOnce: true,
			},
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/cluster-id-1/upgrade",
				ExpectRequestBody: `{
					"version": "v2"
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v2",
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate_upgrade_error(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"version": "v1",
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
			},
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/cluster-id-1/upgrade",
				ExpectRequestBody: `{
					"version": "v2"
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "failed to start upgrade"
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v2",
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
		ExpectError: "failed to start upgrade",
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate_upgrade_rollback(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"state" : "stopped",
							"provider": "AZURE",
							"version": "v2",
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/upgrade/rollback",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v1",
			"state":   api.Error.String(),
			"upgrade_in_progress": []interface{}{
				map[string]interface{}{
					"from_version": "v1",
					"to_version":   "v2",
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate_upgrade_rollback_externally_stopped(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"state" : "externally-stopped",
							"provider": "AZURE",
							"version": "v2",
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/upgrade/rollback",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v1",
			"state":   api.Error.String(),
			"upgrade_in_progress": []interface{}{
				map[string]interface{}{
					"from_version": "v1",
					"to_version":   "v2",
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate_upgrade_rollback_error(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"state" : "stopped",
							"provider": "AZURE",
							"version": "v2",
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/upgrade/rollback",
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "failed to rollback upgrade"
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v1",
			"state":   api.Error.String(),
			"upgrade_in_progress": []interface{}{
				map[string]interface{}{
					"from_version": "v1",
					"to_version":   "v2",
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
		ExpectError: "failed to rollback upgrade",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AZURE_container(t *testing.T) {
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
					if req.CreateRequest.BlobContainerName != "container-1" {
						return fmt.Errorf("error while matching:\nexpected container-1 \nbut got %#v", req.CreateRequest.BlobContainerName)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
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
					"storage_container_name":         "container-1",
				},
			},
		},
		ExpectError: "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AZURE_searchDomain_deprecated(t *testing.T) {
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
					if req.CreateRequest.SearchDomain != "my-domain.com" {
						return fmt.Errorf("error while matching:\nexpected my-domain.com \nbut got %#v", req.CreateRequest.SearchDomain)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
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
					"search_domain":                  "my-domain.com",
				},
			},
		},
		ExpectError: "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AZURE_searchDomain(t *testing.T) {
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
					if req.CreateRequest.SearchDomain != "my-domain.com" {
						return fmt.Errorf("error while matching:\nexpected my-domain.com \nbut got %#v", req.CreateRequest.SearchDomain)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
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
					"network": []interface{}{
						map[string]interface{}{
							"search_domain": "my-domain.com",
						},
					},
				},
			},
		},
		ExpectError: "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}

func TestClusterCreate_AZURE_ASK_cluster(t *testing.T) {
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
					if req.CreateRequest.AksClusterName != "aks-cluster-1" || req.CreateRequest.AcrRegistryName != "acr-registry-1" {
						return fmt.Errorf("error while matching:\nexpected aks-cluster-1, acr-registry-1 \nbut got %#v, %#v", req.CreateRequest.AksClusterName, req.CreateRequest.AcrRegistryName)
					}
					return nil
				},
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().CreateContext,
		State: map[string]interface{}{
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
					"aks_cluster_name":               "aks-cluster-1",
					"acr_registry_name":              "acr-registry-1",
				},
			},
		},
		ExpectError: "failed to create cluster, error: skip",
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate_modifyInstancetype_head(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"version": "v1",
							"clusterConfiguration": {
								"head": {
									"instanceType": "old-head-node-type-1",
									"diskSize": 512
								}
							},
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
				ExpectRequestBody: `{
					"nodeInfo": {
						"nodeType": "head",
						"instanceType": "new-head-node-type-1"
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v1",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "new-head-node-type-1",
					"disk_size":     512,
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate_modifyInstancetype_head_error(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"version": "v1",
							"clusterConfiguration": {
								"head": {
									"instanceType": "old-head-node-type-1",
									"diskSize": 512
								}
							},
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
				ExpectRequestBody: `{
					"nodeInfo": {
						"nodeType": "head",
						"instanceType": "new-head-node-type-1"
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "could not change instance type"
				}`,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v1",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "new-head-node-type-1",
					"disk_size":     512,
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
		ExpectError: "could not change instance type",
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate_modifyInstancetype_rondb(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"version": "v1",
							"ronDB": {
								"configuration": {
									"ndbdDefault": {
										"replicationFactor": 2
									},
									"general": {
										"benchmark": {
											"grantUserPrivileges": false
										}
									}
								},
								"mgmd": {
									"instanceType": "mgm-node-1",
									"diskSize": 30,
									"count": 1
								},
								"ndbd": {
									"instanceType": "data-node-1",
									"diskSize": 512,
									"count": 2
								},
								"mysqld": {
									"instanceType": "mysqld-node-1",
									"diskSize": 100,
									"count": 1
								},
								"api": {
									"instanceType": "api-node-1",
									"diskSize": 50,
									"count": 1
								}
							},
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
				ExpectRequestBody: `{
					"nodeInfo": {
						"nodeType": "rondb_data",
						"instanceType": "new-data-node-1"
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
				RunOnlyOnce: true,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
				ExpectRequestBody: `{
					"nodeInfo": {
						"nodeType": "rondb_mysql",
						"instanceType": "new-mysqld-node-1"
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
				RunOnlyOnce: true,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
				ExpectRequestBody: `{
					"nodeInfo": {
						"nodeType": "rondb_api",
						"instanceType": "new-api-node-1"
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
				RunOnlyOnce: true,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v1",
			"rondb": []interface{}{
				map[string]interface{}{
					"configuration": []interface{}{
						map[string]interface{}{
							"ndbd_default": []interface{}{
								map[string]interface{}{
									"replication_factor": 2,
								},
							},
							"general": []interface{}{
								map[string]interface{}{
									"benchmark": []interface{}{
										map[string]interface{}{
											"grant_user_privileges": false,
										},
									},
								},
							},
						},
					},
					"management_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "mgm-node-1",
							"disk_size":     30,
							"count":         1,
						},
					},
					"data_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-data-node-1",
							"disk_size":     512,
							"count":         2,
						},
					},
					"mysql_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-mysqld-node-1",
							"disk_size":     100,
							"count":         1,
						},
					},
					"api_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-api-node-1",
							"disk_size":     50,
							"count":         1,
						},
					},
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate_modifyInstancetype_rondb_data_error(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"version": "v1",
							"ronDB": {
								"configuration": {
									"ndbdDefault": {
										"replicationFactor": 2
									},
									"general": {
										"benchmark": {
											"grantUserPrivileges": false
										}
									}
								},
								"mgmd": {
									"instanceType": "mgm-node-1",
									"diskSize": 30,
									"count": 1
								},
								"ndbd": {
									"instanceType": "data-node-1",
									"diskSize": 512,
									"count": 2
								},
								"mysqld": {
									"instanceType": "mysqld-node-1",
									"diskSize": 100,
									"count": 1
								},
								"api": {
									"instanceType": "api-node-1",
									"diskSize": 50,
									"count": 1
								}
							},
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
				ExpectRequestBody: `{
					"nodeInfo": {
						"nodeType": "rondb_data",
						"instanceType": "new-data-node-1"
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "could not change rondb datanode instance type"
				}`,
				RunOnlyOnce: true,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v1",
			"rondb": []interface{}{
				map[string]interface{}{
					"configuration": []interface{}{
						map[string]interface{}{
							"ndbd_default": []interface{}{
								map[string]interface{}{
									"replication_factor": 2,
								},
							},
							"general": []interface{}{
								map[string]interface{}{
									"benchmark": []interface{}{
										map[string]interface{}{
											"grant_user_privileges": false,
										},
									},
								},
							},
						},
					},
					"management_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "mgm-node-1",
							"disk_size":     30,
							"count":         1,
						},
					},
					"data_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-data-node-1",
							"disk_size":     512,
							"count":         2,
						},
					},
					"mysql_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-mysqld-node-1",
							"disk_size":     100,
							"count":         1,
						},
					},
					"api_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-api-node-1",
							"disk_size":     50,
							"count":         1,
						},
					},
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
		ExpectError: "could not change rondb datanode instance type",
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate_modifyInstancetype_rondb_mysql_error(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"version": "v1",
							"ronDB": {
								"configuration": {
									"ndbdDefault": {
										"replicationFactor": 2
									},
									"general": {
										"benchmark": {
											"grantUserPrivileges": false
										}
									}
								},
								"mgmd": {
									"instanceType": "mgm-node-1",
									"diskSize": 30,
									"count": 1
								},
								"ndbd": {
									"instanceType": "data-node-1",
									"diskSize": 512,
									"count": 2
								},
								"mysqld": {
									"instanceType": "mysqld-node-1",
									"diskSize": 100,
									"count": 1
								},
								"api": {
									"instanceType": "api-node-1",
									"diskSize": 50,
									"count": 1
								}
							},
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
				ExpectRequestBody: `{
					"nodeInfo": {
						"nodeType": "rondb_data",
						"instanceType": "new-data-node-1"
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
				RunOnlyOnce: true,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
				ExpectRequestBody: `{
					"nodeInfo": {
						"nodeType": "rondb_mysql",
						"instanceType": "new-mysqld-node-1"
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "could not change rondb mysql node instance type"
				}`,
				RunOnlyOnce: true,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v1",
			"rondb": []interface{}{
				map[string]interface{}{
					"configuration": []interface{}{
						map[string]interface{}{
							"ndbd_default": []interface{}{
								map[string]interface{}{
									"replication_factor": 2,
								},
							},
							"general": []interface{}{
								map[string]interface{}{
									"benchmark": []interface{}{
										map[string]interface{}{
											"grant_user_privileges": false,
										},
									},
								},
							},
						},
					},
					"management_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "mgm-node-1",
							"disk_size":     30,
							"count":         1,
						},
					},
					"data_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-data-node-1",
							"disk_size":     512,
							"count":         2,
						},
					},
					"mysql_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-mysqld-node-1",
							"disk_size":     100,
							"count":         1,
						},
					},
					"api_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-api-node-1",
							"disk_size":     50,
							"count":         1,
						},
					},
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
		ExpectError: "could not change rondb mysql node instance type",
	}
	r.Apply(t, context.TODO())
}

func TestClusterUpdate_modifyInstancetype_rondb_api_error(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
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
							"version": "v1",
							"ronDB": {
								"configuration": {
									"ndbdDefault": {
										"replicationFactor": 2
									},
									"general": {
										"benchmark": {
											"grantUserPrivileges": false
										}
									}
								},
								"mgmd": {
									"instanceType": "mgm-node-1",
									"diskSize": 30,
									"count": 1
								},
								"ndbd": {
									"instanceType": "data-node-1",
									"diskSize": 512,
									"count": 2
								},
								"mysqld": {
									"instanceType": "mysqld-node-1",
									"diskSize": 100,
									"count": 1
								},
								"api": {
									"instanceType": "api-node-1",
									"diskSize": 50,
									"count": 1
								}
							},
							"ports":{
								"featureStore": false,
								"onlineFeatureStore": false,
								"kafka": false,
								"ssh": false
							}
						}
					}
				}`,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
				ExpectRequestBody: `{
					"nodeInfo": {
						"nodeType": "rondb_data",
						"instanceType": "new-data-node-1"
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
				RunOnlyOnce: true,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
				ExpectRequestBody: `{
					"nodeInfo": {
						"nodeType": "rondb_mysql",
						"instanceType": "new-mysqld-node-1"
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
				RunOnlyOnce: true,
			},
			{
				Method: http.MethodPut,
				Path:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
				ExpectRequestBody: `{
					"nodeInfo": {
						"nodeType": "rondb_api",
						"instanceType": "new-api-node-1"
					}
				}`,
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "could not change rondb api node instance type"
				}`,
				RunOnlyOnce: true,
			},
		},
		Resource:             clusterResource(),
		OperationContextFunc: clusterResource().UpdateContext,
		Id:                   "cluster-id-1",
		Update:               true,
		State: map[string]interface{}{
			"version": "v1",
			"rondb": []interface{}{
				map[string]interface{}{
					"configuration": []interface{}{
						map[string]interface{}{
							"ndbd_default": []interface{}{
								map[string]interface{}{
									"replication_factor": 2,
								},
							},
							"general": []interface{}{
								map[string]interface{}{
									"benchmark": []interface{}{
										map[string]interface{}{
											"grant_user_privileges": false,
										},
									},
								},
							},
						},
					},
					"management_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "mgm-node-1",
							"disk_size":     30,
							"count":         1,
						},
					},
					"data_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-data-node-1",
							"disk_size":     512,
							"count":         2,
						},
					},
					"mysql_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-mysqld-node-1",
							"disk_size":     100,
							"count":         1,
						},
					},
					"api_nodes": []interface{}{
						map[string]interface{}{
							"instance_type": "new-api-node-1",
							"disk_size":     50,
							"count":         1,
						},
					},
				},
			},
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  false,
					"kafka":                false,
					"feature_store":        false,
					"online_feature_store": false,
				},
			},
		},
		ExpectError: "could not change rondb api node instance type",
	}
	r.Apply(t, context.TODO())
}
