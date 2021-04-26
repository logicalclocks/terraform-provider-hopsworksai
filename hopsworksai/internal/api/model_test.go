package api

import (
	"fmt"
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
		} else {
			if err.Error() != resp.Message {
				t.Fatalf("validate response should throw an error with message: %s, but got %s", resp.Message, err)
			}
		}
	}

}
