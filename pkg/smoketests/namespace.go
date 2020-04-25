package smoketests

import (
	"context"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

// Namespace creates the kube-smoketest namespace
func CreateNamespace(ctx context.Context, client *kubernetes.Clientset) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	opts := metav1.CreateOptions{}
	namespace, err := client.CoreV1().Namespaces().Create(ctx, ns, opts)
	if err != nil {
		glog.Warningf("failed to create namespace %s: %v", namespace, err.Error())
		return err
	}

	glog.Infof("namespace %v created", namespace.Name)
	return nil
}

// DeleteNamespace deletes the test namespace
func DeleteNamespace(ctx context.Context, client *kubernetes.Clientset) error {
	opts := metav1.DeleteOptions{}
	err := client.CoreV1().Namespaces().Delete(ctx, namespace, opts)
	if err != nil {
		glog.Warningf("failed to delete namespace %s: %v", namespace, err.Error())
		return err
	}

	glog.Infof("namespace %v deleted", namespace)
	return nil
}
