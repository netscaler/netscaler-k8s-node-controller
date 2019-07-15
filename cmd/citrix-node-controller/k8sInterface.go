package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"time"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config     *restclient.Config
	err        error
	podcount = 0
)

//This is interface for Kubernetes API Server
type KubernetesAPIServer struct {
	Suffix string
	Client kubernetes.Interface
}

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
}

type QueueUpdate struct {
	Key   string
	Force bool
}

/*
*************************************************************************************************
*   APIName :  ConvertPrefixLenToMask                                                           *
*   Input   :  Prefix Length. 								        *
*   Output  :  Return Net Mask in dotted decimal.	                                        *
*   Descr   :  This API takes Prefix length and generate coresponding dotted Decimal            *
*	       notation of net mask						  		*
*************************************************************************************************
 */
func ConvertPrefixLenToMask(prefixLen string) string {
	len, _ := strconv.Atoi(prefixLen)
	netmask := (uint32)(^(1<<(32-(uint32)(len)) - 1))
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, netmask)
	fmt.Println("NETMASK", bytes)
	netmaskdot := fmt.Sprintf("%d.%d.%d.%d", bytes[0], bytes[1], bytes[2], bytes[3])
	return netmaskdot
}

/*
*************************************************************************************************
*   APIName :  CreateK8sApiserverClient                                                         *
*   Input   :  Nil. 								              	*
*   Output  :  Return Kubernetes APIserver session.	                                        *
*   Descr   :  This API creates a session with kube api server which can be used for   		*
*              wathing  different events. Does not take any input as APi Func parameter.	*
*	       This API automatically get API server informations if the binary running  	*
*	       inside the cluster. If Binary is running outside cluster, cluster kube config    *
*              file must have to be in local nodes $HOME/.kube/config  location                 *
*************************************************************************************************
 */
