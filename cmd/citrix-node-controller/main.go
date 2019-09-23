package main

import (
	"k8s.io/klog"
)

func InitCitrixNodeController() error {
	klog.InitFlags(nil)
	klog.Info("Initializing CNC")
	InitPrefixMaskTable()
	return nil
}

func StartCitrixNodeController() {
	controllerInput := FetchCitrixNodeControllerInput()
	ingressDevice := createIngressDeviceClient(controllerInput)
	api, err := CreateK8sApiserverClient()
	if err != nil {
		klog.Fatal("[ERROR] K8s Client Error", err)
	}
	ConfigMapInputWatcher(api, ingressDevice, controllerInput)
}

func main() {
	InitCitrixNodeController()
	StartCitrixNodeController()
}
