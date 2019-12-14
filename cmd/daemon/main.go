package main

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/roffe/ccart/pkg/controller"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func init() {
	flag.Set("logtostderr", "true")
}

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		glog.Errorln(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorln(err)
	}

	cc := controller.New(clientset)
	defer cc.Stop()

	/*
		ingressInformer := informers.NewIngressInformer(informerFactory, cc)
		serviceInformer := informers.NewServiceInformer(informerFactory, cc)
		endpointsInformer := informers.NewEndpoinsInformer(informerFactory, cc)
	*/
	/*
		cIng := ingressInformer.GetStore()
		cSvc := serviceInformer.GetStore()
		cEnd := endpointsInformer.GetStore()
	*/

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, syscall.SIGINT)

	//t := time.NewTicker(10 * time.Second)

outer:
	for {
		select {
		case s := <-sigChan:
			glog.Infof("got %s", s)
			break outer
			/*
				case <-t.C:
					glog.Info(cIng.ListKeys())
					glog.Info(cSvc.ListKeys())
					glog.Info(cEnd.ListKeys())
			*/
		}
	}
	glog.Info("sleep")
	time.Sleep(1 * time.Second)
}
