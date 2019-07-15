package main

import (
	"k8s.io/klog"
	"os"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)
var (
	NetscalerInit=0x00000008
	NetscalerTerminate=0x00000010
)


type Node struct {
	HostName   string `json:"hostname,omitempty"`
	IPAddr     string `json:"address,omitempty"`
	ExternalIPAddr     string `json:"address,omitempty"`
	PodCIDR    string `json:"podcidr,omitempty"`
	PodVTEP    string `json:"podvtep,omitempty"`
	PodNetMask string `json:"podcidr,omitempty"`
	PodAddress string `json:"podaddress,omitempty"`
	NextPodAddress string `json:"nextpodaddress,omitempty"`
	PodMaskLen string `json:"prefix, omitempty"`
	Type       string `json:"podvtep,omitempty"`
	VxlanPort  string `json:"prefix, omitempty"`
	Count      int    `json:"count,omitempty"`
	Label     string `json:"label,omitempty"`
	Role     string `json:"label,omitempty"`
}

type ControllerInput struct {
	State		      int 
	IngressDeviceIP       string
	IngressDeviceVtepMAC  string
	IngressDeviceNetprof  string
	IngressDeviceUsername string
	IngressDevicePassword string
	IngressDeviceVtepIP  string
	IngressDevicePodCIDR  string
	IngressDevicePodIP  string
	IngressDevicePodSubnet  string
	IngressDeviceVxlanID  int
	IngressDeviceVxlanIDs  string
	IngressDeviceVRID  int
	IngressDeviceVRIDs  string
	NodeSubnetMask	      string
	NodeCIDR	      string
	ClusterCNI            string
	CncOperation 	      string
	ClusterCNIPort        int
	DummyNodeLabel        string
	NodesInfo             map[string]*Node
}


