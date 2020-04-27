package smoketests

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// CreateService creates a ClusterIP service for the Deployment smoketest
func CreateService(ctx context.Context, client *kubernetes.Clientset) error {
	return fmt.Errorf("not implemented")
}