func CreateK8sApiserverClient() (*KubernetesAPIServer, error) {
	klog.Info("[INFO] Creating API Client")
	api := &KubernetesAPIServer{}
	config, err = clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
	 	klog.Error("[WARNING] Citrix Node Controller Runs outside cluster")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
	 	        klog.Error("[ERROR] Did not find valid kube config info")
			klog.Fatal(err)
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

/*
*************************************************************************************************
*   APIName :  NodeWatcher                                                                      *
*   Input   :  Takes API server session called client.             			        *
*   Output  :  Invokes call back functions.	                                                *
*   Descr   :  This API is for watching the Nodes. API Monitors Kubernetes API server for Nodes *
*            events and store in node cache. Based on the events type, call back functions      *
*	     Will execute and perform the desired tasks.					*
*************************************************************************************************
 */
/*
func CitrixNodeWatcher(api *KubernetesAPIServer, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput) {

	nodeListWatcher := cache.NewListWatchFromClient(api.Client.Core().RESTClient(), "nodes", v1.NamespaceAll, fields.Everything())
	_, nodecontroller := cache.NewInformer(nodeListWatcher, &v1.Node{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			CoreHandler(api, obj, nil, "ADD", IngressDeviceClient, ControllerInputObj)
		},
		UpdateFunc: func(obj interface{}, newobj interface{}) {
			CoreHandler(api, obj, newobj, "UPDATE", IngressDeviceClient, ControllerInputObj)
		},
		DeleteFunc: func(obj interface{}) {
			CoreHandler(api, obj, nil, "DELETE", IngressDeviceClient, ControllerInputObj)
		},
	},
	)
	stop := make(chan struct{})
	go nodecontroller.Run(stop)
	return 
}
*/
/*
*************************************************************************************************
*   APIName :  Generate Next PodCIRIP                                                           *
*   Input   :  Podaddr in dotted decimal notation. 						*
*   Output  :  Return Net Mask in dotted decimal.	                                        *
*   Descr   :  This API takes Prefix length and generate coresponding dotted Decimal            *
*	       notation of net mask						  		*
*************************************************************************************************
 */
func GenerateNextPodAddr(PodAddr string) string{
	oct := strings.Split(PodAddr, ".")
	oct3, _ := strconv.Atoi(oct[3])
	if (oct3 >= 254) {
		klog.Errorf("[ERROR] Cannot increment the last octect of the IP as it is 254")
                return "Error"
        }
	oct3 = oct3 + 1
	nextaddr := fmt.Sprintf("%s.%s.%s.%d", oct[0], oct[1], oct[2], oct3)
	return nextaddr
}
/*
*************************************************************************************************
*   APIName :  GetNodeAddress                                           	                *
*   Input   :  Takes Node object.					             		*
*   Output  :  Return Internal IP, External IP and Hostname.					*
*   Descr   :  This API Gets the Address info of the Node if present 				*
*************************************************************************************************
 */
func GetNodeAddress(node v1.Node) (string, string, string){
        var InternalIP, ExternalIP, HostName string
        for _, addr := range node.Status.Addresses {
		if (addr.Type == "InternalIP"){
			InternalIP = addr.Address
        		klog.Info("[INFO] Internal IP of Node:\t", InternalIP)
		}else if (addr.Type == "Hostname"){
			HostName = addr.Address
        		klog.Info("[INFO] Host Name of Node:\t", HostName)
		}else if (addr.Type == "ExternalIP"){
			ExternalIP = addr.Address
        		klog.Info("[INFO] External IP  of Node:\t", ExternalIP)
		}
	}
	return InternalIP, ExternalIP, HostName
}
/*
*************************************************************************************************
*   APIName :  ParseNodeEvents                                                                  *
*   Input   :  Takes Node object, IngressDeviceObject and InputData.             		*
*   Output  :  Return Node Object.						                *
*   Descr   :  This API  Parses the object and prepare node object. 				*
*************************************************************************************************
 */
func ParseNodeEvents(api *KubernetesAPIServer, obj interface{}, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput) *Node {
	node := new(Node)
	node.Role = ""
	node.Label = ""
	originalObjJS, err := json.Marshal(obj)
	if err != nil {
		klog.Errorf("[ERROR] Failed to Marshal original object: %v", err)
	}
	var originalNode v1.Node
	if err = json.Unmarshal(originalObjJS, &originalNode); err != nil {
		klog.Errorf("[ERROR] Failed to unmarshal original object: %v", err)
	}
	if (originalNode.Labels["com.citrix.nodetype"] == "citrixadc"){ 
		node.Label = "citrixadc"
		klog.Info("[INFO] Processing Citrix Dummy Node")
	}
	PodCIDR := originalNode.Spec.PodCIDR
        InternalIP, ExternalIP, HostName := GetNodeAddress(originalNode)
	node.IPAddr = InternalIP
        node.HostName = HostName
        node.ExternalIPAddr = ExternalIP
	if (originalNode.Spec.Taints!=nil){
		klog.Info("[INFO] Taint Infromation", originalNode.Spec.Taints)
		ParseNodeRoles(node, originalNode)
		klog.Info("[INFO] Setting Node Role", node.Role)
	}
        if (PodCIDR != "" || node.Label == "citrixadc" || node.Role == "Master" ){
		if (PodCIDR != "") {
			klog.Info("[INFO] PodCIDR Information is Present: PodCIDR", PodCIDR)
			ParseNodeNetworkInfo(api, obj, IngressDeviceClient, ControllerInputObj, node, PodCIDR)
		}
		if (node.Label == "citrixadc") {
			klog.Info("[INFO] Add event for  Citrix ADC Node")
		}
		if (node.Role == "Master") {
			klog.Info("[INFO] Master Node events")
		}
	}else{
		klog.Errorf("[WARNING] Does not have PodCIDR Information, CNC will Generate itself")
		GenerateNodeNetworkInfo(api, obj, IngressDeviceClient, ControllerInputObj, node, originalNode, PodCIDR)
	} 
	return node
}
/*
*************************************************************************************************
*   APIName :  core_add_handler                                                                 *
*   Input   :  Takes Node object, IngressDeviceObject and InputData.             		*
*   Output  :  Every node addition, it creates a Route entry in Ingress Device.	                *
*   Descr   :  This API being Invoked when an Add node event comes.				*
*	       It parses the Node event object and calls route addition for the new Node.	*
*************************************************************************************************
 */
/*
func CoreAddHandler(api *KubernetesAPIServer, obj interface{}, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput) {
	node := ParseNodeEvents(api, obj, IngressDeviceClient, ControllerInputObj)
	if (node.Label != "citrixadc"){
		if (ControllerInputObj.CncOperation == "ADD"){
			NsInterfaceAddRoute(IngressDeviceClient, ControllerInputObj, node)
		}else{
			klog.Info("[INFO] Citrix Node Controller operation in input is not ADD")
		}
	}else {
		klog.Info("[INFO] Skipping Route addition for Dummy Node")
	}
}
*/
/*
*************************************************************************************************
*   APIName :  CoreDeleteHandler                                                                 *
*   Input   :  Takes Node object, IngressDeviceObject and InputData.             		*
*   Output  :  Every node addition, it creates a Route entry in Ingress Device.	                *
*   Descr   :  This API is for watching the Nodes. API Monitors Kubernetes API server for Nodes *
*            events and store in node cache. Based on the events type, call back functions      *
*	     Will execute and perform the desired tasks.					*
*************************************************************************************************
 */
/*
func CoreDeleteHandler(api *KubernetesAPIServer, obj interface{}, ingressDevice *NitroClient, controllerInput *ControllerInput) {
	node := ParseNodeEvents(api, obj, ingressDevice, controllerInput)
	if (node.Label != "citrixadc"){
		NsInterfaceDeleteRoute(ingressDevice, controllerInput, node)
	}else if ((controllerInput.State & NetscalerTerminate) != NetscalerTerminate){
		klog.Info("[ERROR] Citrix dummy node has been removed Manually")
	}
}
*/
/*
*************************************************************************************************
*   APIName :  CoreUpdateHandler                                                              *
*   Input   :  Takes Node object, IngressDeviceObject and InputData.             		*
*   Output  :  Every node addition, it creates a Route entry in Ingress Device.	                *
*   Descr   :  This API being Invoked when an Add node event comes.				*
*	       It parses the Node event object and calls route addition for the new Node.	*
*************************************************************************************************
 */
/*
func CoreUpdateHandler(api *KubernetesAPIServer, obj interface{}, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput) {
	node := ParseNodeEvents(api, obj, IngressDeviceClient, ControllerInputObj)
	fmt.Println("UPDATE HANDLER", node)
}
*/
/*
*************************************************************************************************
*   APIName :  CoreHandler                                                                     *
*   Input   :  Takes API server session called client.             			        *
*   Output  :  Invokes call back functions.	                                                *
*   Descr   :  This API is for watching the Nodes. API Monitors Kubernetes API server for Nodes *
*            events and store in node cache. Based on the events type, call back functions      *
*	     Will execute and perform the desired tasks.					*
*************************************************************************************************
 */
/*
func CoreHandler(api *KubernetesAPIServer, obj interface{}, newobj interface{}, event string, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput) {
	//create a slice of ops

	if event == "ADD" {
		CoreAddHandler(api, obj, IngressDeviceClient, ControllerInputObj)
	}
	if event == "DELETE" {
		CoreDeleteHandler(api, obj, IngressDeviceClient, ControllerInputObj)
	}
	if event == "UPDATE" {
		//	CoreUpdateHandler(obj, IngressDeviceClient, ControllerInputObj)
	}
}
*/
func GetClusterCNI(api *KubernetesAPIServer, controllerInput *ControllerInput) {
	pods, err := api.Client.Core().Pods("kube-system").List(metav1.ListOptions{})
	if err != nil {
		klog.Error("[ERROR] Error in Pod Listing", err)
	}
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, "flannel") {
			controllerInput.ClusterCNI = "Flannel"
		} else if strings.Contains(pod.Name, "weave") {
			controllerInput.ClusterCNI = "Weave"
		} else if strings.Contains(pod.Name, "calico") {
			controllerInput.ClusterCNI = "Calico"
		} else {
			controllerInput.ClusterCNI = "Flannel"
                }
	}
}
func ConfigDecider(api *KubernetesAPIServer, ingressDevice *NitroClient, controllerInput *ControllerInput) {
	if (controllerInput.ClusterCNI == "") {
		GetClusterCNI(api, controllerInput)
	}
	if controllerInput.ClusterCNI == "Flannel" {
		InitFlannel(api, ingressDevice, controllerInput)
	} else {
		klog.Info("[INFO] Network Automation is not supported for other than Flannel")
	}
}
/*
*************************************************************************************************
*   APIName :  ConfigMapInputWatcher                                                            *
*   Input   :  Takes API server session called client.             			        *
*   Output  :  Invokes call back functions.	                                                *
*   Descr   :  This API is for watching the Nodes. API Monitors Kubernetes API server for Nodes *
*            events and store in node cache. Based on the events type, call back functions      *
*	     Will execute and perform the desired tasks.					*
*************************************************************************************************
 */
