package main

import (
	"crypto/tls"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"log"
	"net/http"
	"strings"
)

type NitroClient struct {
	url       string
	statsURL  string
	username  string
	password  string
	proxiedNs string
	client    *http.Client
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
func NewNitroClient(obj *ControllerInput) *NitroClient {
	c := new(NitroClient)
	c.url = "https://" + strings.Trim(obj.IngressDeviceIP, " /") + "/nitro/v1/config/"
	c.username = obj.IngressDeviceUsername
	c.password = obj.IngressDevicePassword
	transport := &http.Transport{
                TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        }
        c.client = &http.Client{Transport: transport}
	return c
}

type responseHandlerFunc func(resp *http.Response) ([]byte, error)

func createResponseHandler(resp *http.Response) ([]byte, error) {
	switch resp.Status {
	case "201 Created", "200 OK":
		body, _ := ioutil.ReadAll(resp.Body)
		return body, nil
	case "409 Conflict":
		body, _ := ioutil.ReadAll(resp.Body)
		return body, errors.New("failed: " + resp.Status + " (" + string(body) + ")")

	case "207 Multi Status":
		//This happens in case of Bulk operations, which we do not support yet
		body, _ := ioutil.ReadAll(resp.Body)
		return body, nil
	case "400 Bad Request", "401 Unauthorized", "403 Forbidden",
		"404 Not Found", "405 Method Not Allowed", "406 Not Acceptable",
		"503 Service Unavailable", "599 Netscaler specific error":
		body, _ := ioutil.ReadAll(resp.Body)
		klog.Info("[INFO] go-nitro: error = " + string(body))
		return body, errors.New("failed: " + resp.Status + " (" + string(body) + ")")
	default:
		body, err := ioutil.ReadAll(resp.Body)
		return body, err

	}
}

func (c *NitroClient) createHTTPRequest(method string, url string, buff *bytes.Buffer) (*http.Request, error) {
	req, err := http.NewRequest(method, url, buff)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if c.proxiedNs == "" {
		req.Header.Set("X-NITRO-USER", c.username)
		req.Header.Set("X-NITRO-PASS", c.password)
	} else {
		req.SetBasicAuth(c.username, c.password)
		req.Header.Set("_MPS_API_PROXY_MANAGED_INSTANCE_IP", c.proxiedNs)
	}
	return req, nil
}
func (c *NitroClient) doHTTPRequest(method string, url string, bytes *bytes.Buffer, respHandler responseHandlerFunc) ([]byte, error) {
	req, err := c.createHTTPRequest(method, url, bytes)
	if err != nil {
		return []byte{}, err
	}
	resp, err := c.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return []byte{}, err
	}
	klog.Info("[DEBUG] go-nitro: response Status:", resp.Status)
	return respHandler(resp)
}

func (c *NitroClient) createResource(resourceType string, resourceJSON []byte, operation string) ([]byte, error) {
	klog.Info("[DEBUG] go-nitro: Creating resource of type ", resourceType)
	url := c.url + resourceType
	if (operation == "unset") {
		url = url + "?action=unset"	
	}else if strings.Contains(resourceType, "netprofile") {
                url = url + "?idempotent=yes"
        }
	klog.Info("[TRACE] go-nitro: url is ", url)
	return c.doHTTPRequest("POST", url, bytes.NewBuffer(resourceJSON), createResponseHandler)
}

func (c *NitroClient) AddResource(resourceType string, name string, resourceStruct interface{}, operation string) (string, error) {

	nsResource := make(map[string]interface{})
	nsResource[resourceType] = resourceStruct

	resourceJSON, err := JSONMarshal(nsResource)
	if err != nil {
		return "", fmt.Errorf("[ERROR] go-nitro: Failed to create resource of type %s, name=%s, err=%s", resourceType, name, err)
	}

	klog.Info("[TRACE] go-nitro: Resourcejson is " + string(resourceJSON))

	body, err := c.createResource(resourceType, resourceJSON, operation)
	if err != nil {
		return "", fmt.Errorf("[ERROR] go-nitro: Failed to create resource of type %s, name=%s, err=%s", resourceType, name, err)
	}
	_ = body

	return name, nil
}

