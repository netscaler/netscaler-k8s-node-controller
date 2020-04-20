package main

import (
	"k8s.io/klog"
	"runtime"
	"testing"
	//"github.com/stretchr/testify/assert"
	//"net/http/httptest"
	//"net/http/httptest"
	"context"
        "encoding/json" 
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	//"crypto/tls"
	"net"
)

func TestCreateIngressDeviceClient(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	klog.Info("Current test filename: " + filename)
	ControllerInput := FetchCitrixNodeControllerInput()
	client := createIngressDeviceClient(ControllerInput)
	if client == nil {
		t.Error("Expected a Valid Client ")

	}
}

func TestCreateResponseHandler(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	req, _ := http.NewRequest(http.MethodGet, testServer.URL, nil)
	res, _ := http.DefaultClient.Do(req)

	createResponseHandler(res)
	readResponseHandler(res)
	deleteResponseHandler(res)

	testServer = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(400)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	req, _ = http.NewRequest(http.MethodGet, testServer.URL, nil)
	res, _ = http.DefaultClient.Do(req)
	createResponseHandler(res)
	readResponseHandler(res)
	deleteResponseHandler(res)

	testServer = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(404)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	req, _ = http.NewRequest(http.MethodGet, testServer.URL, nil)
	res, _ = http.DefaultClient.Do(req)
	createResponseHandler(res)
	readResponseHandler(res)
	deleteResponseHandler(res)

	testServer = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(409)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	req, _ = http.NewRequest(http.MethodGet, testServer.URL, nil)
	res, _ = http.DefaultClient.Do(req)
	createResponseHandler(res)
	readResponseHandler(res)
	deleteResponseHandler(res)

	testServer = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(207)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	req, _ = http.NewRequest(http.MethodGet, testServer.URL, nil)
	res, _ = http.DefaultClient.Do(req)
	createResponseHandler(res)
	readResponseHandler(res)
	deleteResponseHandler(res)

	testServer = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(707)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	req, _ = http.NewRequest(http.MethodGet, testServer.URL, nil)
	res, _ = http.DefaultClient.Do(req)
	createResponseHandler(res)
	readResponseHandler(res)
	deleteResponseHandler(res)

	testServer = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(207)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	req, _ = http.NewRequest(http.MethodGet, testServer.URL, nil)
	res, _ = http.DefaultClient.Do(req)
	createResponseHandler(res)
	readResponseHandler(res)
	deleteResponseHandler(res)
}

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
		},
	}

	return cli, s.Close
}
func TestGetPrimaryNodeIP(t *testing.T) {
        data := make(map[string]string)
        data["hanode"] = "1.1.1.1"
        newdata, err := json.Marshal(data)
        if err != nil {
            panic(err)
        }
	_, nitro, _ := getClientAndDeviceInfo()
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "key", r.Header.Get("Key"))
		assert.Equal(t, "secret", r.Header.Get("Secret"))
		w.WriteHeader(200)
		w.Write([]byte(newdata))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	nitro.client = httpClient

	getPrimaryNodeIP(nitro)
}
func TestGetClusterInterfaceMac(t *testing.T) {
	_, nitro, _ := getClientAndDeviceInfo()
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "key", r.Header.Get("Key"))
		assert.Equal(t, "secret", r.Header.Get("Secret"))
		w.WriteHeader(200)
		w.Write([]byte("body"))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	nitro.client = httpClient
	getClusterInterfaceMac(nitro)
}
func TestGetInterfaceMac(t *testing.T) {
	_, nitro, _ := getClientAndDeviceInfo()
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "key", r.Header.Get("Key"))
		assert.Equal(t, "secret", r.Header.Get("Secret"))
		w.WriteHeader(200)
		w.Write([]byte("okResponse"))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	nitro.client = httpClient
	getInterfaceMac(nitro, "1/1")
}
func TestGetDefaultDatewayInterface(t *testing.T) {
	_, nitro, _ := getClientAndDeviceInfo()
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "key", r.Header.Get("Key"))
		assert.Equal(t, "secret", r.Header.Get("Secret"))
		w.WriteHeader(200)
		w.Write([]byte("okResponse"))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	nitro.client = httpClient
	getDefaultDatewayInterface(nitro, "0.0.0.0")
}
func TestNsInterfaceConfig(t *testing.T) {
	input, nitro, _ := getClientAndDeviceInfo()
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "key", r.Header.Get("Key"))
		assert.Equal(t, "secret", r.Header.Get("Secret"))
		w.WriteHeader(200)
		w.Write([]byte("okResponse"))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	nitro.client = httpClient
	node := new(Node)
	node.PodAddress = "2.2.2.2"
	node.PodNetMask = "255.255.255.0"
	node.PodAddress = "2.2.2.2"
	node.PodVTEP = "3.3.3.3"

	NsInterfaceAddRoute(nitro, input, node)

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

	AddIngressDeviceConfig(&configPack, nitro)
	BindToNetProfile(input, nitro)
	UnBindNetProfile(input, nitro)

	configPack = ConfigPack{}
	vxlanargs := map[string]string{"id": "1"}
	configPack.Set("vxlan", vxlanargs)

	nsipargs := map[string]string{"ipaddress": "2.2.2.2"}
	configPack.Set("nsip", nsipargs)

	DeleteIngressDeviceConfig(&configPack, nitro)
	NsInterfaceDeleteRoute(nitro, node)
	nitro.deleteResourceWithArgsMap("vxlan", "vxlan", configPack.items[configPack.keys[0]].(map[string]string))
}
