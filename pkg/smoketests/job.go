// Package smoketests ...
package smoketests

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/google/uuid"

	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

// CreateJob ... creates a k8s job, runs the command and exits
func CreateJob(ctx context.Context, client *kubernetes.Clientset, arg string) (*v1.Job, error) {

	uuid, err := uuid.NewUUID()
	uuids := strings.Split(fmt.Sprintf("%s", uuid), "-")
	jobName := strings.Join(uuids, "")[len(uuids)/2:]

	glog.V(2).Infof("creating test job %s, running: %s %q", jobName, "/bin/sh -c", arg)

	jobSpec := &v1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Batch",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: v1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tester",
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						corev1.Container{
							Name:    "box",
							Image:   "busybox",
							Command: []string{"/bin/sh", "-c"},
							Args:    []string{arg},
						},
					},
				},
			},
		},
	}

	job, err := client.BatchV1().Jobs(namespace).Create(ctx, jobSpec, metav1.CreateOptions{})
	if err != nil {
		glog.Errorf("failed to create job %s: %v", jobName, err)
		return nil, err
	}

	glog.V(2).Infof("successfully created job %s", jobName)

	return job, nil
}