type Vxlan struct {
	Dynamicrouting     string `json:"dynamicrouting,omitempty"`
	Id                 int    `json:"id,omitempty"`
	Ipv6dynamicrouting string `json:"ipv6dynamicrouting,omitempty"`
	Port               int    `json:"port,omitempty"`
	Td                 int    `json:"td,omitempty"`
	Vlan               int    `json:"vlan,omitempty"`
}
type Vxlan_srcip_binding struct {
	Id        int    `json:"id,omitempty"`
	Ipaddress string `json:"ipaddress,omitempty"`
	Srcip     string `json:"srcip",omitempty"`
	Netmask   string `json:"netmask,omitempty"`
}
type Route struct {
	Network string `json:"network,omitempty"`
	Gateway string `json:"gateway,omitempty"`
	Netmask string `json:"netmask,omitempty"`
}
type Nsip struct {
	Ipaddress string `json:"ipaddress,omitempty"`
	Netmask   string `json:"netmask,omitempty"`
}
type Netprofile struct {
	Name string `json:"name,omitempty"`
	Srcip string `json:"ipaddress,omitempty"`
}

type Arp struct {
	Ipaddress string `json:"ipaddress,omitempty"`
	Mac       string `json:"mac,omitempty"`
	Vxlan     string `json:"vxlan,omitempty"`
	Vtep      string `json:"vtep,omitempty"`
}

// Key the key of the dictionary
type Key interface{}

// Value the content of the dictionary
type Value interface{}

// ValueDictionary the set of Items
type ConfigPack struct {
	items map[Key]Value
	keys  []string
}

type SameSubnet struct {
	items map[Key]Value
}

// Set adds a new item to the dictionary
func (d *ConfigPack) Set(k Key, v Value) {
	if d.items == nil {
		d.items = make(map[Key]Value)
	}
	d.items[k] = v
	d.keys = append(d.keys, k.(string))
}

func createIngressDeviceClient(input *ControllerInput) *NitroClient {
	client := NewNitroClient(input)
	return client
}
func AddIngressDeviceConfig(config *ConfigPack, client *NitroClient) {
	for ind, _ := range config.keys {
		result, err := client.AddResource(config.keys[ind], "ADD", config.items[config.keys[ind]], "add")
		if err != nil {
			fmt.Println("Result  err ", result, err)
		}
	}
}

func BindToNetProfile(controllerInput *ControllerInput, client *NitroClient){
	var args = map[string]string{"name": controllerInput.IngressDeviceNetprof, "srcip": controllerInput.IngressDevicePodIP}
	result, err := client.AddResource("netprofile", "UPDATE", args, "set")
	fmt.Println("Result  netprofile ", result, err)
}
func UnBindNetProfile(controllerInput *ControllerInput, client *NitroClient){
	var args = map[string]string{"name": controllerInput.IngressDeviceNetprof, "srcip": "true"}
	result, err := client.AddResource("netprofile", "UNSET", args, "unset")
	fmt.Println("Result  netprofile ", result, err)
}

func DeleteIngressDeviceConfig(config *ConfigPack, client *NitroClient) {
	for ind, _ := range config.keys {
		err := client.DeleteResourceWithArgsMap(config.keys[ind], "", config.items[config.keys[ind]].(map[string]string))
		if err != nil {
			fmt.Println("Err ", err)
		}
	}
}

/*
*************************************************************************************************
*   APIName :  NsInterfaceAddRoute                                                              *
*   Input   :  k8sclient, nitro client and node					             	*
*   Output  :  Create Route and Arp.				                                *
*   Descr   :  This API initialize a node and return it.					*
*************************************************************************************************
 */