func ConfigMapInputWatcher(api *KubernetesAPIServer, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput) {

	ConfigMapWatcher := cache.NewListWatchFromClient(api.Client.Core().RESTClient(), "configmaps", "citrix", fields.Everything())
	_, configcontroller := cache.NewInformer(ConfigMapWatcher, &v1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.Info("[INFO] CONFIG MAP ADD EVENT")
			HandleConfigMapAddEvent(api, obj, IngressDeviceClient, ControllerInputObj)
		},
		UpdateFunc: func(obj interface{}, newobj interface{}) {
			klog.Info("[INFO] UPDATE")
			HandleConfigMapUpdateEvent(api, obj, newobj, IngressDeviceClient, ControllerInputObj)
		},
		DeleteFunc: func(obj interface{}) {
			klog.Info("[INFO] Config Map is deleted, CNC clean UP the whole configurations which it has created", obj)
			HandleConfigMapDeleteEvent(api, obj, IngressDeviceClient, ControllerInputObj)
		},
	},
	)
	stop := make(chan struct{})
	defer close(stop)
	go configcontroller.Run(stop)
	select {}
	return
}
func CheckAndWaitForNetscalerInit(ControllerInputObj *ControllerInput){
	if ((ControllerInputObj.State & NetscalerInit) != NetscalerInit){
		klog.Info("[DEBUG] Waiting for NetScaler initialization to complete")
	}
	for {
		if ((ControllerInputObj.State & NetscalerInit) == NetscalerInit){
			break;
		}
	}
}

