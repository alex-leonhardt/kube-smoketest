package smoketests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodLogs retrievs a pod's last 10 log lines and logs them to stdout, it returns with non-nil if any error was found
func PodLogs(ctx context.Context, client *kubernetes.Clientset) error {

	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod list: %v", err)
	}

	if len(podList.Items) < 1 {
		return fmt.Errorf("no pods found: %v", len(podList.Items))
	}

	pod := podList.DeepCopy().Items[0]

	logOptions := &v1.PodLogOptions{}

	req := client.RESTClient().Get().
		Namespace(namespace).
		Name(pod.Name).
		Resource("pods").
		SubResource("log").
		Param("tailLines", strconv.FormatInt(*logOptions.TailLines, 10))

	readCloser, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer readCloser.Close()

	out := bytes.NewBuffer(nil)
	_, err = io.Copy(out, readCloser)

	glog.Infoln(out)
	return err
}
