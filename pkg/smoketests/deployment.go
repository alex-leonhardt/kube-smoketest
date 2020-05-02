// Package smoketests ... create a deployment, confirms that controller-manager and scheduler work
package smoketests

import (
	"context"

	"github.com/golang/glog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateDeployment creates a dummy nginx deployment of 2 pods
func CreateDeployment(ctx context.Context, client *kubernetes.Clientset) error {
	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, "smoketest", metav1.GetOptions{})
	if err == nil {
		glog.V(2).Infof("using existing deployment: %#v", deploy.ObjectMeta.Name)
		return nil
	}

	glog.V(2).Infoln("creating deployment")

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

	if err = WaitFor(ctx, client, Deployment, WithNumReady(numReplicas)); err != nil {
		glog.Warningf("failed to create deployment: %v", err)
		return err
	}

	glog.V(2).Infof("successfully created deployment %s", deploy.ObjectMeta.Name)
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