func NsInterfaceAddRoute(client *NitroClient, input *ControllerInput, node *Node) {
	configPack := ConfigPack{}
	route := Route{
		Network: node.PodAddress,
		Netmask: node.PodNetMask,
		Gateway: node.PodAddress,
	}
	configPack.Set("route", &route)

	arp := Arp{
		Ipaddress: node.PodAddress,
		Mac:       node.PodVTEP,
		Vxlan:     input.IngressDeviceVxlanIDs,
		Vtep:      node.IPAddr,
	}
	configPack.Set("arp", &arp)
	AddIngressDeviceConfig(&configPack, client)
}

func NsInterfaceDeleteRoute(client *NitroClient, obj *ControllerInput, nodeinfo *Node) {
	var argsBundle = map[string]string{"network": nodeinfo.PodAddress, "netmask": nodeinfo.PodNetMask, "gateway": nodeinfo.PodAddress}
	err2 := client.DeleteResourceWithArgsMap("route", "", argsBundle)
	if err2 != nil {
		fmt.Println(err2)
	}
	argsBundle = map[string]string{"Ipaddress": nodeinfo.PodAddress}
	err2 = client.DeleteResourceWithArgsMap("arp", "", argsBundle)
	if err2 != nil {
		fmt.Println(err2)
	}

}
func (ingressDevice *NitroClient) GetVxlanConfig(controllerInput *ControllerInput) {
	klog.Info("GetVxlanConfig")
}

//DeleteResourceWithArgsMap deletes a resource of supplied type and name. Args are supplied as map of key value
func (c *NitroClient) DeleteResourceWithArgsMap(resourceType string, resourceName string, args map[string]string) error {

	_, err := c.listResourceWithArgsMap(resourceType, resourceName, args)
	if err == nil { // resource exists
		klog.Infof("[INFO] go-nitro: DeleteResource found resource of type %s: %s", resourceType, resourceName)
		_, err = c.deleteResourceWithArgsMap(resourceType, resourceName, args)
		if err != nil {
			klog.Errorf("[ERROR] go-nitro: Failed to delete resourceType %s: %s, err=%s", resourceType, resourceName, err)
			return err
		}
	} else {
		klog.Infof("[INFO] go-nitro: Resource %s already deleted ", resourceName)
	}
	return nil
}
func (c *NitroClient) deleteResourceWithArgs(resourceType string, resourceName string, args []string) ([]byte, error) {
	klog.Info("[DEBUG] go-nitro: Deleting resource of type ", resourceType, "with args ", args)
	var url string
	if resourceName != "" {
		url = c.url + fmt.Sprintf("%s/%s?args=", resourceType, resourceName)
	} else {
		url = c.url + fmt.Sprintf("%s?args=", resourceType)
	}
	url = url + strings.Join(args, ",")
	klog.Info("[TRACE] go-nitro: url is ", url)

	return c.doHTTPRequest("DELETE", url, bytes.NewBuffer([]byte{}), deleteResponseHandler)

}