func HandleConfigMapUpdateEvent(api *KubernetesAPIServer, obj interface{}, newobj interface{}, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput){
	node := new(Node)
	klog.Info("[OLD OBJECT]", obj)
	klog.Info("\n[NEW OBJECT]", newobj)
	ConfigMapDataNew := make(map[string]string)
        ConfigMapDataNew = newobj.(*v1.ConfigMap).Data
	ConfigMapDataOld := make(map[string]string)
        ConfigMapDataOld = obj.(*v1.ConfigMap).Data
        if (ConfigMapDataNew["operation"] == "ADD" && ConfigMapDataOld["operation"] == "ADD"){
		for key, value := range ConfigMapDataNew {
			if (strings.Contains(value, ".") && strings.Contains(ConfigMapDataNew[value], ":")) {
				if newval, ok := ConfigMapDataOld[key]; !ok {
					klog.Info("[INFO] Key Value", key, value, newval)
					node.IPAddr = key
        				node.PodAddress = value
        				node.PodVTEP = ConfigMapDataNew[node.PodAddress]
        				node.PodNetMask = ConvertPrefixLenToMask("24")
        				node.PodMaskLen = "24"
					NsInterfaceAddRoute( IngressDeviceClient, ControllerInputObj, node)
				}
			}
		}
	}	
        if (ConfigMapDataNew["operation"] == "ADD" && ConfigMapDataOld["operation"] == "DELETE"){
		ConfigDecider(api, IngressDeviceClient, ControllerInputObj)
		HandleConfigMapAddEvent(api, newobj, IngressDeviceClient, ControllerInputObj)
		for key, value := range ConfigMapDataNew {
			if (strings.Contains(value, ".") && strings.Contains(ConfigMapDataNew[value], ":")) {
					klog.Info("[INFO] Key Value", key, value)
					node.IPAddr = key
        				node.PodAddress = value
        				node.PodVTEP = ConfigMapDataNew[node.PodAddress]
        				node.PodNetMask = ConvertPrefixLenToMask("24")
        				node.PodMaskLen = "24"
					NsInterfaceAddRoute( IngressDeviceClient, ControllerInputObj, node)
			}
		}
	}	
        if (ConfigMapDataNew["operation"] == "DELETE" && ConfigMapDataOld["operation"] == "ADD"){
		HandleConfigMapDeleteEvent(api, obj, IngressDeviceClient, ControllerInputObj)
	}	
        if (ConfigMapDataNew["operation"] == "DELETE" && ConfigMapDataOld["operation"] == "DELETE"){
	}	
}

