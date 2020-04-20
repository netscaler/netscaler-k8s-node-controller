package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"os"
	"fmt"
	"strconv"
	"strings"
)

// Flags which keep track of CNC states
var (
	NetscalerInit      = 0x00000008
	NetscalerTerminate = 0x00000010
        MaxVNID            = 0xffffff  // 16777215
        MinVNID            = 1
        MaxPort            = 0xffff   // 65535
        MinPort            = 1
)

// Node structure keeps the parsed information of a Node Object
type Node struct {
	HostName       string `json:"hostname,omitempty"`
	IPAddr         string `json:"address,omitempty"`
	ExternalIPAddr string `json:"externalip,omitempty"`
	PodCIDR        string `json:"podcidr,omitempty"`
	PodVTEP        string `json:"podvtep,omitempty"`
	PodNetMask     string `json:"podnetmask,omitempty"`
	PodNetwork     string `json:"podnetwork,omitempty"`
	PodAddress     string `json:"podaddress,omitempty"`
	NextPodAddress string `json:"nextpodaddress,omitempty"`
	PodMaskLen     string `json:"prefix,omitempty"`
	Type           string `json:"type,omitempty"`
	VxlanPort      string `json:"vxlanport,omitempty"`
	Count          int    `json:"count,omitempty"`
	Label          string `json:"label,omitempty"`
	Role           string `json:"role,omitempty"`
}

var NodeList[] *Node 
var PrefixMaskTable = make(map[string]string)
func InitPrefixMaskTable(){
        PrefixMaskTable["8"] = "255.0.0.0"
        PrefixMaskTable["9"] = "255.128.0.0"
        PrefixMaskTable["10"] = "255.192.0.0"
        PrefixMaskTable["11"] = "255.224.0.0"
        PrefixMaskTable["12"] = "255.240.0.0"
        PrefixMaskTable["13"] = "255.248.0.0"
        PrefixMaskTable["14"] = "255.252.0.0"
        PrefixMaskTable["15"] = "255.254.0.0"
        PrefixMaskTable["16"] = "255.255.0.0"
        PrefixMaskTable["17"] = "255.255.128.0"
        PrefixMaskTable["18"] = "255.255.192.0"
        PrefixMaskTable["19"] = "255.255.224.0"
        PrefixMaskTable["20"] = "255.255.240.0"
        PrefixMaskTable["21"] = "255.255.248.0"
        PrefixMaskTable["22"] = "255.255.252.0"
        PrefixMaskTable["23"] = "255.255.254.0"
        PrefixMaskTable["24"] = "255.255.255.0"
        PrefixMaskTable["25"] = "255.255.255.128"
        PrefixMaskTable["26"] = "255.255.255.192"
        PrefixMaskTable["27"] = "255.255.255.224"
        PrefixMaskTable["28"] = "255.255.255.240"
        PrefixMaskTable["29"] = "255.255.255.248"
        PrefixMaskTable["30"] = "255.255.255.252"
        PrefixMaskTable["31"] = "255.255.255.254"
        PrefixMaskTable["32"] = "255.255.255.255"
}

// ControllerInput is the inputs passed to Citrix Node Controller
type ControllerInput struct {
	State                  int
	IngressDeviceIP        string
	IngressDeviceVtepMAC   string
	IngressDeviceNetprof   string
	IngressDeviceUsername  string
	IngressDevicePassword  string
	IngressDeviceVtepIP    string
	IngressDevicePodCIDR   string
	IngressDevicePodIP     string
	IngressDeviceVxlanID   int
	IngressDeviceVxlanIDs  string
	IngressDeviceVRID      int
	IngressDeviceVRIDs     string
	NodeSubnetMask         string
	NodeCIDR               string
	ClusterCNI             string
	CncOperation           string
	ClusterCNIPort         int
	VxlanPort	       int
	DummyNodeLabel         string
	Namespace         string
	NodesInfo              map[string]*Node
}

// This function validate given input IP and return error if its not IPv4 standard
func IsValidIP4(ipAddress string) bool {
	ipaddress := strings.Split(ipAddress, ".")
	firstOctect, err := strconv.Atoi(ipaddress[0])
	if err != nil {
		return false
	}
	if firstOctect < 0 || firstOctect > 255 {
		return false
	}
	secondOctect, err := strconv.Atoi(ipaddress[1])
	if err != nil {
		return false
	}
	if secondOctect < 0 || secondOctect > 255 {
		return false
	}
	thirdOctect, err := strconv.Atoi(ipaddress[2])
	if err != nil {
		return false
	}
	if thirdOctect < 0 || thirdOctect > 255 {
		return false
	}
	fourthOctect, err := strconv.Atoi(ipaddress[3])

	if err != nil {
		return false
	}
	if fourthOctect < 0 || fourthOctect > 255 {
		return false
	}
	return true
}

