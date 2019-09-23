package main

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
	//"fmt"
)

func TestFetchCitrixNodeControllerInput(t *testing.T) {
	IngressDeviceIP := os.Getenv("NS_IP")
	os.Setenv("NS_IP", "")
	IngressDeviceVtepMAC := os.Getenv("NS_VTEP_MAC")
	os.Setenv("NS_VTEP_MAC", "")
	IngressDeviceUsername := os.Getenv("NS_LOGIN")
	os.Setenv("NS_LOGIN", "")
	IngressDevicePassword := os.Getenv("NS_PASSWORD")
	os.Setenv("NS_PASSWORD", "")
	IngressDeviceVtepIP := os.Getenv("NS_SNIP")
	os.Setenv("NS_SNIP", "")
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("FetchCitrixNodeControllerInput should have panicked!")
			}
		}()
		// This function should cause a panic
		FetchCitrixNodeControllerInput()
	}()

	os.Setenv("NS_IP", IngressDeviceIP)
	os.Setenv("NS_VTEP_MAC", IngressDeviceVtepMAC)
	os.Setenv("NS_LOGIN", IngressDeviceUsername)
	os.Setenv("NS_PASSWORD", IngressDevicePassword)
	os.Setenv("NS_SNIP", IngressDeviceVtepIP)
	FetchCitrixNodeControllerInput()
}

func TestWaitForConfigMapInput(t *testing.T) {
	input, _, api := getClientAndDeviceInfo()
	api.Client.CoreV1().ConfigMaps("citrix").Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "citrix-node-controller"},
		Data:       map[string]string{"Operation": "ADD"},
	})
	WaitForConfigMapInput(api, input)
}

/*
func TestMonitorIngressDevice(t *testing.T){
	controllerInput := FetchCitrixNodeControllerInput()
        ingressDevice := createIngressDeviceClient(controllerInput)

	MonitorIngressDevice(ingressDevice, controllerInput)
}*/
func TestIsValidIP4(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(false, IsValidIP4("333.22.1.1"), "Invalid IP")
	assert.Equal(false, IsValidIP4("abc.22.1.1"), "Invalid IP")
	assert.Equal(false, IsValidIP4("22.abc.1.1"), "Invalid IP")
	assert.Equal(false, IsValidIP4("22.334.1.1"), "Invalid IP")
	assert.Equal(false, IsValidIP4("22.1.334.1"), "Invalid IP")
	assert.Equal(false, IsValidIP4("22.1.abc.1"), "Invalid IP")
	assert.Equal(false, IsValidIP4("22.1.1.abc"), "Invalid IP")
	assert.Equal(false, IsValidIP4("22.1.1.1111"), "Invalid IP")
	assert.Equal(true, IsValidIP4("22.1.1.1"), "Invalid IP")
}
