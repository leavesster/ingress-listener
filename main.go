package main

import (
	"fmt"
	"log"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	informerFactory := informers.NewSharedInformerFactory(clientset, 0)
	ingressInformer := informerFactory.Networking().V1().Ingresses().Informer()

	ingressInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ingress := obj.(*networkingv1.Ingress)
			fmt.Printf("Ingress %s/%s created\n", ingress.Namespace, ingress.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldIngress := oldObj.(*networkingv1.Ingress)
			newIngress := newObj.(*networkingv1.Ingress)
			fmt.Printf("Ingress %s/%s updated from %+v to %+v\n", newIngress.Namespace, newIngress.Name, oldIngress.Spec, newIngress.Spec)
		},
		DeleteFunc: func(obj interface{}) {
			ingress := obj.(*networkingv1.Ingress)
			fmt.Printf("Ingress %s/%s deleted\n", ingress.Namespace, ingress.Name)
		},
	})

	informerFactory.Start(nil)
	informerFactory.WaitForCacheSync(nil)
}
