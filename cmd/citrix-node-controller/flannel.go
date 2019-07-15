package main

import (
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
        "time"
)

/*
*************************************************************************************************
*   APIName :  InitializeNode                                                                   *
*   Input   :  Nil.					             			        *
*   Output  :  Nil.				                                                *
*   Descr   :  This API initialize a node and return it.					*
*************************************************************************************************
 */
func InitializeNode(obj *ControllerInput) *v1.Node {
	klog.Info("[INFO] Initializing a Dummy Node")
	backend_data := fmt.Sprintf("{VtepMAC:%s}", obj.IngressDeviceVtepMAC)
	NewNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "citrixadc",
		},
		Spec: v1.NodeSpec{
                       PodCIDR: obj.IngressDevicePodCIDR,
		       Unschedulable: true,
		       Taints: []v1.Taint{
				{Key: "key1", Value: "value1", Effect: "NoSchedule"},
				{Key: "key2", Value: "value2", Effect: "NoExecute"},
		       }, 
                },
                Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:               v1.NodeReady,
					Status:             v1.ConditionTrue,
					Reason:             "KubeletReady",
					Message:            "kubelet is posting ready status",
				},
			},
		},	
	}
	//NewNode.Sepc.PodCIDR = obj.IngressDevicePodCIDR
	NewNode.Labels = make(map[string]string)
	NewNode.Labels["com.citrix.nodetype"] = obj.DummyNodeLabel
	NewNode.Annotations = make(map[string]string)
	NewNode.Annotations["flannel.alpha.coreos.com/kube-subnet-manager"] = "true"
	NewNode.Annotations["flannel.alpha.coreos.com/backend-type"] = "vxlan"
	NewNode.Annotations["flannel.alpha.coreos.com/public-ip"] = obj.IngressDeviceVtepIP
	NewNode.Annotations["flannel.alpha.coreos.com/backend-data"] = backend_data
	NewNode.Annotations["flannel.alpha.coreos.com/public-ip-overwrite"] = obj.IngressDeviceVtepIP
	return NewNode
}

// DeleteDummyNode deletes the citrix adc node from kubernetes cluster.
// It takes Node as input and return true if able to delete node else false.
func (api KubernetesAPIServer) DeleteDummyNode(node *v1.Node) bool{
	klog.Info("[INFO] Deleting Citrix ADC Node")
	err := api.Client.CoreV1().Nodes().Delete(node.GetObjectMeta().GetName(), metav1.NewDeleteOptions(0))
	if err != nil {
		klog.Error("[ERROR] Node Deletion has failed", err)
		return false
	}
	klog.Info("[INFO] Deleted Citrix ADC Node ")
	return true
}

/*
*************************************************************************************************
*   APIName :  CreateDummyNode                                                                  *
*   Input   :  Takes API server session called client.             			        *
*   Output  :  Nil.				                                                *
*   Descr   :  This API  Creates a Dummy Node on K8s CLuster.					*
*************************************************************************************************
 */
func (api KubernetesAPIServer) CreateDummyNode(obj *ControllerInput) *v1.Node {
	klog.Info("[INFO] Creating Citrix ADC Node")
	NsAsDummyNode := InitializeNode(obj)
	node, err := api.Client.CoreV1().Nodes().Create(NsAsDummyNode)
	if err != nil {
		klog.Error("[ERROR] Node Creation has failed", err)
		return node
	}
        time.Sleep(10 * time.Second) //TODO, We have to wait till Node is available.
	klog.Info("[INFO] Created Citrix ADC Node of name=", node.GetObjectMeta().GetName())
	return node
}

/*
*************************************************************************************************
*   APIName :  GetDummyNode	                                                                *
*   Input   :  Takes API server session called client.             			        *
*   Output  :  Node Object if it present else retun Nil.				        *
*   Descr   :  This API  Get the Citrix Adc node if its present in the Cluster.			*
*************************************************************************************************
 */
func (api KubernetesAPIServer) GetDummyNode(obj *ControllerInput) *v1.Node {
	opts := metav1.GetOptions{}
	node, err := api.Client.CoreV1().Nodes().Get(obj.DummyNodeLabel, opts)
	if err != nil {
		return nil
	}
	klog.Info("[INFO] Get Node: Node Name:", node.GetObjectMeta().GetName())
	return node
}

