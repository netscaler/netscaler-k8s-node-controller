package main

import (
	"fmt"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

var fakeK8sApi *KubernetesAPIServer
var NsNitroObj *NitroClient
var InputObj *ControllerInput

func getClientAndDeviceInfo() (*ControllerInput, *NitroClient, *KubernetesAPIServer) {
	if NsNitroObj != nil && fakeK8sApi != nil && InputObj != nil {
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


func TestInitFlannel(t *testing.T) {
	input, nsObj, api := getClientAndDeviceInfo()
	InitFlannel(api, nsObj, input)
}
func TestTerminateFlannel(t *testing.T) {
	input, nsObj, api := getClientAndDeviceInfo()
	TerminateFlannel(api, nsObj, input)
}
