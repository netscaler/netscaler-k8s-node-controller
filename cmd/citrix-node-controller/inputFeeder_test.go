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
	IngressDeviceUsername := os.Getenv("NS_LOGIN")
	os.Setenv("NS_USER", "")
	IngressDevicePassword := os.Getenv("NS_PASSWORD")
	os.Setenv("NS_PASSWORD", "")
	IngressDeviceVtepIP := os.Getenv("REMOTE_VTEPIP")
	os.Setenv("REMOTE_VTEPIP", "")
	Vnid := os.Getenv("VNID")
	os.Setenv("VNID", "")
	Network := os.Getenv("NETWORK")
	os.Setenv("NETWORK", "")
	VxlanPort := os.Getenv("VXLAN_PORT")
	os.Setenv("VXLAN_PORT", "")
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("FetchCitrixNodeControllerInput should have panicked!")
			}
		}()
		// This function should cause a panic
		FetchCitrixNodeControllerInput()
	}()

	if IngressDeviceIP == "" {
		IngressDeviceIP = "127.0.0.1"
	}
	os.Setenv("NS_IP", IngressDeviceIP)
	
	if IngressDeviceUsername == "" {
		IngressDeviceUsername = "test"
	}
	os.Setenv("NS_USER", IngressDeviceUsername)
	if IngressDevicePassword == "" {
		IngressDevicePassword = "test"
	}
	os.Setenv("NS_PASSWORD", IngressDevicePassword)
	if IngressDeviceVtepIP == "" {
		IngressDeviceVtepIP = "127.0.0.1"
	}
	os.Setenv("REMOTE_VTEPIP", IngressDeviceVtepIP)
	if Vnid == "" {
		Vnid = "999"
	}
	os.Setenv("VNID", Vnid)
	if Network == "" {
		Network = "192.128.1.0/24"
	}
	os.Setenv("NETWORK", Network)
	if VxlanPort == "" {
		VxlanPort = "8999"
	}
	os.Setenv("VXLAN_PORT", VxlanPort)
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
	assert.Equal(true, IsValidIP4("22.1.1.1"), "Valid IP")
}

func TestIsValidVxlanPort(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(true, IsValidVxlanPort(1), "Valid Port")
	assert.Equal(true, IsValidVxlanPort(65535), "Valid Port")
	assert.Equal(true, IsValidVxlanPort(100), "Valid Port")
	assert.Equal(false, IsValidVxlanPort(65537), "Invalid Port")
	assert.Equal(false, IsValidVxlanPort(0), "Invalid Port")
	assert.Equal(false, IsValidVxlanPort(-1), "Invalid Port")
}

func TestIsValidVxlanID(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(true, IsValidVxlanID(1), "Valid VNID")
	assert.Equal(true, IsValidVxlanID(16777215), "Valid VNID")
	assert.Equal(true, IsValidVxlanID(100), "Valid VNID")
	assert.Equal(false, IsValidVxlanID(16777216), "Invalid VNID")
	assert.Equal(false, IsValidVxlanID(0), "Invalid VNID")
	assert.Equal(false, IsValidVxlanID(-1), "Invalid VNID")
}