func HandleConfigMapAddEvent(api *KubernetesAPIServer, obj interface{}, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput){
	ControllerInputObj.CncOperation = "ADD"
	command := []string{"/bin/bash", "-c"}
	args := []string{
                    "vtepmac=`ifconfig flannel.1 | grep -o -E '([[:xdigit:]]{1,2}:){5}[[:xdigit:]]{1,2}' `; echo \"InterfaceInfo ${vtepmac}\"; theIPaddress=`ip -4 addr show flannel.1  | grep inet | awk '{print $2}' | cut -d/ -f1`;  hostip=`ip -4 addr show eth0  | grep inet | awk '{print $2}' | cut -d/ -f1`; echo \"IP Addredd ${theIPaddress}\"; echo \"Host IP Address ${hostip}\"; `kubectl patch configmap citrix-node-controller  -p '{\"data\":{\"'\"$theIPaddress\"'\": \"'\"$vtepmac\"'\"}}'`;  `kubectl patch configmap citrix-node-controller  -p '{\"data\":{\"'\"$hostip\"'\": \"'\"$theIPaddress\"'\"}}'`;  ip route add ${network}  via  ${nexthop} dev flannel.1 onlink; arp -s ${nexthop}  ${ingmac}  dev flannel.1;bridge fdb add ${ingmac} dev flannel.1 dst ${vtepip}; sleep 3d;"}
	
        SecurityContext := new(v1.SecurityContext)
	Capabilities := new(v1.Capabilities)
	Capabilities.Add = append(Capabilities.Add, "NET_ADMIN")
	SecurityContext.Capabilities = Capabilities
	
	DaemonSet := &appv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "citrixrouteaddpod",
			Namespace: "citrix",
			Labels: map[string]string{
				"app":  "citrixrouteaddpod",
			},
		},
		Spec: appv1.DaemonSetSpec{
			MinReadySeconds: 2,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":	"citrixrouteaddpod",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "citrixrouteaddpod",
					},
					Name: "citrixrouteaddpod",
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "citrix-node-controller",
					HostNetwork: true,
					Containers: []v1.Container{
						{
							Name:  "citrixdummypod"+strconv.Itoa(podcount),
							Image: "quay.io/citrix/dummynode:latest",	
							Command: command,
							Args: args,
							SecurityContext: SecurityContext,
							Env: []v1.EnvVar{
								{Name: "network", Value: ControllerInputObj.IngressDevicePodSubnet},
								{Name: "nexthop", Value: ControllerInputObj.IngressDevicePodIP},
								{Name: "ingmac", Value: ControllerInputObj.IngressDeviceVtepMAC},
								{Name: "vtepip", Value: ControllerInputObj.IngressDeviceVtepIP},
							},    
						},
					},
				},
			},
		},
	}
	_, err := api.Client.AppsV1().DaemonSets("citrix").Create(DaemonSet)
	if err != nil {
		klog.Error("[ERROR] Failed to create daemon set:", err)
	}
	CLeanupHandler(api, "citrixroutecleanuppod")
}
func CLeanupHandler(api *KubernetesAPIServer, DaemonSet string){
	ds, err := api.Client.AppsV1().DaemonSets("citrix").Get(DaemonSet, metav1.GetOptions{})
	if (ds != nil){
		falseVar := false
		deleteOptions := &metav1.DeleteOptions{OrphanDependents: &falseVar}
		err = api.Client.AppsV1().DaemonSets("citrix").Delete(ds.Name, deleteOptions)
	}
	if (err != nil) {
		fmt.Print(err)
	}
}

func HandleConfigMapDeleteEvent(api *KubernetesAPIServer, obj interface{}, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput){
	ControllerInputObj.CncOperation = "DELETE"
	CheckAndWaitForNetscalerInit(ControllerInputObj)
	command := []string{"/bin/bash", "-c"}
	args := []string{
                    "vtepmac=`ifconfig flannel.1 | grep -o -E '([[:xdigit:]]{1,2}:){5}[[:xdigit:]]{1,2}' `; echo \"InterfaceInfo ${vtepmac}\"; theIPaddress=`ip -4 addr show flannel.1  | grep inet | awk '{print $2}' | cut -d/ -f1`;  hostip=`ip -4 addr show eth0  | grep inet | awk '{print $2}' | cut -d/ -f1`; echo \"IP Addredd ${theIPaddress}\"; echo \"Host IP Address ${hostip}\";ip route delete ${network}  via  ${nexthop} dev flannel.1 onlink; arp -d ${nexthop}  dev flannel.1; bridge fdb delete ${ingmac} dev flannel.1 dst ${vtepip}; sleep 3d;"}
	
        SecurityContext := new(v1.SecurityContext)
	Capabilities := new(v1.Capabilities)
	Capabilities.Add = append(Capabilities.Add, "NET_ADMIN")
	SecurityContext.Capabilities = Capabilities
	
	DaemonSet := &appv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "citrixroutecleanuppod",
			Namespace: "citrix",
			Labels: map[string]string{
				"app":  "citrixroutecleanuppod",
			},
		},
		Spec: appv1.DaemonSetSpec{
			MinReadySeconds: 2,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":	"citrixroutecleanuppod",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "citrixroutecleanuppod",
					},
					Name: "citrixroutecleanuppod",
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "citrix-node-controller",
					HostNetwork: true,
					Containers: []v1.Container{
						{
							Name:  "citrixdummypod"+strconv.Itoa(podcount),
							Image: "quay.io/citrix/dummynode:latest",	
							Command: command,
							Args: args,
							SecurityContext: SecurityContext,
							Env: []v1.EnvVar{
								{Name: "network", Value: ControllerInputObj.IngressDevicePodSubnet},
								{Name: "nexthop", Value: ControllerInputObj.IngressDevicePodIP},
								{Name: "ingmac", Value: ControllerInputObj.IngressDeviceVtepMAC},
								{Name: "vtepip", Value: ControllerInputObj.IngressDeviceVtepIP},
							},    
						},
					},
				},
			},
		},
	}
	_, err := api.Client.AppsV1().DaemonSets("citrix").Create(DaemonSet)
	if err != nil {
		klog.Error("[ERROR] Failed to create daemon set:", err)
	}
	
	ClearAllRoutes(api, obj, IngressDeviceClient, ControllerInputObj)

	TerminateFlannel(api, IngressDeviceClient, ControllerInputObj)
	CLeanupHandler(api, "citrixrouteaddpod")
}

