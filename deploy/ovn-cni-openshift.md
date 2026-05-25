# OVN-CNI support for OpenShift using NetScaler node controller

The node controller and router pod now support OVN-Kubernetes, the default CNI in OpenShift 4.x clusters. VXLAN overlay network is established between the OpenShift nodes and the Citrix ADC using the `ovn-k8s-mp0` interface.

## Deploy OpenShift with OVN-Kubernetes

### Prerequisites

-  OpenShift 4.x cluster with OVN-Kubernetes CNI
-  Citrix ADC (NetScaler) reachable from all cluster nodes
-  A dedicated subnet for VTEP overlay (must not overlap with pod or node CIDRs)
-  `kubectl`/`oc` CLI with cluster-admin access

### Step 1: Create the Namespace and Secret

```bash
oc new-project citrix-system

kubectl create secret generic nslogin \
  --from-literal=username='nsroot' \
  --from-literal=password='<your-adc-password>' \
  -n citrix-system
```

### Step 2: Create the Security Context Constraint (OpenShift Only)

Router pods require privileged access on OpenShift. Create a Security Context Constraint (SCC) binding for the service account:

```bash
oc adm policy add-scc-to-user privileged \
  system:serviceaccount:citrix-system:citrix-node-controller
```

### Step 3: Deploy the Node Controller

Download and edit `citrix-k8s-node-controller.yaml`. Set the following environment variables:

| Variable | Description | Example |
|---|---|---|
| `NS_IP` | Citrix ADC management IP (NSIP/SNIP/CLIP) | `10.10.10.1` |
| `NS_USER` / `NS_PASSWORD` | ADC credentials (through secret `nslogin`) | — |
| `NETWORK` | VTEP overlay subnet — must not overlap with pod/node CIDRs | `172.16.3.0/24` |
| `REMOTE_VTEPIP` | Citrix ADC SNIP used as VTEP endpoint | `10.10.10.2` |
| `VNID` | VXLAN VNI — must not conflict with existing VXLANs on ADC | `175` |
| `VXLAN_PORT` | VXLAN UDP port — must not conflict with existing VXLANs on ADC | `8472` |
| `CNI_TYPE` | **Set to `ovn` for OpenShift OVN-Kubernetes** | `ovn` |
| `CNC_ROUTER_NAME` | Router pod name prefix | `kube-cnc-router` |
| `CNC_CONFIGMAP` | ConfigMap name used for node state | `citrix-node-controller` |

Update the image references in `citrix-k8s-node-controller.yaml`:

```yaml
# Node controller
image: quay.io/netscaler/netscaler-k8s-node-controller:3.0.0

# Router pod (set via CNC_ROUTER_IMAGE env or input)
image: quay.io/netscaler/nsnc-router:2.0.0
```

Deploy:

```bash
kubectl apply -f citrix-k8s-node-controller.yaml -n citrix-system
```

### Step 4: Apply the ConfigMap

```bash
kubectl apply -f config_map.yaml -n citrix-system
```

The default `config_map.yaml` has an empty tolerations list. To schedule router pods on tainted nodes (for example infra/worker nodes with custom taints), edit `tolerations.json` accordingly.

### Step 5: Verify the Deployment

**Check node controller pod is running:**

```bash
kubectl get pods -n citrix-system
```

**Check router pods are created per node:**

```bash
kubectl get pods -n citrix-system | grep kube-cnc-router
```

**Check the CNC router ConfigMap is populated:**

```bash
kubectl get configmap citrix-node-controller -n citrix-system -o yaml
```

Each node must have entries: `Host-<node>`, `Node-<ip>`, `Mac-<ip>`, `Interface-<ip>`, `CNI-<ip>`.

**Verify on Citrix ADC:**

```
show vxlan
show bridgetable
show route
show ip
```

Expected: VXLAN tunnel, bridge/FDB entries per node, SNIP, and pod network routes are all present.

## Cleanup / Uninstall

```bash
# Delete the configmap first — this triggers ADC cleanup (routes, VXLAN, SNIP)
kubectl delete -f config_map.yaml -n citrix-system

# Then delete the controller
kubectl delete -f citrix-k8s-node-controller.yaml -n citrix-system
```

>**Important:**
>
> Always delete the ConfigMap **before** the deployment. Deleting the ConfigMap triggers the node controller to remove all VXLAN, SNIP, route, and bridge table configuration from the Citrix ADC. Deleting the deployment first leaves stale config on the ADC.

## Limitations

-  OVN-Kubernetes cluster network is queried from `network.config.openshift.io` (OpenShift API). If this API is unavailable, a fallback calculation is used and a warning is logged.
-  Bridge table entries are not cleaned up from the ADC during a full ConfigMap delete (only routes, VXLAN, and SNIP are removed).
