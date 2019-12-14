package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/roffe/ccart/pkg/caddycfg"
	v1 "k8s.io/api/core/v1"
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

func (c *Controller) addIngress(ingress *v1beta1.Ingress) {
	for {
		if c.endpoints.GetController().HasSynced() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	for _, rule := range ingress.Spec.Rules {
		upstreams := []caddycfg.Upstream{}
		for _, p := range rule.IngressRuleValue.HTTP.Paths {
			//glog.Infof("%s:%d", p.Backend.ServiceName, p.Backend.ServicePort.IntValue())
			key := fmt.Sprintf("%s/%s", ingress.Namespace, p.Backend.ServiceName)
			v, exists, err := c.endpoints.GetStore().GetByKey(key)
			if err != nil {
				glog.Error(err)
				return
			}
			if exists {
				endpoints, ok := v.(*v1.Endpoints)
				if !ok {
					glog.Error("typecast failed")
				}
				for _, subset := range endpoints.Subsets {
					for _, ee := range subset.Addresses {
						//glog.Info(ee.IP, subset.Ports[i].Port)
						upstreams = append(upstreams, caddycfg.Upstream{
							Dial: fmt.Sprintf("%s:%d", ee.IP, subset.Ports[0].Port),
						})
					}
				}
			} else {
				glog.Errorf("key %q does not exist", key)
				return
			}
		}

		route := caddycfg.Route{
			Handle: []caddycfg.Handle{
				caddycfg.Handle{
					Handler:   caddycfg.ReverseProxy,
					Upstreams: upstreams,
				},
			},
			Match: []caddycfg.Match{
				caddycfg.Match{
					Host: []string{rule.Host},
				},
			},
		}

		if err := srv.AddRoute(route); err != nil {
			if err == caddycfg.ErrRouteAlreadyExists {
				glog.V(2).Info("route already up to date")
				return
			}
			glog.Error(err)
			return
		}
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
	jsonStr, err := srv.ParseJSON()
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