func ClearAllRoutes(api *KubernetesAPIServer, obj interface{}, ingressDevice *NitroClient, controllerInput *ControllerInput){
	node := new(Node)
	ConfigMapData := make(map[string]string)
        ConfigMapData = obj.(*v1.ConfigMap).Data
        klog.Info("JANRAJ CONFIG MAP DATA", ConfigMapData)
	for key, value := range ConfigMapData {
		if (strings.Contains(value, ".")) {
			klog.Info("[INFO] Key Value", key, value)
        		node.PodAddress = value
        		node.PodVTEP = ConfigMapData[node.PodAddress]
        		node.PodNetMask = ConvertPrefixLenToMask("24")
        		node.PodMaskLen = "24"
			NsInterfaceDeleteRoute(ingressDevice, controllerInput, node)
		}
	}	
}
/*
*************************************************************************************************
*   APIName :  ParseNodeNetworkInfo                                                             *
*   Input   :  Takes API server session called client.             			        *
*   Output  :  Invokes call back functions.	                                                *
*   Descr   :  This API is for watching the Nodes. API Monitors Kubernetes API server for Nodes *
*            events and store in node cache. Based on the events type, call back functions      *
*	     Will execute and perform the desired tasks.					*
*************************************************************************************************
 */
func ParseNodeNetworkInfo(api *KubernetesAPIServer, obj interface{}, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput, node *Node, PodCIDR string) {
	splitString := strings.Split(PodCIDR, "/")
	address, masklen := splitString[0], splitString[1]
	backendData := []byte(obj.(*v1.Node).Annotations["flannel.alpha.coreos.com/backend-data"])
	vtepMac := make(map[string]string)
	err = json.Unmarshal(backendData, &vtepMac)
	if err != nil {
		klog.Error("[ERROR] Issue with Json unmarshel", err)
	}
	if (node.HostName != ""){
		node.HostName = "Citrix"
	}
	if (node.IPAddr != ""){
		node.IPAddr = obj.(*v1.Node).Annotations["flannel.alpha.coreos.com/public-ip"]
	}
	node.PodVTEP = vtepMac["VtepMAC"]
	node.PodAddress = address
	NextPodAddress := GenerateNextPodAddr(address)
	if (NextPodAddress != "Error"){
		node.NextPodAddress = NextPodAddress
	}else{
		node.NextPodAddress = address
	}
	node.PodNetMask = ConvertPrefixLenToMask(masklen)
	node.PodMaskLen = masklen
	node.Type = obj.(*v1.Node).Annotations["flannel.alpha.coreos.com/backend-type"]
	ControllerInputObj.NodesInfo[node.IPAddr] = node

}
/*
*************************************************************************************************
*   APIName :  GenerateNodeInfo                                                            	*
*   Input   :  Takes API server session called client.             			        *
*   Output  :  Invokes call back functions.	                                                *
*   Descr   :  This API is for watching the Nodes. API Monitors Kubernetes API server for Nodes *
*            events and store in node cache. Based on the events type, call back functions      *
*	     Will execute and perform the desired tasks.					*
*************************************************************************************************
 */
