// Package smoketests ... this part creates a service for the deployment, to veirfy that vIPs work (kube-proxy and network routing)
package smoketests

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// CreateService creates a ClusterIP service for the Deployment smoketest
func CreateService(ctx context.Context, client *kubernetes.Clientset) error {
	svc, err := client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err == nil && svc != nil {
		// return early, as the service already exists, probably from an earlier run
		glog.V(2).Infof("service %s already exists, not creating a new one", serviceName)
		return nil
	}

	glog.V(2).Info("attempting to create service", serviceName)

	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				"app":     "smoketest",
				"part-of": "smoketest",
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": "smoketest",
			},
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: "http",
					Port: int32(80),
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 80,
					},
					Protocol: v1.ProtocolTCP,
				},
			},
		},
	}

	svc, err = client.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		glog.Errorf("failed to create service %s: %v", serviceName, err)
		return err
	}

	glog.V(2).Infof("successfully created service %s", serviceName)

	return TestService(ctx, client)
}

// CreateNodePortService creates a NodePort service for the Deployment smoketest
func CreateNodePortService(ctx context.Context, client *kubernetes.Clientset) error {
	serviceName := serviceNameNodePort

	svc, err := client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err == nil && svc != nil {
		// return early, as the service already exists, probably from an earlier run
		glog.V(2).Infof("nodePort service %s already exists, not creating a new one", serviceName)
		return nil
	}

	glog.V(2).Info("attempting to create nodePort service", serviceName)

	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				"app":     "smoketest",
				"part-of": "smoketest",
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": "smoketest",
			},
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: "http-np",
					Port: int32(80),
					// NodePort: int32(31080),
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 80,
					},
					Protocol: v1.ProtocolTCP,
				},
			},
		},
	}

	svc, err = client.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		glog.Errorf("failed to create nodePort service %s: %v", serviceName, err)
		return err
	}

	glog.V(2).Infof("successfully created nodePort service %s", serviceName)

	return TestNodePortService(ctx, client)
}

// DeleteService deletes the smoketest service
func DeleteService(ctx context.Context, client *kubernetes.Clientset) error {
	glog.Errorf("failed to delete service %s: %v", serviceName, ErrNotImplemented)
	return ErrNotImplemented
}

// TestService creates a pod and curls the service endpoint, if that was not successful, then a error is returned
func TestService(ctx context.Context, client *kubernetes.Clientset) error {
	glog.V(2).Info("start testing service", serviceName)

	job, err := CreateJob(ctx, client, fmt.Sprintf("wget -o /dev/null -O /dev/null %s && echo \"Success\" || echo \"Failed\"", serviceName))
	if err != nil {
		glog.Errorf("failed to create svc test job: %v", err)
		return err
	}

	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", job.GetLabels()["job-name"]),
	})

	// quick loop as it may take a few seconds for Pods to be scheduled and created
	maxTries := 3
	for maxTries > 0 {
		if len(pods.Items) < 1 {
			pods, err = client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("job-name=%s", job.GetLabels()["job-name"]),
			})
		}
		maxTries--
		if maxTries < 1 {
			return err
		}
		time.Sleep(time.Second)
	}

	pod := pods.Items[0] // no need to guess which pod, as we should only have one that matches the label

	if err = WaitFor(ctx, client, Pod, WithPodName(pod.Name), WithStatus(PodCompleted)); err != nil {
		glog.Errorf("%v", err)
		return err
	}

	output, err := GetPodLogs(ctx, client, pod.Name)
	if err != nil {
		glog.Errorf("%v", err)
	}

	if !strings.Contains(strings.Join(output, " "), "Success") {
		return fmt.Errorf("test failed, did not find \"Success\" in output: %v", output)
	}

	return nil
}

// TestNodePortService calles the NodePort Service on the automatically selected port and
// expects a 200 response, returns an error otherwise
func TestNodePortService(ctx context.Context, client *kubernetes.Clientset) error {
	serviceName := serviceNameNodePort
	glog.V(2).Info("start testing service", serviceName)

	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})

	addresses := nodes.Items[0].Status.DeepCopy().Addresses
	candidateIPs := []string{}

	for _, addr := range addresses {
		if addr.Type != "Hostname" {
			candidateIPs = append(candidateIPs, addr.Address)
		}
	}

	// -- get the nodePort service as we did not specify a port so a random
	//    port can be picked automatically
	svc, err := client.CoreV1().Services(namespace).Get(ctx, serviceNameNodePort, metav1.GetOptions{})
	ports := svc.Spec.DeepCopy().Ports

	var nodePort int32
	for _, port := range ports {
		if port.Name == "http-np" {
			nodePort = port.NodePort
		}
	}
	// -- setup http connection
	url := fmt.Sprintf("http://%s:%d", candidateIPs[0], nodePort)
	glog.V(10).Info(serviceNameNodePort, " url: ", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	hc := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: time.Second,
			}).Dial,
			TLSHandshakeTimeout:   time.Second,
			ResponseHeaderTimeout: time.Second,
			ExpectContinueTimeout: time.Second,
		},
		Timeout: 5 * time.Second,
	}

	// --- Do the http request with custom timeout
	resp, err := hc.Do(req)

	if resp == nil && err != nil {
		maxTries := 3
		time.Sleep(time.Second)

		for maxTries > 0 {
			maxTries--
			glog.V(10).Infof("request failed: %v, retrying.. will retry %d more times", err, maxTries)
			resp, err = hc.Do(req)
			if resp != nil && err == nil {
				break
			}
			time.Sleep(time.Second)
		}

	}

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("%v", resp.Status)
	}

	if glog.V(10) {
		bbody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			glog.Warningf("failed to read response body: %v", err)
			return nil
		}
		glog.V(10).Infof("%s", string(bbody))
	}

	return nil
}
