package main

import (
	"fmt"
	"k8s.io/client-go/kubernetes/fake"
	"runtime"
	"testing"
	"github.com/stretchr/testify/assert"
)

var fakeK8sApi *KubernetesAPIServer = nil
var NsNitroObj *NitroClient = nil
var InputObj   *ControllerInput = nil
func getClientAndDeviceInfo() (*ControllerInput, *NitroClient, *KubernetesAPIServer) {
	if (NsNitroObj !=  nil && fakeK8sApi != nil && InputObj != nil){
		return InputObj, NsNitroObj, fakeK8sApi
	}
	fmt.Println("TEST Flannel: Setting up the K8s interface and Ns interface")
	InputObj := FetchCitrixNodeControllerInput()
	NsNitroObj := createIngressDeviceClient(InputObj)
	fake := fake.NewSimpleClientset()
	fakeK8sApi = &KubernetesAPIServer{
		 Suffix: "Test",
		 Client: fake,
	}
	return InputObj, NsNitroObj, fakeK8sApi
}

func TestGetDummyNode(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	fmt.Println("Current test filename: " + filename)
	input, _, api := getClientAndDeviceInfo()
	input.DummyNodeLabel = "citrixadc"
	node := api.GetDummyNode(input)
	if node == nil {
		node1 := api.CreateDummyNode(input)
		if node1 == nil {
			t.Error("Failed to create Dummy Node")
		}
	} else {
		node = api.CreateDummyNode(input)
		if node == nil {
			t.Error("Expected Nil since there is already a nOde with same Label ")
		}
	}
}

func TestInitializeNode(t *testing.T) {
	input, _, _ := getClientAndDeviceInfo()
	node := InitializeNode(input)
	if node == nil {
		t.Error("Expected Node but its NULL ")
	}
}
func TestCreateDummyNode(t *testing.T) {
	input, _, api := getClientAndDeviceInfo()
	node := api.CreateDummyNode(input)
	if (node != nil){
		node := api.CreateDummyNode(input)
		if (node != nil) {
			t.Error("Expected node creation failed")
		}
	}
}
func TestInitFlannel(t *testing.T) {
	input, nsObj, api := getClientAndDeviceInfo()
	InitFlannel(api, nsObj, input)
}
func TestTerminateFlannel(t *testing.T){
	input, nsObj, api := getClientAndDeviceInfo()
        TerminateFlannel(api, nsObj, input) 
}
func TestDeleteDummyNode(t *testing.T){
	assert := assert.New(t)
	input, _, api := getClientAndDeviceInfo()
	api.CreateDummyNode(input)
	node := api.GetDummyNode(input)
	assert.Equal(true, api.DeleteDummyNode(node), "Removing the existing Node")
	assert.Equal(false, api.DeleteDummyNode(node), "Removing the non existing Node")
}
