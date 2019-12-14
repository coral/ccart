package informers

import (
	"github.com/roffe/ccart/pkg/controller"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// NewServiceInformer returns a Service informer
func NewServiceInformer(kif informers.SharedInformerFactory, cc *controller.Controller) cache.SharedInformer {
	svcInformer := kif.Core().V1().Services().Informer()
	svcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cc.Add,
		DeleteFunc: cc.Delete,
		UpdateFunc: cc.Update,
	})
	return svcInformer
}

// NewIngressInformer returns a Ingress informer
func NewIngressInformer(kif informers.SharedInformerFactory, cc *controller.Controller) cache.SharedInformer {
	ingInformer := kif.Networking().V1beta1().Ingresses().Informer()
	ingInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cc.Add,
		DeleteFunc: cc.Delete,
		UpdateFunc: cc.Update,
	})
	return ingInformer
}

// NewEndpoinsInformer returns a Endpoints informer
func NewEndpoinsInformer(kif informers.SharedInformerFactory, cc *controller.Controller) cache.SharedInformer {
	endpointsInformer := kif.Core().V1().Endpoints().Informer()
	endpointsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cc.Add,
		DeleteFunc: cc.Delete,
		UpdateFunc: cc.Update,
	})
	return endpointsInformer
}