// This function validate given vxlan Port is valid.
func IsValidVxlanPort(port int) bool {
     if (port >= MinPort && port <= MaxPort) {
         return true    
     }
     return false
}

// This function validate given vxlan ID is a valid as per RFC.
func IsValidVxlanID(vni int) bool {
     if (vni >= MinVNID && vni <= MaxVNID) {
         return true    
     }
     return false
}

// FetchCitrixNodeControllerInput parses whole input provided by the user and store into controller input
func FetchCitrixNodeControllerInput() *ControllerInput {
	InputDataBuff := ControllerInput{}
	InputDataBuff.IngressDeviceIP = os.Getenv("NS_IP")
	configError := 0
	if len(InputDataBuff.IngressDeviceIP) == 0 {
		klog.Error("[ERROR] Ingress Device IP (NS_IP) is required, SNIP with Management access enabled")
		configError = 1
	}
	if !(IsValidIP4(InputDataBuff.IngressDeviceIP)) {
		klog.Error("[ERROR] Invalid IPV4 ")
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
	InputDataBuff.IngressDevicePodCIDR = os.Getenv("NETWORK")
	if len(InputDataBuff.IngressDevicePodCIDR) == 0 {
		klog.Infof("[ERROR] Provide Ingress device pod subnet CIDR (NETWORK)")
		configError = 1
	}
	InputDataBuff.IngressDeviceVtepIP = os.Getenv("REMOTE_VTEPIP")
	if len(InputDataBuff.IngressDeviceVtepIP) == 0 {
		klog.Info("[INFO] Ingress Device VTEP IP (REMOTE_VTEPIP)  is empty, Hence taking NS_SNIP as VTEP IP = ", InputDataBuff.IngressDeviceIP)
		InputDataBuff.IngressDeviceVtepIP = InputDataBuff.IngressDeviceIP
		configError = 1
	}
	if !(IsValidIP4(InputDataBuff.IngressDeviceVtepIP)) {
		klog.Error("[ERROR] Invalid IP ")
		configError = 1
	}
	VxlanPort := os.Getenv("VXLAN_PORT")
        if len(VxlanPort) == 0 {
                fmt.Println("[ERROR] VxlanPort (VXLAN_PORT) is must for extending the route")
                configError = 1
        }
	InputDataBuff.VxlanPort, _ = strconv.Atoi(VxlanPort)
    
        if !(IsValidVxlanPort(InputDataBuff.VxlanPort)) {
		klog.Error("[ERROR] VXLAN Port is  not in the range, Minimum value: 1, Maximum value: 65535")
		configError = 1
        }

	InputDataBuff.IngressDeviceVxlanIDs = os.Getenv("VNID")
	InputDataBuff.IngressDeviceVxlanID, _ = strconv.Atoi(InputDataBuff.IngressDeviceVxlanIDs)
	if InputDataBuff.IngressDeviceVxlanID == 0 {
		klog.Info("[ERROR] vxlan id (VNID) has Not Given")
                configError = 1
	}

        if !(IsValidVxlanID(InputDataBuff.IngressDeviceVxlanID)) {
		klog.Error("[ERROR] VNI not in the range, Minimum value: 1, Maximum value: 16777215")
		configError = 1
        }
     	
	if configError == 1 {
		klog.Error("Unable to get the above mentioned input from YAML")
		panic("[ERROR] Killing Container.........Please restart Citrix Node Controller with Valid Inputs")
	}
	splitString := strings.Split(InputDataBuff.IngressDevicePodCIDR, "/")
	InputDataBuff.NodeSubnetMask = PrefixMaskTable[splitString[1]]
	return &InputDataBuff
}

// This waits for the config map input. This function keep CNC for getting COnfig map input.
func WaitForConfigMapInput(api *KubernetesAPIServer, ControllerInputObj *ControllerInput) {
	klog.Info("[INFO] Waiting for the Config Map input...")
	for {
		configmap, err := api.Client.CoreV1().ConfigMaps("citrix").Get("citrix-node-controller", metav1.GetOptions{})
		if err == nil {
			ConfigMapData := make(map[string]string)
			ConfigMapData = configmap.Data
			klog.Info("[INFO] Config Map Data", ConfigMapData)
			ControllerInputObj.CncOperation = ConfigMapData["operation"]
			break
		}
	}
}

