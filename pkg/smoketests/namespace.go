package smoketests

import (
	"context"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

// CreateNamespace creates the kube-smoketest namespace
func CreateNamespace(ctx context.Context, client *kubernetes.Clientset) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	// get our namespace and if it already exists, then we'll return early
	nsget, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		glog.V(2).Infof("namespace %s exists, not creating; found: %#v", namespace, nsget)
		return nil
	}

	opts := metav1.CreateOptions{}
	ns, err = client.CoreV1().Namespaces().Create(ctx, ns, opts)
	if err != nil {
		glog.Errorf("failed to create namespace %s: %v", namespace, err.Error())
		return err
	}

	if err = WaitFor(ctx, client, Namespace); err != nil {
		glog.Errorf("failed to create namespace %s: %v", namespace, err.Error())
		return err
	}

	glog.V(2).Infof("namespace %v created", ns.Name)
	return nil
}

// DeleteNamespace deletes the test namespace
func DeleteNamespace(ctx context.Context, client *kubernetes.Clientset) error {
	opts := metav1.DeleteOptions{}
	err := client.CoreV1().Namespaces().Delete(ctx, namespace, opts)
	if err != nil {
		glog.Errorf("failed to delete namespace %s: %v", namespace, err.Error())
		return err
	}

	glog.V(2).Infof("namespace %v deleted", namespace)
	return nil
}
