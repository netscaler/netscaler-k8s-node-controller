package main

import (
	"encoding/binary"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config     *restclient.Config
	err        error
	podcount   = 0
)

// This is interface for Kubernetes API Server
type KubernetesAPIServer struct {
	Suffix string
	Client kubernetes.Interface
}

// This for go client
type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
}

type QueueUpdate struct {
	Key   string
	Force bool
}

// ConvertPrefixLenToMask convert the prefix len to netmask (dotted) format.
func ConvertPrefixLenToMask(prefixLen string) string {
	len, _ := strconv.Atoi(prefixLen)
	netmask := (uint32)(^(1<<(32-(uint32)(len)) - 1))
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, netmask)
	fmt.Println("NETMASK", bytes)
	netmaskdot := fmt.Sprintf("%d.%d.%d.%d", bytes[0], bytes[1], bytes[2], bytes[3])
	return netmaskdot
}

// This creates go client.
func CreateK8sApiserverClient() (*KubernetesAPIServer, error) {
	klog.Info("[INFO] Creating API Client")
	api := &KubernetesAPIServer{}
	config, err = clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		klog.Error("[WARNING] Citrix Node Controller Runs outside cluster")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.Error("[ERROR] Did not find valid kube config info")
			return nil, err
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Error("[ERROR] Failed to establish connection")
		klog.Fatal(err)
	}
	klog.Info("[INFO] Kubernetes Client is created")
	api.Client = client
	return api, nil
}



//ConfigDecider function choose the overlay mechanism for establish route between cluster and Netscaler ADC.
// Flannel and Canal it uses VXLAN. 
func ConfigDecider(api *KubernetesAPIServer, ingressDevice *NitroClient, controllerInput *ControllerInput) {
	klog.Info("[INFO] CNI Detected on the cluster is", controllerInput.ClusterCNI)
	InitFlannel(api, ingressDevice, controllerInput)
}

// ConfigMapInputWatcher creates a watch goroutine for configmaps ADD, DELETE and UPDATE events.
// Function takes api server, ingress client and user input as arguments.
func ConfigMapInputWatcher(api *KubernetesAPIServer, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput) {
	ConfigMapWatcher := cache.NewListWatchFromClient(api.Client.Core().RESTClient(), "configmaps", v1.NamespaceAll, fields.Everything())
	_, configcontroller := cache.NewInformer(ConfigMapWatcher, &v1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			HandleConfigMapAddEvent(api, obj, IngressDeviceClient, ControllerInputObj)
		},
		UpdateFunc: func(obj interface{}, newobj interface{}) {
			HandleConfigMapUpdateEvent(api, obj, newobj, IngressDeviceClient, ControllerInputObj)
		},
		DeleteFunc: func(obj interface{}) {
			HandleConfigMapDeleteEvent(api, obj, IngressDeviceClient, ControllerInputObj)
		},
	},
	)
	stop := make(chan struct{})
	defer close(stop)
	go configcontroller.Run(stop)
	select {}
}

func HandleConfigMapUpdateEvent(api *KubernetesAPIServer, obj interface{}, newobj interface{}, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput) {
	node := new(Node)
	ConfigMapDataNew := make(map[string]string)
	ConfigMapDataNew = newobj.(*v1.ConfigMap).Data
	ConfigMapDataOld := make(map[string]string)
	ConfigMapDataOld = obj.(*v1.ConfigMap).Data
	if _, ok := ConfigMapDataNew["EndpointIP"]; ok {
		klog.Info("[INFO] CONFIG MAP UPDATE EVENT Old Object", obj, "New Object", newobj)
		for key, value := range ConfigMapDataNew {
			if _, ok := ConfigMapDataOld[key]; !ok {
				if (strings.Contains(key, "Node")){
					klog.Info("[INFO] Key Value", key, value)
					node.IPAddr = value
					kv := strings.Split(key, "-")
					node.PodAddress = ConfigMapDataNew["Interface-"+kv[1]]
					node.PodVTEP = ConfigMapDataNew["Mac-"+kv[1]]
					cni := ConfigMapDataNew["CNI-"+kv[1]]
					Network := strings.Split(cni, "/")
					fmt.Println("[INFO] Interface Mac CNI", node.PodAddress, node.PodVTEP, cni)
					if (len(Network) == 2){
						node.PodNetwork = Network[0]
						node.PodMaskLen = Network[1]
					}else{
						klog.Error("[ERROR] Could not fetch Network, need enhancements", Network)
					}
					node.PodNetMask = ConvertPrefixLenToMask(node.PodMaskLen)
					NsInterfaceAddRoute(IngressDeviceClient, ControllerInputObj, node)
					NodeList = append(NodeList, node)
				}
			}
		}
		for key, value := range ConfigMapDataOld {
			if _, ok := ConfigMapDataNew[key]; !ok {
				if (strings.Contains(key, "Node")){
					klog.Info("[INFO] Key Value", key, value)
					node.IPAddr = value
					kv := strings.Split(key, "-")
					node.PodAddress = ConfigMapDataOld["Interface-"+kv[1]]
					node.PodVTEP = ConfigMapDataOld["Mac-"+kv[1]]
					Network := strings.Split(ConfigMapDataOld["CNI-"+kv[1]], "/")
					node.PodNetwork = Network[0]
					node.PodMaskLen = Network[1]
					node.PodNetMask = ConvertPrefixLenToMask(node.PodMaskLen)
					NsInterfaceDeleteRoute(IngressDeviceClient, node)
				}
			}
		}
	}
}