func (c *NitroClient) deleteResourceWithArgsMap(resourceType string, resourceName string, argsMap map[string]string) ([]byte, error) {
	args := make([]string, len(argsMap))
	i := 0
	for key, value := range argsMap {
		args[i] = fmt.Sprintf("%s:%s", key, value)
		i++
	}
	return c.deleteResourceWithArgs(resourceType, resourceName, args)

}
func (c *NitroClient) listResourceWithArgsMap(resourceType string, resourceName string, argsMap map[string]string) ([]byte, error) {
	args := make([]string, len(argsMap))
	i := 0
	for key, value := range argsMap {
		args[i] = fmt.Sprintf("%s:%s", key, value)
		i++
	}
	return c.listResourceWithArgs(resourceType, resourceName, args)

}
func (c *NitroClient) listResourceWithArgs(resourceType string, resourceName string, args []string) ([]byte, error) {
	klog.Info("[DEBUG] go-nitro: listing resource of type ", resourceType, ", name: ", resourceName, ", args:", args)
	var url string

	if resourceName != "" {
		url = c.url + fmt.Sprintf("%s/%s", resourceType, resourceName)
	} else {
		url = c.url + fmt.Sprintf("%s", resourceType)
	}
	strArgs := strings.Join(args, ",")
	url2 := url + "?args=" + strArgs
	klog.Info("[TRACE] go-nitro: url is ", url)

	data, err := c.doHTTPRequest("GET", url2, bytes.NewBuffer([]byte{}), readResponseHandler)
	if err != nil {
		klog.Info("[DEBUG] go-nitro: error listing with args, trying filter")
		url2 = url + "?filter=" + strArgs
		return c.doHTTPRequest("GET", url2, bytes.NewBuffer([]byte{}), readResponseHandler)
	}
	return data, err

}
func readResponseHandler(resp *http.Response) ([]byte, error) {
	switch resp.Status {
	case "200 OK":
		body, _ := ioutil.ReadAll(resp.Body)
		return body, nil
	case "404 Not Found":
		body, _ := ioutil.ReadAll(resp.Body)
		klog.Info("[DEBUG] go-nitro: read: 404 not found")
		return body, errors.New("go-nitro: read: 404 not found: ")
	case "400 Bad Request", "401 Unauthorized", "403 Forbidden",
		"405 Method Not Allowed", "406 Not Acceptable",
		"409 Conflict", "503 Service Unavailable", "599 Netscaler specific error":
		body, _ := ioutil.ReadAll(resp.Body)
		klog.Info("[INFO] go-nitro: read: error = " + string(body))
		return body, errors.New("[INFO] go-nitro: failed read: " + resp.Status + " (" + string(body) + ")")
	default:
		body, err := ioutil.ReadAll(resp.Body)
		klog.Info("[INFO] go-nitro: read error = " + string(body))
		return body, err

	}
}
func deleteResponseHandler(resp *http.Response) ([]byte, error) {
	switch resp.Status {
	case "200 OK", "404 Not Found":
		body, _ := ioutil.ReadAll(resp.Body)
		return body, nil

	case "400 Bad Request", "401 Unauthorized", "403 Forbidden",
		"405 Method Not Allowed", "406 Not Acceptable",
		"409 Conflict", "503 Service Unavailable", "599 Netscaler specific error":
		body, _ := ioutil.ReadAll(resp.Body)
		klog.Info("[INFO] go-nitro: delete: error = " + string(body))
		return body, errors.New("[INFO] delete failed: " + resp.Status + " (" + string(body) + ")")
	default:
		body, err := ioutil.ReadAll(resp.Body)
		return body, err

	}
}

func getNetScalerDefaultGateway(IngressDeviceClient *NitroClient) string{
	var data map[string]interface{}        
	url := IngressDeviceClient.url + "route"
        url = url + "?args=network:0.0.0.0,netmask:0.0.0.0"

	result, err := IngressDeviceClient.doHTTPRequest("GET", url, bytes.NewBuffer([]byte{}), readResponseHandler)
	if err != nil {
		klog.Info("[DEBUG] go-nitro: error listing with args, trying filter")
		return "error"
	}
        if err = json.Unmarshal(result, &data); err != nil {
                klog.Info("[ERROR] go-nitro: FindResourceArray: Failed to unmarshal Netscaler Response!", err, data)
		return "error"
        }
        rsrcs, ok := data["route"]
        if !ok || rsrcs == nil {
		klog.Info("[ERROR]")
		return "error"
        }
        resources := data["route"].([]interface{})
        ret := make([]map[string]interface{}, len(resources), len(resources))
        for i, v := range resources {
                ret[i] = v.(map[string]interface{})
        }
	var gateway string = ret[0]["gateway"].(string)
        log.Println("[INFO] gateway", gateway)
	return gateway
}

