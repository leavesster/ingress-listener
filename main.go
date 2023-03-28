package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

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
			fmt.Printf("add ingress: %s", ingressDescription(ingress))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldIngress := oldObj.(*networkingv1.Ingress)
			oldAddress := getAddress(oldIngress)
			newIngress := newObj.(*networkingv1.Ingress)
			newAddress := getAddress(newIngress)

			if oldAddress != newAddress {
				fmt.Printf("Ingress %s/%s updated from %s to %s", newIngress.Namespace, newIngress.Name, oldAddress, newAddress)
				updateDns(getHosts(newIngress), newAddress)
			}
		},
		DeleteFunc: func(obj interface{}) {
			ingress := obj.(*networkingv1.Ingress)
			fmt.Printf("Ingress %s/%s deleted\n", ingress.Namespace, ingress.Name)
		},
	})

	informerFactory.Start(nil)
	informerFactory.WaitForCacheSync(nil)
}

func getHosts(ingress *networkingv1.Ingress) []string {
	hosts := []string{}
	tls := ingress.Spec.TLS
	if len(tls) > 0 {
		for _, tls := range tls {
			hosts = append(hosts, tls.Hosts...)
		}
		return hosts
	}
	for _, rule := range ingress.Spec.Rules {
		hosts = append(hosts, rule.Host)
	}

	return hosts
}

func getAddress(ingress *networkingv1.Ingress) string {
	if ingress.Status.LoadBalancer.Ingress != nil {
		for _, lbIngress := range ingress.Status.LoadBalancer.Ingress {
			if lbIngress.Hostname != "" {
				return lbIngress.Hostname
			}
			if lbIngress.IP != "" {
				return lbIngress.IP
			}
		}
	}

	return ""
}

func ingressDescription(ingress *networkingv1.Ingress) string {
	json, err := json.Marshal(ingress)
	if err != nil {
		return ""
	}

	return string(json)
}

func updateDns(hosts []string, address string) {
	// if address is IP, update DNS A record
	if net.ParseIP(address) != nil {
		for _, host := range hosts {
			// update DNS A record
			fmt.Printf("update DNS A record for %s to %s", host, address)
		}
		return
	}

	// if address is hostname, update DNS CNAME record
	if _, err := net.LookupHost(address); err != nil {
		for _, host := range hosts {
			// update DNS CNAME record
			fmt.Printf("update DNS CNAME record for %s to %s", host, address)
		}
		return
	}
}
