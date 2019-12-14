package controller

import (
	"sync"

	"github.com/golang/glog"
	"github.com/roffe/ccart/pkg/caddycfg"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
)

var (
	routes = map[string]caddycfg.Route{}
)

// Controller holds our caddy ingress controller
type Controller struct {
	services  []*v1.Service
	endpoints []*v1.Endpoints
	ingresses []*v1beta1.Ingress
	sync.Mutex
}

// New returns a new caddy controller
func New() *Controller {
	c := &Controller{}
	go c.run()
	return c
}

func (c *Controller) run() {
	glog.Info("new controller")
	cfg := caddycfg.New()
	setConfig(cfg)
}

// Add ...
func (c *Controller) Add(obj interface{}) {
	c.Lock()
	defer c.Unlock()
	switch t := obj.(type) {
	case *v1.Service:
		c.services = append(c.services, t)
	case *v1.Endpoints:
		c.endpoints = append(c.endpoints, t)
	case *v1beta1.Ingress:
		c.ingresses = append(c.ingresses, t)
		addIngress(t)
	default:
		glog.Infof("%T", t)
	}
	//glog.Infof("added: %s \n", obj)
}

// Delete ...
func (c *Controller) Delete(obj interface{}) {
	switch t := obj.(type) {
	case *v1.Service:
	case *v1.Endpoints:
	case *v1beta1.Ingress:
		deleteIngress(t)
	default:
		glog.Infof("%T", t)
	}
	//glog.Infof("deleted: %s \n", obj)
}

// Update ...
func (c *Controller) Update(oldObj, newObj interface{}) {
	switch t := newObj.(type) {
	case *v1.Service:
	case *v1.Endpoints:
	case *v1beta1.Ingress:
	default:
		glog.Infof("%T", t)
	}
	//glog.Infof("changed: %s \n", newObj)
}
