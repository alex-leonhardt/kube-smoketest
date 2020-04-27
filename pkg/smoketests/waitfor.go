package smoketests

import (
	"context"
	"errors"
	"time"

	"github.com/golang/glog"
	"github.com/jpillora/backoff"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Resource is a k8s resource
type Resource int

// Enums for Resource
const (
	Namespace Resource = iota + 1
	Pod
	Deployment
	StatefulSet
	PVC
	ConfigMap
	Secret
)

// ErrUnknownResourceType is returned when the resource type provided is not being dealt with (yet)
var ErrNotImplemented = errors.New("not yet implemented")

// ErrUnknownResourceType is returned when the resource type provided is not being dealt with (yet)
var ErrUnknownResourceType = errors.New("unknown resource type")

// ErrWrongTypeForArgument is returned when a optinal argument was given, but its value had the wrong type (e.g. int instead of int32)
var ErrWrongTypeForArgument = errors.New("wrong type for optional argument")

// --- optinoal arguments to WaitFor

type options struct {
	NumReady int32
	PodName  string
}

// Option represents a optional argument to WaitFor
type Option interface {
	apply(*options)
}

// ---
type podNameOption string

func (s podNameOption) apply(opts *options) {
	opts.PodName = string(s)
}

// WithPodName sets PodName
func WithPodName(n string) Option {
	return podNameOption(n)
}

// ---
type numReadyOption int32

func (s numReadyOption) apply(opts *options) {
	opts.NumReady = int32(s)
}

// WithNumReady sets NumReady as int32
func WithNumReady(n int32) Option {
	return numReadyOption(n)
}

// ---

// WaitFor waits for a resource to be in a ready, unready, etc. state and
// returns with an error when the ctx timed out, or with nil
func WaitFor(ctx context.Context, client *kubernetes.Clientset, resource Resource, opts ...Option) error {

	options := options{}
	for _, o := range opts {
		o.apply(&options)
	}

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

		switch resource {
		case Namespace:
			return ErrNotImplemented
		case Deployment:
			deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, "smoketest", metav1.GetOptions{})
			if err != nil {
				continue
			}
			if deployment != nil {
				if deployment.Status.AvailableReplicas == options.NumReady {
					return nil
				}
			}
			glog.V(2).Infof("waiting for pods to become available: %v", time.Since(t))

		case Pod:
			tmpPod, err := client.CoreV1().Pods(namespace).Get(ctx, options.PodName, metav1.GetOptions{})
			if err != nil {
				continue
			}
			if tmpPod.Status.Phase == v1.PodRunning {
				return nil
			}
			glog.V(2).Infof("waiting for pod to be ready: %v", time.Since(t))

		case StatefulSet:
			return ErrNotImplemented
		case PVC:
			return ErrNotImplemented
		case ConfigMap:
			return ErrNotImplemented
		case Secret:
			return ErrNotImplemented
		default:
			return ErrUnknownResourceType
		}

	}
}