func getDefaultDatewayInterface(IngressDeviceClient *NitroClient, gateway string) string{
	var arpdata map[string]interface{}        
	url := IngressDeviceClient.url + "arp"
	arg := fmt.Sprintf("IPAddress:%s", gateway)
        url = url + "?filter=" + arg

	result, err := IngressDeviceClient.doHTTPRequest("GET", url, bytes.NewBuffer([]byte{}), readResponseHandler)
	if err != nil {
		log.Println("[DEBUG] go-nitro: error listing with args, trying filter")
	}
        if err = json.Unmarshal(result, &arpdata); err != nil {
                klog.Info("[ERROR] go-nitro: FindResourceArray: Failed to unmarshal Netscaler Response!", err, arpdata)
		return "error"
        }
        rsrcs, ok := arpdata["arp"]
        if !ok || rsrcs == nil {
		log.Printf("[ERROR]")
                return "error"
        }
        resources := arpdata["arp"].([]interface{})
        ret := make([]map[string]interface{}, len(resources), len(resources))
        for i, v := range resources {
                ret[i] = v.(map[string]interface{})
        }
        ifnum := ret[0]["ifnum"].(string)
        log.Println("[INFO ifnum]", ifnum)
	return ifnum
}

func getInterfaceMac(IngressDeviceClient *NitroClient, ifnum string) string{
	var ifdata map[string]interface{}        
	url := IngressDeviceClient.url + "Interface"
	ifnums := strings.Split(ifnum, "/")
        url = url + "/"+ifnums[0]+"%252F"+ifnums[1]

	result, err := IngressDeviceClient.doHTTPRequest("GET", url, bytes.NewBuffer([]byte{}), readResponseHandler)
	if err != nil {
		log.Println("[DEBUG] go-nitro: error listing with args, trying filter")
	}
        if err = json.Unmarshal(result, &ifdata); err != nil {
                klog.Info("[ERROR] go-nitro: FindResourceArray: Failed to unmarshal Netscaler Response!", err, ifdata)
		return "error"
        }
        rsrcs, ok := ifdata["Interface"]
        if !ok || rsrcs == nil {
		log.Printf("[ERROR]")
                return "error"
        }
        resources := ifdata["Interface"].([]interface{})
        ret := make([]map[string]interface{}, len(resources), len(resources))
        for i, v := range resources {
                ret[i] = v.(map[string]interface{})
        }
	mac:= ret[0]["mac"].(string)
	vmac:= ret[0]["vmac"].(string)
        log.Println("[INFO MAC]=", mac, "\t VMAC=",vmac)
	return vmac
}

func getClusterInterfaceMac(IngressDeviceClient *NitroClient) string{
	gateway := getNetScalerDefaultGateway(IngressDeviceClient)
	if (gateway == "error"){
		return "error"
	}
        ifnum  := getDefaultDatewayInterface(IngressDeviceClient, gateway)
	if (ifnum == "error"){
		return "error"
	}
	VtepMac	:= getInterfaceMac(IngressDeviceClient, ifnum)
	return VtepMac
}
func getPrimaryNodeIP(IngressDeviceClient *NitroClient) string{
	var hanode map[string]interface{}        
	url := IngressDeviceClient.url + "hanode"
	log.Println("[JANRAJ]",url)
	log.Println("[JANRAJ CJ]",url)

	result, err := IngressDeviceClient.doHTTPRequest("GET", url, bytes.NewBuffer([]byte{}), readResponseHandler)
	if err != nil {
		log.Println("[DEBUG] go-nitro: error listing with args, trying filter")
	}
        if err = json.Unmarshal(result, &hanode); err != nil {
                klog.Info("[JANRAJ ERROR] go-nitro: FindResourceArray: Failed to unmarshal Netscaler Response!", err, hanode)
		return "error"
        }
        rsrcs, ok := hanode["hanode"]
        if !ok || rsrcs == nil {
		log.Printf("[JANRAJ ERROR]")
                return "error"
        }
        resources := hanode["hanode"].([]interface{})
        ret := make([]map[string]interface{}, len(resources), len(resources))
        for i, v := range resources {
                ret[i] = v.(map[string]interface{})
		state := ret[i]["state"].(string)
		if (state == "Primary") {
			log.Println("[JANRAJ] PRIMARY", ret[i]["ipaddress"])
			
		}else if (state == "Secondary"){
			log.Println("[JANRAJ] SECONDARY", ret[i]["ipaddress"])
		}
		
        }
	return "error"
}