func HandleConfigMapAddEvent(api *KubernetesAPIServer, obj interface{}, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput) {
	ControllerInputObj.CncOperation = "ADD"
	ConfigMapData := make(map[string]string)
	MetaData := obj.(*v1.ConfigMap).ObjectMeta
	ConfigMapData = obj.(*v1.ConfigMap).Data
	if _, ok := ConfigMapData["EndpointIP"]; ok {
		klog.Info("[INFO] CONFIG MAP ADD EVENT Obect", obj)
		ControllerInputObj.IngressDevicePodIP = ConfigMapData["EndpointIP"]
		ConfigDecider(api, IngressDeviceClient, ControllerInputObj)
		for key, value := range ConfigMapData {
			if (strings.Contains(key, "Node")){
				node := new(Node)
				klog.Info("[INFO] Key Value", key, value)
				node.IPAddr = value
				kv := strings.Split(key, "-")
				node.PodAddress = ConfigMapData["Interface-"+kv[1]]
				node.PodVTEP = ConfigMapData["Mac-"+kv[1]]
				Network := strings.Split(ConfigMapData["CNI-"+kv[1]], "/")
				node.PodNetwork = Network[0]
				node.PodMaskLen = Network[1]
				node.PodNetMask = ConvertPrefixLenToMask(node.PodMaskLen)
				NsInterfaceAddRoute(IngressDeviceClient, ControllerInputObj, node)
				NodeList = append(NodeList, node)
			}
		}
	}else if MetaData.Name == "citrix-node-controller"{
		ConfigDecider(api, IngressDeviceClient, ControllerInputObj)
		for id, _ := range NodeList {
			NsInterfaceAddRoute(IngressDeviceClient, ControllerInputObj, NodeList[id])
        	}
	}
}


func HandleConfigMapDeleteEvent(api *KubernetesAPIServer, obj interface{}, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput) {
	ControllerInputObj.CncOperation = "DELETE"
	node := new(Node)
	MetaData := obj.(*v1.ConfigMap).ObjectMeta
	ConfigMapData := make(map[string]string)
	ConfigMapData = obj.(*v1.ConfigMap).Data
	if _, ok := ConfigMapData["EndpointIP"]; ok {
		klog.Info("[INFO] CONFIG MAP DELETE EVENT Obect", obj)
		ControllerInputObj.IngressDevicePodIP = ConfigMapData["EndpointIP"]
		for key, value := range ConfigMapData {
			if (strings.Contains(key, "Node")){
				klog.Info("[INFO] Key Value", key, value)
				node.IPAddr = value
				kv := strings.Split(key, "-")
				node.PodAddress = ConfigMapData["Interface-"+kv[1]]
				node.PodVTEP = ConfigMapData["Mac-"+kv[1]]
				Network := strings.Split(ConfigMapData["CNI-"+kv[1]], "/")
				node.PodNetwork = Network[0]
				node.PodMaskLen = Network[1]
				node.PodNetMask = ConvertPrefixLenToMask(node.PodMaskLen)
				NsInterfaceDeleteRoute(IngressDeviceClient, node)
			}
		}
		TerminateFlannel(api, IngressDeviceClient, ControllerInputObj)
	}else if MetaData.Name == "citrix-node-controller"{
		for id, _ := range NodeList {
			NsInterfaceDeleteRoute(IngressDeviceClient, NodeList[id])
        	}
		TerminateFlannel(api, IngressDeviceClient, ControllerInputObj)
	}
}