func GenerateNodeNetworkInfo(api *KubernetesAPIServer, obj interface{}, IngressDeviceClient *NitroClient, ControllerInputObj *ControllerInput, node *Node, originalNode v1.Node, PodCIDR string) {
	podcount = podcount + 1
	klog.Info("[INFO] Generating PODCIDR and Node Information")
	patchBytes := []byte(fmt.Sprintf(`{"metadata":{"labels":{"NodeIP":"%s"}}}`, node.IPAddr))
	if (node.IPAddr == ""){
		patchBytes = []byte(fmt.Sprintf(`{"metadata":{"labels":{"NodeIP":"%s"}}}`, node.ExternalIPAddr))
	}
        time.Sleep(10 * time.Second) //TODO, We have to wait till Node is available.
        if _, err = api.Client.CoreV1().Nodes().Patch(originalNode.Name, types.StrategicMergePatchType, patchBytes); err != nil {  
            	klog.Errorf("[ERROR] Failed to Patch label %v",err)
        }else {
            	klog.Info("[INFO] Updated node  label")
	}
	command := []string{"/bin/bash", "-c"}
	args := []string{
                    "vtepmac=`ifconfig flannel.1 | grep -o -E '([[:xdigit:]]{1,2}:){5}[[:xdigit:]]{1,2}' `; echo \"InterfaceInfo ${vtepmac}\"; theIPaddress=`ip -4 addr show flannel.1  | grep inet | awk '{print $2}' | cut -d/ -f1`;  hostip=`ip -4 addr show eth0  | grep inet | awk '{print $2}' | cut -d/ -f1`; echo \"IP Addredd ${theIPaddress}\"; echo \"Host IP Address ${hostip}\"; `kubectl patch configmap citrix-node-controller  -p '{\"data\":{\"'\"$theIPaddress\"'\": \"'\"$vtepmac\"'\"}}'`;  `kubectl patch configmap citrix-node-controller  -p '{\"data\":{\"'\"$hostip\"'\": \"'\"$theIPaddress\"'\"}}'`;  ip route add ${network}  via  ${nexthop} dev flannel.1 onlink; arp -s ${nexthop}  ${ingmac}  dev flannel.1; bridge fdb add ${ingmac} dev flannel.1 dst ${vtepip}; sleep 3d;"}
	
        SecurityContext := new(v1.SecurityContext)
	Capabilities := new(v1.Capabilities)
	Capabilities.Add = append(Capabilities.Add, "NET_ADMIN")
	SecurityContext.Capabilities = Capabilities
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "citrixdummypod"+strconv.Itoa(podcount),
		},
		Spec: v1.PodSpec{
			ServiceAccountName: "citrix-node-controller",
			HostNetwork: true,
			Containers: []v1.Container{
				{
					Name:  "citrixdummypod"+strconv.Itoa(podcount),
					Image: "quay.io/citrix/dummynode:latest",	
					Command: command,
					Args: args,
					SecurityContext: SecurityContext,
					Env: []v1.EnvVar{
						{Name: "network", Value: ControllerInputObj.IngressDevicePodSubnet},
						{Name: "nexthop", Value: ControllerInputObj.IngressDevicePodIP},
						{Name: "ingmac", Value: ControllerInputObj.IngressDeviceVtepMAC},
						{Name: "vtepip", Value: ControllerInputObj.IngressDeviceVtepIP},
					},    
				},
			},
		},
	}
	nodeSelector :=  make(map[string]string)
	nodeSelector["NodeIP"] = node.IPAddr
	pod.Spec.NodeSelector = nodeSelector
        //time.Sleep(10 * time.Second) //TODO, We have to wait till Pod is available.
        //if _, err = api.Client.CoreV1().Pods("citrix").Create(pod); err != nil {  
        //    	klog.Error("Failed to Create a Pod " + err.Error())
        //}
        time.Sleep(60 * time.Second) //TODO, We have to wait till Node is available.

	//pod, err = api.Client.CoreV1().Pods("citrix").Get(pod.Name, metav1.GetOptions{})
	//if err != nil {
	//	fmt.Errorf("pod Get API error: %v \n pod: %v", err, pod)
	//}
	configMaps, err := api.Client.CoreV1().ConfigMaps("citrix").Get("citrix-node-controller", metav1.GetOptions{})
	if err != nil {
		fmt.Errorf("ConfigMap Get API error: %v \n pod: %v", configMaps, err)
	}
	if (configMaps != nil) {
		ConfigMapData := make(map[string]string)
		ConfigMapData = configMaps.Data
		klog.Info("CONFIG MAP DATA", ConfigMapData)
		node.PodAddress = ConfigMapData[node.IPAddr]
		node.PodVTEP = ConfigMapData[node.PodAddress]
		node.PodNetMask = ConvertPrefixLenToMask("24")
	        node.PodMaskLen = "24"
	}else {
		 klog.Error("Config MAP is Empty \n")
	}
}
/*
*************************************************************************************************
*   APIName :  GenerateNodeInfo                                                            	*
*   Input   :  Takes API server session called client.             			        *
*   Output  :  Invokes call back functions.	                                                *
*   Descr   :  This API is for watching the Nodes. API Monitors Kubernetes API server for Nodes *
*            events and store in node cache. Based on the events type, call back functions      *
*	     Will execute and perform the desired tasks.					*
*************************************************************************************************
 */