/*
*************************************************************************************************
*   APIName :  CreateVxlanConfig	                                                        *
*   Input   :  Takes ingress Device session, controller input and node.		             	*
*   Output  :  Create Config Pack for VXLAM.				        		*
*   Descr   :  This API  calls if the CNI is flannel and it creates a VXLAN COnfig for that.	*
*************************************************************************************************
 */
//func CreateVxlanConfig(ingressDevice *NitroClient, controllerInput *ControllerInput, node *Node) {
func CreateVxlanConfig(ingressDevice *NitroClient, controllerInput *ControllerInput) {

	configPack := ConfigPack{}
	vxlan := Vxlan{
		Id:   controllerInput.IngressDeviceVxlanID,
		Port: controllerInput.ClusterCNIPort,
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
	BindToNetProfile(controllerInput, ingressDevice)
}
/*
*************************************************************************************************
*   APIName :  DeleteVxlanConfig	                                                        *
*   Input   :  Takes ingress Device session, controller input and node.		             	*
*   Output  :  Delete Config Pack for VXLAM.				        		*
*   Descr   :  This API  calls if the CNI is flannel and it clears a VXLAN COnfig for that.	*
*************************************************************************************************
 */
func DeleteVxlanConfig(ingressDevice *NitroClient, controllerInput *ControllerInput, node *Node) {
	
	UnBindNetProfile(controllerInput, ingressDevice)

	configPack := ConfigPack{}
	vxlanargs := map[string]string{"id": controllerInput.IngressDeviceVxlanIDs}
	configPack.Set("vxlan", vxlanargs)
   
	nsipargs := map[string]string{"ipaddress": controllerInput.IngressDevicePodIP}
	configPack.Set("nsip", nsipargs)
	DeleteIngressDeviceConfig(&configPack, ingressDevice)
}

/*
*************************************************************************************************
*   APIName :  InitFlannel	                                                                *
*   Input   :  Takes Api, ingress Device session and controller input.		           	*
*   Output  :  Retun Nil.								        *
*   Descr   :  This API  Initialize flannel Config by creating Dummy Node Vxlan Config.		*
*************************************************************************************************
 */
func InitFlannel(api *KubernetesAPIServer, ingressDevice *NitroClient, controllerInput *ControllerInput) {
	klog.Info("[INFO] Initializing Flannel Config")
	dummyNode := api.GetDummyNode(controllerInput)
	ingressDevice.GetVxlanConfig(controllerInput)
	if dummyNode == nil {
		api.CreateDummyNode(controllerInput)
		dummyNode = api.GetDummyNode(controllerInput)
	}
	//node := ParseNodeEvents(api, dummyNode, ingressDevice, controllerInput)
	//node.PodNetMask = "255.255.0.0" //Automate to find next highest number
	//CreateVxlanConfig(ingressDevice, controllerInput, node)
	CreateVxlanConfig(ingressDevice, controllerInput)
	controllerInput.State |= NetscalerInit 
}
/*
*************************************************************************************************
*   APIName :  TerminateFlannel	                                                                *
*   Input   :  Takes Api, ingress Device session and controller input.		           	*
*   Output  :  Retun Nil.								        *
*   Descr   :  This API  Initialize flannel Config by creating Dummy Node Vxlan Config.		*
*************************************************************************************************
 */
func TerminateFlannel(api *KubernetesAPIServer, ingressDevice *NitroClient, controllerInput *ControllerInput) {
	klog.Info("[INFO] Terminating Flannel Config")
	dummyNode := api.GetDummyNode(controllerInput)
	if dummyNode == nil {
		klog.Info("[ERROR] Expecting Dummy node to be Present Citrix ADC node \n")
		return 
	}
	node := ParseNodeEvents(api, dummyNode, ingressDevice, controllerInput)
	node.PodNetMask = controllerInput.NodeSubnetMask
	DeleteVxlanConfig(ingressDevice, controllerInput, node)
	api.DeleteDummyNode(dummyNode)
	controllerInput.State |= NetscalerTerminate 
}
