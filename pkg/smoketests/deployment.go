// create a deployment, confirms that controller-manager and scheduler work
package smoketests

import (
	"context"
	"time"

	"github.com/golang/glog"
	"github.com/jpillora/backoff"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateDeployment creates a dummy nginx deployment of 2 pods
func CreateDeployment(ctx context.Context, client *kubernetes.Clientset) error {

	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, "smoketest", metav1.GetOptions{})
	if err == nil {
		glog.Infof("using existing deployment: %#v", deploy.ObjectMeta.Name)
		return nil
	}
	numReplicas := int32(2)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "smoketest",
			Labels: map[string]string{
				"testName": "deployment",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &numReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "smoketest",
				},
			},
			MinReadySeconds: int32(7),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nginx",
					Labels: map[string]string{
						"app": "smoketest",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Image: "nginx",
							Name:  "webserver",
						},
					},
				},
			},
		},
	}
	deploy, err = client.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		glog.Errorf("failed to create deployment: %v", err)
		return err
	}

	if err = WaitForDeployment(ctx, client, numReplicas); err != nil {
		glog.Warningf("failed to create deployment: %v", err)
		return err
	}

	glog.Infof("created deployment %s", deploy.ObjectMeta.Name)
	return nil
}

// DeleteDeployment deletes the deployment ..
func DeleteDeployment(ctx context.Context, client *kubernetes.Clientset) error {
	if err := client.AppsV1().Deployments(namespace).Delete(ctx, "smoketest", metav1.DeleteOptions{}); err != nil {
		glog.Errorf("failed to delete deployment: %v", err)
		return err
	}
	return nil
}

// WaitForDeployment waits until the deployment is set to Ready status
func WaitForDeployment(ctx context.Context, client *kubernetes.Clientset, numReady int32) error {

	bo := backoff.Backoff{
		Min:    time.Second,
		Max:    5 * time.Second,
		Jitter: true,
		// Factor: float64(0.3),
	}

	t := time.Now()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(bo.Duration()):
			// continue below
		}
		deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, "smoketest", metav1.GetOptions{})
		if err != nil {
			continue
		}
		if deployment != nil {
			if deployment.Status.AvailableReplicas == numReady {
				return nil
			}
		}
		glog.V(2).Infof("waiting for pods to become available: %v", time.Since(t))
	}
}
