// Package smoketests ... create a secret, verify that its content is in fact encrypted
package smoketests

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateSecret ... creates a secret
func CreateSecret(ctx context.Context, client *kubernetes.Clientset) error {

	exists, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err == nil && exists != nil {
		glog.V(2).Infof("not creating secret %s, already exists", secretName)
		return nil
	}

	glog.V(2).Infof("creating secret %s", secretName)

	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"user": []byte("YWRtaW4K"),
		},
	}

	secret, err = client.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create secret: %v", err)
	}

	glog.V(2).Infof("successfully created secret %s", secretName)

	// verify the secret is encrypted ...

	if err = TestSecret(ctx, client); err != nil {
		return err
	}

	return nil
}

// TestSecret verifies the secret directly interrogating etcd,
// it checks the secret's etcd content for a encryption prefix
func TestSecret(ctx context.Context, client *kubernetes.Clientset) error {
	glog.V(2).Infoln("start verifying secret is encrypted")
	// find ETCD hosts in cluster .. this only works in stacked deployment scenarios for now (e.g. kubeadm was used to bootstrap)
	masterNodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: "node-role.kubernetes.io/master=",
	})

	nodes := masterNodes.DeepCopy().Items
	etcdEndpoints := []string{}
	for _, n := range nodes {
		addresses := n.Status.DeepCopy().Addresses
		for _, addr := range addresses {
			if addr.Type == "InternalIP" {
				etcdEndpoints = append(etcdEndpoints, fmt.Sprintf("%s:%d", addr.Address, etcdPort))
			}
		}
	}

	glog.V(10).Infof("list of etcd endpoints found: %v", etcdEndpoints)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithTimeout(time.Second),
		},
		DialKeepAliveTimeout: time.Second,
		LogConfig: &zap.Config{
			Level:    zap.NewAtomicLevelAt(zapcore.ErrorLevel),
			Encoding: "console",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %v", err)
	}

	etcdCtx, etcdCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer etcdCancel()

	if _, err := cli.Cluster.MemberList(etcdCtx); err != nil {
		return fmt.Errorf("failed to get etcd members using endpoint/s %v: %v", etcdEndpoints, err)
	}

	if err != nil {
		return fmt.Errorf("failed to create etcd client: %v", err)
	}
	defer cli.Close()

	etcdCtx, etcdCancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer etcdCancel()

	resp, err := cli.KV.Get(etcdCtx, secretName)
	if err != nil {
		return fmt.Errorf("failed to get etcd key %s: %v", secretName, err)
	}

	// TODO: actually check for the correct prefix , but how do we do hexdump ? is that even necessary ?
	for _, kv := range resp.Kvs {
		val := kv.Value
		fmt.Println(val)
	}

	return nil
}