func IsValidIP4(ipAddress string) bool {
        ipaddress := strings.Split(ipAddress, ".")
	firstOctect, err := strconv.Atoi(ipaddress[0])
	if (err != nil) {
		return false
	}
        if (firstOctect <0 || firstOctect >255) {
		return false
	}
	secondOctect, err := strconv.Atoi(ipaddress[1])
	if (err != nil) {
		return false
	}
        if (secondOctect<0 || secondOctect>255) {
		return false
	}
	thirdOctect, err := strconv.Atoi(ipaddress[2]); 
	if (err != nil) {
		return false
	}
        if (thirdOctect<0 || thirdOctect>255) {
		return false
	}
	fourthOctect, err := strconv.Atoi(ipaddress[3]); 

	if (err != nil) {
		return false
	}
        if (fourthOctect<0 || fourthOctect>255) {
		return false
	}
	return true
}
/*
func FetchOptionalInputs(InputDataBuff *ControllerInput) *ControllerInput {
	InputDataBuff.IngressDeviceVtepIP = os.Getenv("NS_VTEP_IP")
	if len(InputDataBuff.IngressDeviceVtepIP) == 0 {
		klog.Info("[INFO] Ingress Device VTEP IP (NS_VTEP_IP)  is empty, Hence taking NS_SNIP as VTEP IP = ", InputDataBuff.IngressDeviceIP)
		InputDataBuff.IngressDeviceVtepIP = InputDataBuff.IngressDeviceIP 
	}
	if (!(IsValidIP4(InputDataBuff.IngressDeviceVtepIP))) {
		klog.Error("[ERROR] Invalid IP ")
		panic("[ERROR] Killing Container.........Please restart Citrix Node Controller with Valid Inputs")
	}
	splitString := strings.Split(InputDataBuff.IngressDevicePodCIDR, "/")
        subnet := strings.Split(splitString[0], ".") 
        InputDataBuff.IngressDevicePodIP = subnet[0] + "." + subnet[1] + "." +subnet[2]+".1"
        InputDataBuff.IngressDevicePodSubnet = subnet[0] + "." + subnet[1] + "." +subnet[2]+".0/"+splitString[1]
	InputDataBuff.DummyNodeLabel = "citrixadc"
        InputDataBuff.IngressDeviceVxlanIDs = os.Getenv("NS_VXLAN_ID")
	InputDataBuff.IngressDeviceVxlanID, _ = strconv.Atoi(InputDataBuff.IngressDeviceVxlanIDs)
	if InputDataBuff.IngressDeviceVxlanID == 0 {
		klog.Info("[INFO] VXLAN ID has Not Given, taking 1 as default VXLAN_ID (flannel uses 1 as default)")
		InputDataBuff.IngressDeviceVxlanID = 1
		InputDataBuff.IngressDeviceVxlanIDs = "1"
	}
        InputDataBuff.IngressDeviceVRIDs = os.Getenv("NS_VRID")
	InputDataBuff.IngressDeviceVRID, _ = strconv.Atoi(InputDataBuff.IngressDeviceVRIDs)
	if InputDataBuff.IngressDeviceVRID == 0 {
		klog.Info("[INFO] VRID has Not Given, taking 99 as default VRID")
		InputDataBuff.IngressDeviceVRID = 99
		InputDataBuff.IngressDeviceVRIDs = "99"
	}
	InputDataBuff.ClusterCNIPort, _ = strconv.Atoi(os.Getenv("K8S_VXLAN_PORT"))
	if InputDataBuff.ClusterCNIPort == 0 {
		klog.Info("[INFO] K8S_VXLAN_PORT has Not Given, taking default 8472 as Vxlan Port")
		InputDataBuff.ClusterCNIPort = 8472
	}
	InputDataBuff.ClusterCNI = os.Getenv("K8S_CNI")
	if len(InputDataBuff.ClusterCNI) == 0 {
		klog.Infof("[INFO] Cluster CNI information is Empty")
	}
	InputDataBuff.NodesInfo = make(map[string]*Node)
	return InputDataBuff
}

func FetchMandatoryInputs() *ControllerInput {
	InputDataBuff := ControllerInput{}
	InputDataBuff.IngressDeviceIP = os.Getenv("NS_IP")
	configError := 0
	if len(InputDataBuff.IngressDeviceIP) == 0 {
		klog.Error("[ERROR] Ingress Device IP (NS_IP) is required, SNIP with Management access enabled")
		configError = 1
	}
	if (!(IsValidIP4(InputDataBuff.IngressDeviceIP))) {
		klog.Error("[ERROR] Invalid IP ")
		configError = 1
	}
	InputDataBuff.IngressDeviceUsername = os.Getenv("NS_USER")
	if len(InputDataBuff.IngressDeviceUsername) == 0 {
		klog.Error("[ERROR] Ingress Device user name (NS_USER) is  required")
		configError = 1
	}
	InputDataBuff.IngressDevicePassword = os.Getenv("NS_PASSWORD")
	if len(InputDataBuff.IngressDevicePassword) == 0 {
		klog.Error("[ERROR] Ingress Device password (NS_PASSWORD) is  required")
		configError = 1
	}
	InputDataBuff.IngressDevicePodCIDR = os.Getenv("NS_POD_CIDR")
	if len(InputDataBuff.IngressDevicePodCIDR) == 0 {
		klog.Infof("[ERROR] Provide Ingress device pod subnet CIDR (NS_POD_CIDR)")
		configError = 1
	}
	InputDataBuff.NodeCIDR = os.Getenv("NODE_CNI_CIDR")
	if len(InputDataBuff.NodeCIDR) == 0 {
		klog.Infof("[ERROR] Provide Node subnet CIDR (NODE_CNI_CIDR: 10.241.0.0/16)")
		configError = 1
	}
	nodecidr := strings.Split(InputDataBuff.NodeCIDR, "/")
	//InputDataBuff.NodeSubnet
	InputDataBuff.NodeSubnetMask = ConvertPrefixLenToMask(nodecidr[1])
	if configError == 1 {
		klog.Error("Unable to get the above mentioned input from YAML")
		panic("[ERROR] Killing Container.........Please restart Citrix Node Controller with Valid Inputs")
	}
	return &InputDataBuff
}
*/
func FetchCitrixNodeControllerInput() *ControllerInput {
	InputDataBuff := ControllerInput{}
	InputDataBuff.IngressDeviceIP = os.Getenv("NS_IP")
	configError := 0
	if len(InputDataBuff.IngressDeviceIP) == 0 {
		klog.Error("[ERROR] Ingress Device IP (NS_IP) is required, SNIP with Management access enabled")
		configError = 1
	}
	if (!(IsValidIP4(InputDataBuff.IngressDeviceIP))) {
		klog.Error("[ERROR] Invalid IP ")
		configError = 1
	}
	InputDataBuff.IngressDeviceUsername = os.Getenv("NS_USER")
	if len(InputDataBuff.IngressDeviceUsername) == 0 {
		klog.Error("[ERROR] Ingress Device user name (NS_USER) is  required")
		configError = 1
	}
	InputDataBuff.IngressDevicePassword = os.Getenv("NS_PASSWORD")
	if len(InputDataBuff.IngressDevicePassword) == 0 {
		klog.Error("[ERROR] Ingress Device password (NS_PASSWORD) is  required")
		configError = 1
	}
	InputDataBuff.IngressDeviceVtepMAC = os.Getenv("NS_VTEP_MAC")
	if len(InputDataBuff.IngressDeviceVtepMAC) == 0 {
		klog.Error("[ERROR] Ingress Device VtepMAC (NS_VTEP_MAC) is  required. Please configure VMAC on the Interface towards kubernetes cluster and provide that VMAC as NS_VTEP_MAC (https://docs.citrix.com/en-us/netscaler/12/system/high-availability-introduction/configuring-virtual-mac-addresses-high-availability.html).")
		configError = 1
	}
	InputDataBuff.IngressDeviceNetprof = os.Getenv("NS_NETPROFILE")
	if len(InputDataBuff.IngressDeviceNetprof) == 0 {
		klog.Error("[ERROR] Provide Netprofile name used in Citrix Ingress Controller")
		configError = 1
	}
	InputDataBuff.IngressDevicePodCIDR = os.Getenv("NS_POD_CIDR")
	if len(InputDataBuff.IngressDevicePodCIDR) == 0 {
		klog.Infof("[ERROR] Provide Ingress device pod subnet CIDR (NS_POD_CIDR)")
		configError = 1
	}
	InputDataBuff.NodeCIDR = os.Getenv("NODE_CNI_CIDR")
	if len(InputDataBuff.NodeCIDR) == 0 {
		klog.Infof("[ERROR] Provide Node subnet CIDR (NODE_CNI_CIDR: 10.241.0.0/16)")
		configError = 1
	}
	nodecidr := strings.Split(InputDataBuff.NodeCIDR, "/")
	//InputDataBuff.NodeSubnet
	InputDataBuff.NodeSubnetMask = ConvertPrefixLenToMask(nodecidr[1])
	if configError == 1 {
		klog.Error("Unable to get the above mentioned input from YAML")
		panic("[ERROR] Killing Container.........Please restart Citrix Node Controller with Valid Inputs")
	}
	InputDataBuff.IngressDeviceVtepIP = os.Getenv("NS_VTEP_IP")
	if len(InputDataBuff.IngressDeviceVtepIP) == 0 {
		klog.Info("[INFO] Ingress Device VTEP IP (NS_VTEP_IP)  is empty, Hence taking NS_SNIP as VTEP IP = ", InputDataBuff.IngressDeviceIP)
		InputDataBuff.IngressDeviceVtepIP = InputDataBuff.IngressDeviceIP 
	}
	if (!(IsValidIP4(InputDataBuff.IngressDeviceVtepIP))) {
		klog.Error("[ERROR] Invalid IP ")
		panic("[ERROR] Killing Container.........Please restart Citrix Node Controller with Valid Inputs")
	}
	splitString := strings.Split(InputDataBuff.IngressDevicePodCIDR, "/")
        subnet := strings.Split(splitString[0], ".") 
        InputDataBuff.IngressDevicePodIP = subnet[0] + "." + subnet[1] + "." +subnet[2]+".1"
        InputDataBuff.IngressDevicePodSubnet = subnet[0] + "." + subnet[1] + "." +subnet[2]+".0/"+splitString[1]
	InputDataBuff.DummyNodeLabel = "citrixadc"
        InputDataBuff.IngressDeviceVxlanIDs = os.Getenv("NS_VXLAN_ID")
	InputDataBuff.IngressDeviceVxlanID, _ = strconv.Atoi(InputDataBuff.IngressDeviceVxlanIDs)
	if InputDataBuff.IngressDeviceVxlanID == 0 {
		klog.Info("[INFO] VXLAN ID has Not Given, taking 1 as default VXLAN_ID (flannel uses 1 as default)")
		InputDataBuff.IngressDeviceVxlanID = 1
		InputDataBuff.IngressDeviceVxlanIDs = "1"
	}
        InputDataBuff.IngressDeviceVRIDs = os.Getenv("NS_VRID")
	InputDataBuff.IngressDeviceVRID, _ = strconv.Atoi(InputDataBuff.IngressDeviceVRIDs)
	if InputDataBuff.IngressDeviceVRID == 0 {
		klog.Info("[INFO] VRID has Not Given, taking 99 as default VRID")
		InputDataBuff.IngressDeviceVRID = 99
		InputDataBuff.IngressDeviceVRIDs = "99"
	}
	InputDataBuff.ClusterCNIPort, _ = strconv.Atoi(os.Getenv("K8S_VXLAN_PORT"))
	if InputDataBuff.ClusterCNIPort == 0 {
		klog.Info("[INFO] K8S_VXLAN_PORT has Not Given, taking default 8472 as Vxlan Port")
		InputDataBuff.ClusterCNIPort = 8472
	}
	InputDataBuff.ClusterCNI = os.Getenv("K8S_CNI")
	if len(InputDataBuff.ClusterCNI) == 0 {
		klog.Infof("[INFO] Cluster CNI information is Empty")
	}
	InputDataBuff.NodesInfo = make(map[string]*Node)
	return &InputDataBuff
}
/*
*************************************************************************************************
*   APIName :  WaitForConfigMapInput                                                            *
*   Input   :  Takes API server session called client and Controller input.             	*
*   Output  :  Wait till COnfig map applied and extract Operation field to proceed further.	*
*   Descr   :  This API waits till Citrix Node Config map input is supplied.                    *
*************************************************************************************************
 */
