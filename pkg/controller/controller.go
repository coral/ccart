package controller

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/roffe/ccart/pkg/caddycfg"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	routes = map[string][]caddycfg.Route{}
	srv    = caddycfg.Server{
		AutomaticHTTPS: caddycfg.AutomaticHTTPS{
			Disable: true,
		},
		Listen: []string{":80"},
		Routes: []caddycfg.Route{},
	}
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
	return c
}

//Stop the controller
func (c *Controller) Stop() {
	close(c.stop)
}

func (c *Controller) init() {
	glog.Info("init controller")
	cfg := caddycfg.New()
	setInitialConfig(cfg)
}

// Add ...
func (c *Controller) Add(obj interface{}) {
	c.Lock()
	defer c.Unlock()
	switch t := obj.(type) {
	case *v1.Service:
	case *v1.Endpoints:
		glog.V(2).Info("%+v", t)
	case *v1beta1.Ingress:
		annotations := t.GetAnnotations()
		if class, ok := annotations[IngressClassAnnotationKey]; ok && class == IngressClass {
			c.addIngress(t)
		}
	default:
		glog.Infof("%T", t)
	}
}

// Delete ...
func (c *Controller) Delete(obj interface{}) {
	switch t := obj.(type) {
	case *v1.Service:
	case *v1.Endpoints:
		glog.V(2).Info("%+v", t)
	case *v1beta1.Ingress:
		annotations := t.GetAnnotations()
		if class, ok := annotations[IngressClassAnnotationKey]; ok && class == IngressClass {
			c.deleteIngress(t)
		}
	default:
		glog.Infof("%T", t)
	}
	glog.Infof("deleted: %s \n", obj)
}

// Update ...
func (c *Controller) Update(oldObj, newObj interface{}) {
	switch t := newObj.(type) {
	case *v1.Service:
	case *v1.Endpoints:
		glog.V(2).Info("%+v", t)
	case *v1beta1.Ingress:
		annotations := t.GetAnnotations()
		if class, ok := annotations[IngressClassAnnotationKey]; ok && class == IngressClass {
			c.addIngress(t)
		}
	default:
		glog.Infof("%T", t)
	}
	glog.Infof("changed: %s \n", newObj)
}

func (c *Controller) addIngress(ingress *v1beta1.Ingress) {
	for {
		if c.service.GetController().HasSynced() && c.endpoints.GetController().HasSynced() {
			break
		}
		glog.V(2).Info("not synched, retry")
		time.Sleep(100 * time.Millisecond)
	}
	for _, ingressRule := range ingress.Spec.Rules {
		for _, p := range ingressRule.HTTP.Paths {
			endpointKey := fmt.Sprintf("%s/%s", ingress.Namespace, p.Backend.ServiceName)

			svc, err := c.getService(endpointKey)
			if err != nil {
				panic(err)
			}

			var targetPort intstr.IntOrString

			for _, port := range svc.Spec.Ports {
				if port.Port == p.Backend.ServicePort.IntVal {
					targetPort = port.TargetPort
					break
				}
			}

			upstreams, err := c.getEndpoints(endpointKey, targetPort)
			if err != nil {
				panic(err)
			}
			r := caddycfg.Route{}
			r.Handle = []caddycfg.Handle{
				caddycfg.Handle{
					Handler:   caddycfg.ReverseProxy,
					Upstreams: upstreams,
				},
			}
			match := caddycfg.Match{
				Host: []string{ingressRule.Host},
			}
			if p.Path != "" {
				match.Path = []string{p.Path}
			}
			r.Match = []caddycfg.Match{match}

			if err := srv.AddRoute(r); err != nil {
				if err == caddycfg.ErrRouteAlreadyExists {
					glog.V(2).Info("route already up to date")
					return
				}
				glog.Error(err)
				return
			}
		}
	}
	updateServer("kubernetes-ingress")
}

func (c *Controller) getService(key string) (*v1.Service, error) {
	v, exists, err := c.service.GetStore().GetByKey(key)
	if err != nil {
		return nil, fmt.Errorf("error getting service: %s", err)
	}
	if !exists {
		return nil, fmt.Errorf("service %q does not exist", key)
	}
	svc, ok := v.(*v1.Service)
	if !ok {
		return nil, errors.New("failed to typecast as service")
	}
	return svc, nil
}

func (c *Controller) getEndpoints(key string, targetPort intstr.IntOrString) ([]caddycfg.Upstream, error) {
	v, exists, err := c.endpoints.GetStore().GetByKey(key)
	if err != nil {
		return []caddycfg.Upstream{}, fmt.Errorf("error getting endpoint: %s", err)
	}
	if !exists {
		return []caddycfg.Upstream{}, fmt.Errorf("key %q does not exist", key)
	}

	var upstreams []caddycfg.Upstream

	endpoints, ok := v.(*v1.Endpoints)
	if !ok {
		glog.Error("typecast of endpoint failed")
	}

	for _, subset := range endpoints.Subsets {
		for _, addr := range subset.Addresses {
			upstreams = append(upstreams, caddycfg.Upstream{
				Dial: fmt.Sprintf("%s:%d", addr.IP, targetPort.IntValue()),
			})
		}
	}
	return upstreams, nil
}

func (c *Controller) deleteIngress(ingress *v1beta1.Ingress) {
	for _, ingressRule := range ingress.Spec.Rules {
		for _, p := range ingressRule.HTTP.Paths {
			endpointKey := fmt.Sprintf("%s/%s", ingress.Namespace, p.Backend.ServiceName)

			svc, err := c.getService(endpointKey)
			if err != nil {
				panic(err)
			}

			var targetPort intstr.IntOrString

			for _, port := range svc.Spec.Ports {
				if port.Port == p.Backend.ServicePort.IntVal {
					targetPort = port.TargetPort
					break
				}
			}

			upstreams, err := c.getEndpoints(endpointKey, targetPort)
			if err != nil {
				panic(err)
			}
			r := caddycfg.Route{}
			r.Handle = []caddycfg.Handle{
				caddycfg.Handle{
					Handler:   caddycfg.ReverseProxy,
					Upstreams: upstreams,
				},
			}
			match := caddycfg.Match{
				Host: []string{ingressRule.Host},
			}
			if p.Path != "" {
				match.Path = []string{p.Path}
			}
			r.Match = []caddycfg.Match{match}
			srv.DeleteRoute(r)
		}
	}

	updateServer("kubernetes-ingress")
}
