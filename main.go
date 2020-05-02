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
	debug := flag.Bool("debug", false, "with debug set, the namespace is not cleaned up at the end of the test, before re-running the test, you must manually delete the NS and wait for it to be gone before re-running kube-smoketest")
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

	// -------------------------------------------------

	err = smoketests.ComponentStatus(ctx, client)
	if err != nil {
		glog.Errorf("\t🔴 Component statuses: %v", err)
		errors.Errors = append(errors.Errors, err)
		LogAndExit(errors) // exit early as if components are failed
	}
	if err == nil {
		glog.Infoln("\t✅ Component statuses")
	}

	// -------------------------------------------------

	err = smoketests.CreateNamespace(ctx, client)
	if err != nil {
		glog.Errorf("\t🔴 Create namespace: %v", err)
		errors.Errors = append(errors.Errors, err)
		LogAndExit(errors) // exit early as if there's no namespace, then we cannot run
	}
	if err == nil {
		glog.Infoln("\t✅ Create namespace")
	}

	// -------------------------------------------------

	err = smoketests.PodLogs(ctx, client)
	if err != nil {
		errors.Errors = append(errors.Errors, err)
		glog.Errorf("\t🔴 Pod + Logs: %v", err)
	}
	if err == nil {
		glog.Infoln("\t✅ Pod + Logs")
	}

	// -------------------------------------------------

	err = smoketests.CreateDeployment(ctx, client)
	if err != nil {
		errors.Errors = append(errors.Errors, err)
		glog.Errorf("\t🔴 Deployment: %v", err)
	}
	if err == nil {
		glog.Infoln("\t✅ Deployment")
	}

	// -------------------------------------------------

	err = smoketests.CreateService(ctx, client)
	if err != nil {
		errors.Errors = append(errors.Errors, err)
		glog.Errorf("\t🔴 Service: %v", err)
	}
	if err == nil {
		glog.Infoln("\t✅ Service")
	}

	// -------------------------------------------------

	err = smoketests.CreateNodePortService(ctx, client)
	if err != nil {
		errors.Errors = append(errors.Errors, err)
		glog.Errorf("\t🔴 NodePort Service: %v", err)
	}
	if err == nil {
		glog.Infoln("\t✅ NodePort Service")
	}

	// -------------------------------------------------

	err = smoketests.CreateSecret(ctx, client)
	if err != nil {
		errors.Errors = append(errors.Errors, err)
		glog.Errorf("\t🔴 Secret: %v", err)
	}
	if err == nil {
		glog.Infoln("\t✅ Secret")
	}

	// -------------------------------------------------

	// don't delete the namespace when debug is set to true
	if *debug != false {
		glog.Infoln("\t⚠️  Namespace remains for debugging")
		LogAndExit(errors)
	}

	err = smoketests.DeleteNamespace(ctx, client)
	if err != nil {
		errors.Errors = append(errors.Errors, err)
		glog.Errorf("\t🔴 Delete namespace: %v", err)
	}
	if err == nil {
		glog.Infoln("\t✅ Delete namespace")
	}

	LogAndExit(errors)
}

// LogAndExit does just that...
func LogAndExit(errors multierror.Error) {

	if errors.ErrorOrNil() != nil {
		glog.Errorln("\t🔴 FAILED: too many errors found, expected: 0, actual:", len(errors.Errors))
	} else {
		glog.Infoln("\t✅ SUCCESS: all tests passed")
	}
	os.Exit(len(errors.Errors)) // Exits > 0 if any errors occured :)
}