func WaitForConfigMapInput(api *KubernetesAPIServer, ControllerInputObj *ControllerInput){
	klog.Info("[INFO] Waiting for the Config Map input...")
	for{	 
		configmap, err := api.Client.CoreV1().ConfigMaps("citrix").Get("citrix-node-controller", metav1.GetOptions{})
		if (err == nil) {
			ConfigMapData := make(map[string]string)
			ConfigMapData = configmap.Data
			klog.Info("[INFO] Config Map Data", ConfigMapData)
			ControllerInputObj.CncOperation = ConfigMapData["operation"]
			break;
		}
	}
}
/*
*************************************************************************************************
*   APIName :  MonitorIngressDevice                                                             *
*   Input   :  Takes API server session called client.             			        *
*   Output  :  Invokes call back functions.	                                                *
*   Descr   :  This API is for watching the Nodes. API Monitors Kubernetes API server for Nodes *
*            events and store in node cache. Based on the events type, call back functions      *
*	     Will execute and perform the desired tasks.					*
*************************************************************************************************
 */
func MonitorIngressDevice(IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput) {
	vtepMac := getClusterInterfaceMac(IngressDeviceClient)
	if (vtepMac != "error" && vtepMac != "00:00:00:00:00:00"){
		ControllerInputObj.IngressDeviceVtepMAC = vtepMac
	} else {
		ControllerInputObj.IngressDeviceVtepMAC = os.Getenv("NS_VTEP_MAC")
		if len(ControllerInputObj.IngressDeviceVtepMAC) == 0 {
			klog.Error("[ERROR] Ingress Device VtepMAC (NS_VTEP_MAC) is  required")
			panic("[ERROR] Killing Container.........Please restart Citrix Node Controller with NS_VTEP_MAC as CNC could not get it from NS")
		}
	}
}
