package main

import (
	"os"

	"github.com/kr/pretty"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	kubeconfigpath := os.Getenv("KUBECONFIG")

	if kubeconfigpath == "" {
		kubeconfigpath = "~/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigpath)
	if err != nil {
		panic(err.Error())
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pretty.Println(config)
	pretty.Println(client)

}
