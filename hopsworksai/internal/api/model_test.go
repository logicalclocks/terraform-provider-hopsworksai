package api

import (
	"fmt"
	"reflect"
	"testing"
)

func TestIsAWSCluster(t *testing.T) {
	cluster := &Cluster{
		Provider: AWS,
	}
	if !cluster.IsAWSCluster() {
		t.Fatal("is aws cluster should return true")
	}
	cluster.Provider = AZURE
	if cluster.IsAWSCluster() {
		t.Fatal("is aws cluster should return false")
	}
	cluster.Provider = ""
	if cluster.IsAWSCluster() {
		t.Fatal("is aws cluster should return false")
	}
}

func TestIsAZURECluster(t *testing.T) {
	cluster := &Cluster{
		Provider: AZURE,
	}
	if !cluster.IsAzureCluster() {
		t.Fatal("is azure cluster should return true")
	}
	cluster.Provider = AWS
	if cluster.IsAzureCluster() {
		t.Fatal("is azure cluster should return false")
	}
	cluster.Provider = ""
	if cluster.IsAzureCluster() {
		t.Fatal("is azure cluster should return false")
	}
}

func TestIsGCPCluster(t *testing.T) {
	cluster := &Cluster{
		Provider: GCP,
	}
	if !cluster.IsGCPCluster() {
		t.Fatal("is gcp cluster should return true")
	}
	cluster.Provider = AWS
	if cluster.IsGCPCluster() {
		t.Fatal("is gcp cluster should return false")
	}
	cluster.Provider = ""
	if cluster.IsGCPCluster() {
		t.Fatal("is gcp cluster should return false")
	}
}

func TestValidateResponse(t *testing.T) {
	resp := BaseResponse{}

	for _, code := range []int{200, 201, 202} {
		resp.Code = code
		if err := resp.validate(); err != nil {
			t.Fatalf("validate response should not throw an error, but got %s", err)
		}
	}

	for _, code := range []int{300, 400, 500} {
		resp.Code = code
		resp.Message = fmt.Sprintf("messagee for %d", code)
		if err := resp.validate(); err == nil {
			t.Fatal("validate response should throw an error")
		} else if err.Error() != resp.Message {
			t.Fatalf("validate response should throw an error with message: %s, but got %s", resp.Message, err)
		}
	}
}

func TestGetSupportedNodesByNodeType(t *testing.T) {
	types := SupportedInstanceTypes{
		Head: SupportedInstanceTypeList{
			{
				Id:     "head-type-1",
				CPUs:   10,
				Memory: 30,
			},
		},
		Worker: SupportedInstanceTypeList{
			{
				Id:     "worker-type-1",
				CPUs:   16,
				Memory: 100,
			},
		},
		RonDB: SupportedRonDBInstanceTypes{
			ManagementNode: SupportedInstanceTypeList{
				{
					Id:     "rondb-mgm-type-1",
					CPUs:   2,
					Memory: 30,
				},
			},
			DataNode: SupportedInstanceTypeList{
				{
					Id:     "rondb-data-type-1",
					CPUs:   16,
					Memory: 100,
				},
				{
					Id:     "rondb-data-type-2",
					CPUs:   32,
					Memory: 200,
				},
			},
			MySQLNode: SupportedInstanceTypeList{
				{
					Id:     "rondb-mysql-type-1",
					CPUs:   8,
					Memory: 100,
				},
				{
					Id:     "rondb-mysql-type-2",
					CPUs:   16,
					Memory: 100,
				},
			},
			APINode: SupportedInstanceTypeList{
				{
					Id:     "rondb-api-type-1",
					CPUs:   8,
					Memory: 100,
				},
				{
					Id:     "rondb-api-type-2",
					CPUs:   16,
					Memory: 100,
				},
			},
		},
	}

	cases := []struct {
		input    NodeType
		expected SupportedInstanceTypeList
	}{
		{
			input:    HeadNode,
			expected: types.Head,
		},
		{
			input:    WorkerNode,
			expected: types.Worker,
		},
		{
			input:    RonDBManagementNode,
			expected: types.RonDB.ManagementNode,
		},
		{
			input:    RonDBDataNode,
			expected: types.RonDB.DataNode,
		},
		{
			input:    RonDBMySQLNode,
			expected: types.RonDB.MySQLNode,
		},
		{
			input:    RonDBAPINode,
			expected: types.RonDB.APINode,
		},
		{
			input:    NodeType(""),
			expected: nil,
		},
	}

	for _, c := range cases {
		output := types.GetByNodeType(c.input)
		if !reflect.DeepEqual(c.expected, output) {
			t.Fatalf("error while matching [%s] :\nexpected %#v \nbut got %#v", c.input.String(), c.expected, output)
		}
	}
}

func TestSortSupportedNodeTypes(t *testing.T) {
	input := SupportedInstanceTypeList{
		{
			Id:     "node-type-2",
			CPUs:   32,
			Memory: 64,
		},
		{
			Id:     "node-type-4",
			CPUs:   4,
			Memory: 16,
		},
		{
			Id:     "node-type-5",
			CPUs:   2,
			Memory: 8,
		},
		{
			Id:     "node-type-6",
			CPUs:   16,
			Memory: 32,
		},
		{
			Id:     "node-type-8",
			CPUs:   8,
			Memory: 16,
		},
		{
			Id:     "node-type-11",
			CPUs:   2,
			Memory: 16,
		},
	}

	expected := SupportedInstanceTypeList{
		{
			Id:     "node-type-5",
			CPUs:   2,
			Memory: 8,
		},
		{
			Id:     "node-type-11",
			CPUs:   2,
			Memory: 16,
		},
		{
			Id:     "node-type-4",
			CPUs:   4,
			Memory: 16,
		},
		{
			Id:     "node-type-8",
			CPUs:   8,
			Memory: 16,
		},
		{
			Id:     "node-type-6",
			CPUs:   16,
			Memory: 32,
		},
		{
			Id:     "node-type-2",
			CPUs:   32,
			Memory: 64,
		},
	}

	input.Sort()
	if !reflect.DeepEqual(expected, input) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, input)
	}
}

func TestGetAllNodeTypes(t *testing.T) {
	expected := []string{
		HeadNode.String(),
		WorkerNode.String(),
		RonDBManagementNode.String(),
		RonDBDataNode.String(),
		RonDBMySQLNode.String(),
		RonDBAPINode.String(),
	}

	output := GetAllNodeTypes()
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestClusterStateString(t *testing.T) {
	states := []ClusterState{
		Starting,
		Pending,
		Initializing,
		Running,
		Stopping,
		Stopped,
		Error,
		TerminationWarning,
		ShuttingDown,
		Updating,
		Decommissioning,
		RonDBInitializing,
		StartingHopsworks,
		WorkerPending,
		WorkerInitializing,
		WorkerStarting,
		WorkerError,
		WorkerShuttingdown,
		WorkerDecommissioning,
		ClusterDeleted,
	}

	for _, v := range states {
		if string(v) != v.String() {
			t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", string(v), v.String())
		}
	}
}

func TestActivationStateString(t *testing.T) {
	states := []ActivationState{
		Startable,
		Stoppable,
		Terminable,
	}

	for _, v := range states {
		if string(v) != v.String() {
			t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", string(v), v.String())
		}
	}
}

func TestBackupStateString(t *testing.T) {
	states := []BackupState{
		PendingBackup,
		ProcessingBackup,
		InitializingBackup,
		DeletingBackup,
		BackupSucceed,
		BackupFailed,
		BackupDeleted,
	}

	for _, v := range states {
		if string(v) != v.String() {
			t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", string(v), v.String())
		}
	}
}
