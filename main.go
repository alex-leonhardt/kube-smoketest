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
		glog.Errorf("🔴 Component statuses: %v", err)
		errors.Errors = append(errors.Errors, err)
		LogAndExit(errors) // exit early as if components are failed
	}
	if err == nil {
		glog.Infoln("🧩 Component statuses")
	}

	// -------------------------------------------------

	err = smoketests.CreateNamespace(ctx, client)
	if err != nil {
		glog.Errorf("🔴 Create namespace: %v", err)
		errors.Errors = append(errors.Errors, err)
		LogAndExit(errors) // exit early as if there's no namespace, then we cannot run
	}
	if err == nil {
		glog.Infoln("🎮 Create namespace")
	}

	// -------------------------------------------------

	err = smoketests.PodLogs(ctx, client)
	if err != nil {
		errors.Errors = append(errors.Errors, err)
		glog.Errorf("🔴 Pod + Logs: %v", err)
	}
	if err == nil {
		glog.Infoln("🐳 Pod + 📜 Logs")
	}

	// -------------------------------------------------

	err = smoketests.CreateDeployment(ctx, client)
	if err != nil {
		errors.Errors = append(errors.Errors, err)
		glog.Errorf("🔴 Deployment: %v", err)
	}
	if err == nil {
		glog.Infoln("🚀 Deployment")
	}

	// -------------------------------------------------

	// -------------------------------------------------
	// delete the namespace when debug is set to false, which is the default

	if *debug == false {
		err = smoketests.DeleteNamespace(ctx, client)
		if err != nil {
			errors.Errors = append(errors.Errors, err)
			glog.Errorf("🔴 Delete namespace: %v", err)
		}
	}
	if err == nil {
		glog.Infoln("⦿ Delete namespace")
	}

	LogAndExit(errors)
}

// LogAndExit does just that...
func LogAndExit(errors multierror.Error) {
	if errors.ErrorOrNil() != nil {
		glog.Errorln("🔥too many errors found, expected: 0, actual:", len(errors.Errors))
	}
	os.Exit(len(errors.Errors)) // Exits > 0 if any errors occured :)
}