func ParseNodeRoles(node *Node, originalNode v1.Node){
        for _, Role := range originalNode.Spec.Taints {
		if (Role.Key == "node-role.kubernetes.io/master"){
			node.Role = "Master"	
		}
	}
}

/*
func IsNodeTolerable(taints []apiv1.Taint, tolerations []apiv1.Toleration) bool {
	for _, taint := range taints {
		var taintIsTolerated bool
		for _, toleration := range tolerations {
			if taint.Key == toleration.Key && taint.Value == toleration.Value {
				taintIsTolerated = true
				break
			}
		}
		if !taintIsTolerated {
			return false
		}
	}
	return true
}
func waitForPodsToComeUP(){
	var nodesMissingPod []string
	for {
		nodesMissingPod, err := getNodesMissingPod()
		if err != nil {
			klog.Errorf("[ERROR] Failed Getting Missing Node Info. Retrying.", err)
			continue
		}
		if len(nodesMissingPod) <= 0 {
			break
		}
	}
}


func getNodesMissingPod() ([]string, error) {

	var nodesMissingDSPods []string

	nodes, err := dsc.client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nodesMissingDSPods, err
	}

	pods, err := dsc.client.CoreV1().Pods(dsc.Namespace).List(metav1.ListOptions{
		IncludeUninitialized: true,
		LabelSelector:        "app=" + dsc.DaemonSetName + ",source=kuberhealthy",
	})
	if err != nil {
		return nodesMissingDSPods, err
	}

	// populate a node status map. default status is "false", meaning there is
	// not a pod deployed to that node.  We are only adding nodes that tolerate
	// our list of dsc.tolerations
	nodeStatuses := make(map[string]bool)
	for _, n := range nodes.Items {
		if taintsAreTolerated(n.Spec.Taints, dsc.tolerations) {
			nodeStatuses[n.Name] = false
		}
	}

	// Look over all daemonset pods.  Mark any hosts that host one of the pods
	// as "true" in the nodeStatuses map, indicating that a daemonset pod is
	// deployed there.
	//Additionally, only look on nodes with taints that we tolerate
	for _, pod := range pods.Items {
		// the pod should be ready
		if pod.Status.Phase != "Running" {
			continue
		}
		for _, node := range nodes.Items {
			for _, nodeip := range node.Status.Addresses {
				// We are looking for the Internal IP and it needs to match the host IP
				if nodeip.Type != "InternalIP" || nodeip.Address != pod.Status.HostIP {
					continue
				}
				if taintsAreTolerated(node.Spec.Taints, dsc.tolerations) {
					nodeStatuses[node.Name] = true
					break
				}
			}
		}
	}

	// pick out all the nodes without daemonset pods on them and
	// add them to the final results
	for nodeName, hasDS := range nodeStatuses {
		if !hasDS {
			nodesMissingDSPods = append(nodesMissingDSPods, nodeName)
		}
	}

	return nodesMissingDSPods, nil
}
func AddAllRoutes(api *KubernetesAPIServer, obj interface{}, ingressDevice *NitroClient, controllerInput *ControllerInput){
	node := new(Node)
	ConfigMapData := make(map[string]string)
        ConfigMapData = obj.(*v1.ConfigMap).Data
        klog.Info("CONFIG MAP DATA", ConfigMapData)
	for key, value := range ConfigMapData {
		if (strings.Contains(value, ".")) {
			klog.Info("[INFO] Key Value", key, value)
        		node.PodAddress = value
        		node.PodVTEP = ConfigMapData[node.PodAddress]
        		node.PodNetMask = ConvertPrefixLenToMask("24")
        		node.PodMaskLen = "24"
			NsInterfaceAddRoute(ingressDevice, controllerInput, node)
		}
	}	
}
*/
