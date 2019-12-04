package main

import (
	"fmt"
	"k8s.io/klog"
	"runtime"
	"testing"
	//"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertPrefixLenToMask(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	fmt.Println("Current test filename: " + filename)
	t.Log("Prefix Length 24")
	mask := ConvertPrefixLenToMask("24")
	if mask != "255.255.255.0" {
		t.Error("Expected 255.255.255.0, got ", mask)
	}
	mask = ConvertPrefixLenToMask("8")
	if mask != "255.0.0.0" {
		t.Error("Expected 255.0.0.0, got ", mask)
	}
	mask = ConvertPrefixLenToMask("16")
	if mask != "255.255.0.0" {
		t.Error("Expected 255.255.0.0, got ", mask)
	}
	mask = ConvertPrefixLenToMask("30")
	if mask != "255.255.255.252" {
		t.Error("Expected 255.255.255.252, got ", mask)
	}
	mask = ConvertPrefixLenToMask("29")
	if mask != "255.255.255.248" {
		t.Error("Expected 255.255.255.248, got ", mask)
	}
	mask = ConvertPrefixLenToMask("25")
	if mask != "255.255.255.128" {
		t.Error("Expected 255.255.255.128, got ", mask)
	}
	mask = ConvertPrefixLenToMask("19")
	if mask != "255.255.224.0" {
		t.Error("Expected 255.255.224.0, got ", mask)
	}
	mask = ConvertPrefixLenToMask("17")
	if mask != "255.255.128.0" {
		t.Error("Expected 255.255.128.0, got ", mask)
	}
	mask = ConvertPrefixLenToMask("30")
	if mask != "255.255.255.252" {
		t.Error("Expected 255.255.255.252, got ", mask)
	}
}

func TestConfigDecider(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	fmt.Println("Current test filename: " + filename)
	input, nsobj, api := getClientAndDeviceInfo()
	if err != nil {
		klog.Fatal("K8s Client Error", err, nsobj)
	}
	ConfigDecider(api, nsobj, input)
}
func TestHandleConfigMapAddEvent(t *testing.T) {
	input, obj, api := getClientAndDeviceInfo()
	HandleConfigMapAddEvent(api, obj, obj, input)

}
func TestHandleConfigMapDeleteEvent(t *testing.T) {
	input, obj, api := getClientAndDeviceInfo()
	api.Client.CoreV1().ConfigMaps("citrix").Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "citrix-node-controller"},
		Data:       map[string]string{"Operation": "ADD"},
	})
	configobj, _ := api.Client.CoreV1().ConfigMaps("citrix").Get("citrix-node-controller", metav1.GetOptions{})
	input.State = NetscalerInit
	HandleConfigMapDeleteEvent(api, configobj, obj, input)
}
func TestHandleConfigMapUpdateEvent(t *testing.T) {
	input, obj, api := getClientAndDeviceInfo()
	api.Client.CoreV1().ConfigMaps("citrix").Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "citrix-node-controller"},
		Data:       map[string]string{"Operation": "ADD"},
	})
	configobj, _ := api.Client.CoreV1().ConfigMaps("citrix").Get("citrix-node-controller", metav1.GetOptions{})
	HandleConfigMapUpdateEvent(api, configobj, configobj, obj, input)
}
