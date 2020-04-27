package smoketests

import (
	"context"

	"github.com/golang/glog"
	"github.com/hashicorp/go-multierror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ComponentStatus checks for control plane components
func ComponentStatus(ctx context.Context, client *kubernetes.Clientset) error {
	multierr := multierror.Error{}

	statuses, err := client.CoreV1().ComponentStatuses().List(ctx, metav1.ListOptions{})
	if err != nil {
		multierr.Errors = append(multierr.Errors, err)
	}
	for _, status := range statuses.Items {
		for _, cond := range status.Conditions {
			glog.V(2).Infof("component=%s health=%s msg=%s error=%q", status.ObjectMeta.Name, cond.Status, cond.Message, cond.Error)
		}
	}

	return multierr.ErrorOrNil()
}
