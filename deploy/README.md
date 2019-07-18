# Deploy the Citrix k8s node controller

Citrix k8s node controller is controlled using a [config map](https://github.com/citrix/citrix-k8s-node-controller/blob/master/deploy/config_map.yaml). The [config map](https://github.com/citrix/citrix-k8s-node-controller/blob/master/deploy/config_map.yaml) file contains a `data.operation:` field that you can use to define Citrix k8s node controller to automatically create, apply, and delete routing configuration on Citrix ADC. You can use the following values for the `data.operation:` field:

| **Value** | **Description** |
| ----- | ----------- |
| ADD | Citrix k8s node controller creates a routing configuration on the Citrix ADC instance. |
| DELETE | Citrix k8s node controller deletes the routing configuration on the Citrix ADC instance. |

[config_map.yaml](https://github.com/citrix/citrix-k8s-node-controller/blob/master/deploy/config_map.yaml):

```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: citrix-node-controller
  namespace: citrix
data:
  operation: "ADD"
```

## Deploy the Citrix k8s node controller

Perform the following:

1.  Download the `citrix-k8s-node-controller.yaml` deployment file using the following command:

        wget  https://raw.githubusercontent.com/citrix/citrix-k8s-node-controller/master/deploy/citrix-k8s-node-controller.yaml

    The deployment file contains definitions for the following:

    -  Cluster Role (`ClusterRole`)

    -  Cluster Role Bindings (`ClusterRoleBinding`)

    -  Service Account (`ServiceAccount`)

    -  Citrix Node Controller service (`citrix-node-controller`)

    You don't have to modify the definitions for `ClusterRole`, `ClusterRoleBinding`, and `ServiceAccount` definitions. The definitions are used by Citrix node controller to monitor Kubernetes events. But, in the`citrix-node-controller` definition you have to provide the values for the environment variables that is required for Citrix k8s node controller to configure the Citrix ADC.

    You must provide values for the following environment variables in the Citrix k8s node controller service definition:

    | Environment Variable | Mandatory or Optional | Description |
    | -------------------- | --------------------- | ----------- |
    | NS_IP | Mandatory | Citrix k8s node controller uses this IP address to configure the Citrix ADC. The NS_IP can be anyone of the following: </br>- SNIP for high availability and standalone deployments (Ensure that management access is enabled) </br> - CLIP for Cluster deployments |
    | NS_USER and NS_PASSWORD | Mandatory | The user name and password of Citrix ADC. Citrix k8s node controller uses these credentials to authenticate with Citrix ADC. You can either provide the user name and password or Kubernetes secrets. If you want to use a non-default Citrix ADC user name and password, you can [create a system user account in Citrix ADC](https://developer-docs.citrix.com/projects/citrix-k8s-ingress-controller/en/latest/deploy/deploy-cic-yaml/#create-system-user-account-for-citrix-ingress-controller-in-citrix-adc). </br> The deployment file uses Kubernetes secrets, create a secret for the user name and password using the following command: </br> `kubectl create secret  generic nslogin --from-literal=username='nsroot' --from-literal=password='nsroot'` </br> **Note**: If you want to use a different secret name other than `nslogin`, ensure that you update the `name` field in the `citrix-node-controller` definition. |
    | NODE_CNI_CIDR | Mandatory | Provide the node [CIDR](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing) of the Kubernetes cluster. Use the following command to view the node CIDR: </br> `cat /run/flannel/subnet.env` </br> The node CIDR is displayed as `FLANNEL_NETWORK`.|
    | NS_POD_CIDR | Mandatory | Provide a pod CIDR from the node CIDR in the Kubernetes cluster to create an overlay network between Citrix ADC and Kubernetes cluster. </br> For example, if the node CIDR in the Kubernetes cluster is `10.244.0.0/16` and the pod CIDRs of the nodes are `10.244.0.1/24`, `10.244.1.1/24`, `10.244.2.1/24`. You can provide a pod CIDR `10.244.254.1/24` that is not allocated to the nodes.|
    | NS_VTEP_MAC | Mandatory | Provide [VMAC](https://docs.citrix.com/en-us/netscaler/12/system/high-availability-introduction/configuring-virtual-mac-addresses-high-availability.html) that you have configured on the Citrix ADC as an interface towards your Kubernetes cluster. |
    | NS_NETPROFILE | Mandatory | Provide the network profile (netprofile) name that you have used in the Citrix ingress controller.|
    | NS_VXLAN_ID | Optional | This argument is only applicable for Flannel CNI. If Flannel uses a different `VXLAN_ID`, Use this argument to provide the `VXLAN_ID` |
    | K8S_VXLAN_PORT | Optional | If the Kubernetes cluster VXLAN port is other than 8472, you have to provide the Kubernetes VXLAN port number using this parameter. |

1.  After you have updated the Citrix k8s node controller deployment YAML file, deploy it using the following command:

        kubectl create -f citrix-k8s-node-controller.yaml

1.  Apply the [config map](https://github.com/citrix/citrix-k8s-node-controller/blob/master/deploy/config_map.yaml) using the following command:

        kubectl apply -f https://raw.githubusercontent.com/citrix/citrix-k8s-node-controller/master/deploy/config_map.yaml
