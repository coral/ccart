package controller

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/roffe/ccart/pkg/caddycfg"
	"k8s.io/api/networking/v1beta1"
)

var (
	caddyURL = "http://localhost:2019/config/"
	srv      = caddycfg.Server{
		AutomaticHTTPS: caddycfg.AutomaticHTTPS{
			Disable: true,
		},
		Listen: []string{":80"},
		Routes: []caddycfg.Route{},
	}
)

func addIngress(ingress *v1beta1.Ingress) {

	for _, rule := range ingress.Spec.Rules {

		for _, p := range rule.IngressRuleValue.HTTP.Paths {
			glog.Infof("%s:%d", p.Backend.ServiceName, p.Backend.ServicePort.IntValue())
		}
		route := caddycfg.Route{
			Handle: []caddycfg.Handle{
				caddycfg.Handle{
					Handler: caddycfg.ReverseProxy,
					Upstreams: []caddycfg.Upstream{
						caddycfg.Upstream{
							Dial: "localhost:8080",
						},
					},
				},
			},
			Match: []caddycfg.Match{
				caddycfg.Match{
					Host: []string{rule.Host},
				},
			},
		}
		srv.AddRoute(route)
	}

	updateServer("kubernetes-ingress")
}

func deleteIngress(ingress *v1beta1.Ingress) {
	for _, rule := range ingress.Spec.Rules {
		route := caddycfg.Route{
			Handle: []caddycfg.Handle{
				caddycfg.Handle{
					Handler: caddycfg.ReverseProxy,
					Upstreams: []caddycfg.Upstream{
						caddycfg.Upstream{
							Dial: "localhost:8080",
						},
					},
				},
			},
			Match: []caddycfg.Match{
				caddycfg.Match{
					Host: []string{rule.Host},
				},
			},
		}
		srv.DeleteRoute(route)
	}

	updateServer("kubernetes-ingress")
}

func updateServer(name string) {
	jsonStr, err := json.Marshal(srv)
	if err != nil {
		panic(err)
	}
	reqURL := caddyURL + "apps/http/servers/" + name
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	glog.Infoln("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	if resp.Status != "200 OK" {
		body, _ := ioutil.ReadAll(resp.Body)
		glog.Infoln("response Body:", string(body))

	}
}

func setConfig(cfg caddycfg.Config) {
	jsonStr, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	glog.Info(string(jsonStr))

	req, err := http.NewRequest("POST", caddyURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	glog.Infoln("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	if resp.Status != "200 OK" {
		body, _ := ioutil.ReadAll(resp.Body)
		glog.Infoln("response Body:", string(body))

	}
}
