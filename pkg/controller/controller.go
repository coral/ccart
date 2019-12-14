package controller

import (
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/roffe/ccart/pkg/caddycfg"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	// IngressClassAnnotationKey is the annotation key used to decide the ingress class
	IngressClassAnnotationKey = "kubernetes.io/ingress.class"
	// IngressClass is the type we will be listening for
	IngressClass = "caddy"
)

var (
	routes = map[string]caddycfg.Route{}
)

// Controller holds our caddy ingress controller
type Controller struct {
	ingress   cache.SharedInformer
	service   cache.SharedInformer
	endpoints cache.SharedInformer
	stop      chan struct{}
	sync.Mutex
}

// New returns a new caddy controller
func New(clientset *kubernetes.Clientset) *Controller {
	informerFactory := kubeinformers.NewSharedInformerFactory(clientset, time.Second*30)
	stop := make(chan struct{})
	c := &Controller{
		stop: stop,
	}
	c.ingress = NewIngressInformer(informerFactory, c)
	c.service = NewServiceInformer(informerFactory, c)
	c.endpoints = NewEndpoinsInformer(informerFactory, c)
	c.init()
	informerFactory.Start(stop)
	/*
		for {
			if c.ingress.GetController().HasSynced() &&
				c.service.GetController().HasSynced() &&
				c.endpoints.GetController().HasSynced() {
				break
			}
			time.Sleep(1 * time.Second)
		}
		glog.Info("controller in sync")
	*/
	return c
}

//Stop the controller
func (c *Controller) Stop() {
	close(c.stop)
}

func (c *Controller) init() {
	glog.Info("init controller")
	cfg := caddycfg.New()
	setConfig(cfg)
}

// Add ...
func (c *Controller) Add(obj interface{}) {
	c.Lock()
	defer c.Unlock()
	switch t := obj.(type) {
	case *v1.Service:
	case *v1.Endpoints:
	case *v1beta1.Ingress:
		annotations := t.GetAnnotations()
		if class, ok := annotations[IngressClassAnnotationKey]; ok && class == IngressClass {
			c.addIngress(t)
		}
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
		annotations := t.GetAnnotations()
		if class, ok := annotations[IngressClassAnnotationKey]; ok && class == IngressClass {
			deleteIngress(t)
		}
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
		annotations := t.GetAnnotations()
		if class, ok := annotations[IngressClassAnnotationKey]; ok && class == IngressClass {
			c.addIngress(t)
		}
	default:
		glog.Infof("%T", t)
	}
	//glog.Infof("changed: %s \n", newObj)
}
