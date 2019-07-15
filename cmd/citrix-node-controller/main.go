package main

import (
	"k8s.io/klog"
)

func InitCitrixNodeController() error {
	klog.InitFlags(nil)
	klog.Info("Initializing CNC")
	return nil
}
func StartCitrixNodeController() {
	controllerInput := FetchCitrixNodeControllerInput()
	ingressDevice := createIngressDeviceClient(controllerInput)
	if (len(controllerInput.IngressDeviceVtepMAC) == 0){
        	MonitorIngressDevice(ingressDevice, controllerInput)
	}

	api, err := CreateK8sApiserverClient()
	if err != nil {
		klog.Fatal("[ERROR] K8s Client Error", err)
	}
	WaitForConfigMapInput(api, controllerInput)
	ConfigDecider(api, ingressDevice, controllerInput)
	ConfigMapInputWatcher(api, ingressDevice, controllerInput)
	//CitrixNodeWatcher(api, ingressDevice, controllerInput)
	//StartRestServer()
}
func main() {
	InitCitrixNodeController()
	StartCitrixNodeController()
}
