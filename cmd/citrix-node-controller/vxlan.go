package main

import (
	"k8s.io/klog"
)

//func CreateVxlanConfig(ingressDevice *NitroClient, controllerInput *ControllerInput, node *Node) {
func CreateVxlanConfig(ingressDevice *NitroClient, controllerInput *ControllerInput) {

	configPack := ConfigPack{}
	vxlan := Vxlan{
		Id:   controllerInput.IngressDeviceVxlanID,
		Port: controllerInput.VxlanPort,
	}
	configPack.Set("vxlan", &vxlan)
	vxlanbind := Vxlan_srcip_binding{
		Id:    controllerInput.IngressDeviceVxlanID,
		Srcip: controllerInput.IngressDeviceVtepIP,
	}
	configPack.Set("vxlan_srcip_binding", &vxlanbind)

	nsip := Nsip{
		Ipaddress: controllerInput.IngressDevicePodIP,
		Netmask:   controllerInput.NodeSubnetMask,
	}
	configPack.Set("nsip", &nsip)
	AddIngressDeviceConfig(&configPack, ingressDevice)
}

func DeleteVxlanConfig(ingressDevice *NitroClient, controllerInput *ControllerInput) {

	configPack := ConfigPack{}
	vxlanargs := map[string]string{"id": controllerInput.IngressDeviceVxlanIDs}
	configPack.Set("vxlan", vxlanargs)

	nsipargs := map[string]string{"ipaddress": controllerInput.IngressDevicePodIP}
	configPack.Set("nsip", nsipargs)
	DeleteIngressDeviceConfig(&configPack, ingressDevice)
}

func InitFlannel(api *KubernetesAPIServer, ingressDevice *NitroClient, controllerInput *ControllerInput) {
	klog.Info("[INFO] Initializing Vxlan Config")
	ingressDevice.GetVxlanConfig(controllerInput)
	CreateVxlanConfig(ingressDevice, controllerInput)
	controllerInput.State |= NetscalerInit
}

func TerminateFlannel(api *KubernetesAPIServer, ingressDevice *NitroClient, controllerInput *ControllerInput) {
	klog.Info("[INFO] Terminating VXLAN Config")
	DeleteVxlanConfig(ingressDevice, controllerInput)
	controllerInput.State |= NetscalerTerminate
}
