package main

import (
	"context"
	"flag"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/alex-leonhardt/kube-smoketest/pkg/smoketests"
	"github.com/golang/glog"
	"github.com/hashicorp/go-multierror"
)

func main() {

	flag.Parse()

	kubeconfigpath := os.Getenv("KUBECONFIG")

	if kubeconfigpath == "" {
		homeDir, _ := os.UserHomeDir()
		kubeconfigpath = homeDir + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigpath)
	if err != nil {
		glog.Fatalln(err.Error())
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalln(err.Error())
	}

	// ------------------------
	errors := multierror.Error{} // collect all errors here...
	// ------------------------

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err = smoketests.PodLogs(ctx, client)
	if err != nil {
		errors.Errors = append(errors.Errors, err)
		glog.Warningln("error: smoketess.PodLogs")
	}

	if errors.ErrorOrNil() != nil {
		for _, e := range errors.Errors {
			glog.Errorln(e)
		}
		glog.Errorln("fatal: too many errors found, expected: 0, actual:", len(errors.Errors))
	}

	os.Exit(len(errors.Errors)) // Exits > 0 if any errors occured :)
}
