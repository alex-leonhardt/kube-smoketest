package smoketests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/jpillora/backoff"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreatePod creates a pod, seems obvious :)
//
// testName is mandatory;
// testImage (default: alpine),
// command (default: /bin/sh),
// args (default: while true; do echo `date`; sleep 1; done)
func CreatePod(ctx context.Context, client *kubernetes.Clientset, testName string, testImage string, command, args []string) (*v1.Pod, error) {

	if testName == "" {
		return nil, fmt.Errorf("failed to create pod: must specify a testName when creating a pod")
	}

	if testImage == "" {
		testImage = "alpine"
	}

	if len(command) < 1 && len(args) < 1 {
		command = []string{"/bin/sh"}
		args = []string{"-c", "while true; do echo `date`; sleep 1; done"}
	}

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ToLower(testName),
			Namespace: namespace,
			Labels: map[string]string{
				"testName": testName,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "yolo",
					Image:           testImage,
					ImagePullPolicy: v1.PullAlways,
					Command:         command,
					Args:            args,
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
	opts := metav1.CreateOptions{}

	pod, err := client.CoreV1().Pods(namespace).Create(ctx, pod, opts)
	if err != nil {
		glog.Warning(err.Error())
		return nil, err
	}
	glog.Infof("pod %s created", pod.GetName())

	return pod, nil
}

// PodLogs retrievs a pod's last 10 log lines and logs them to stdout, it returns with non-nil if any error was found
func PodLogs(ctx context.Context, client *kubernetes.Clientset) error {

	pod, err := CreatePod(ctx, client, "PodLogs", "", nil, nil)
	if err != nil {
		glog.Errorf("failed to create pod for test: %v", err)
		return err
	}

	if err = WaitForPod(ctx, client, pod); err != nil {
		glog.Errorf("failed waiting for pod to become ready: %v", err.Error())
		return err
	}

	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		glog.Errorf("failed to get pod list: %v", err)
		return err
	}

	if len(podList.Items) < 1 {
		glog.Warningf("no pods found: %v", len(podList.Items))
		return err
	}

	pod = &podList.DeepCopy().Items[0]

	logLines := int64(10)
	logOptions := &v1.PodLogOptions{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		TailLines: &logLines,
	}

	rc := client.RESTClient()
	req := rc.Get().
		Prefix("/api/v1/"). // TODO: find out why this is necessary to make this request work ?
		Resource("pods").
		Namespace(namespace).
		Name(pod.Name).
		SubResource("log").
		Param("tailLines", strconv.FormatInt(*logOptions.TailLines, 10))

	if glog.V(10) {
		// debug output
		r := *req.URL()
		glog.Infof("Request: %s://%s%s", r.Scheme, r.Host, r.Path)
	}

	readCloser, err := req.Stream(ctx)
	if err != nil {
		glog.Warningln(err.Error())
		return err
	}
	defer readCloser.Close()

	out := bytes.NewBuffer(nil)
	_, err = io.Copy(out, readCloser)
	if err != nil {
		glog.Warningln(err.Error())
		return err
	}

	glog.Infoln(out)
	return nil
}

// WaitForPod waits for a Pod to be in a Running state
func WaitForPod(ctx context.Context, client *kubernetes.Clientset, pod *v1.Pod) error {

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
		opts := metav1.GetOptions{}
		tmpPod, err := client.CoreV1().Pods(namespace).Get(ctx, pod.Name, opts)
		if err != nil {
			continue
		}
		if tmpPod.Status.Phase == v1.PodRunning {
			return nil
		}
		glog.V(2).Infof("waiting for pod to be ready: %v", time.Since(t))
	}
}
