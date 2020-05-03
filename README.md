# kube-smoketest

`kube-smoketest` is a minimum set of automated tests that can run against a freshly created cluster, independent where that cluster
is running or what was used to set it up. It's not meant to be a exhaustive test suite, more a quick basic functionality check. It's
inspired by [kelseyhightower/kubernetes-the-hard-way](https://github.com/kelseyhightower/kubernetes-the-hard-way/blob/master/docs/13-smoke-test.md)'s
list of smoketest tasks.

# setup & configuration

`kube-smoketest` requires very few things to be able to run all its tests.

## kube config

`kube-smoketest` uses by default the currently configure kubernetes cluster, using `~/.kube/config`'s active context, to change this, set the `KUBECONFIG` environment variable to an alternative config file.

## etcd certs, keys and CA

`kube-smoketest` requires a valid etcd client certificate and key, and the
corresponding etcd CA certificate. It looks for the following files

- `etcd.ca` the CA cert
- `etcd.crt` the client certificate
- `etcd.key` the client key

to be present in the direcotry the binary is run from.

# tests

`kube-smoketest` runs the following tests in sequence ...

- check componentstatuses
    - verifies that essential components are working
- create the `kube-smoketest` namespace
    - this is where all test resources are going to be created in
- create a pod, wait for pod, get its logs
    - uses `busybox` container image
- create a deployment
    - uses `nginx` container image
- create a service (using the _deployment_)
    - a standard ClusterIP service, tested for internal access
    - test is run as a `job` resource
- create a node port service (using the _deployment_)
    - the NodePort service uses a random port allocated by k8s
- create a secret, check etcd for `:enc:` string in hexdump
    - creates a opaque secret, then checks etcd for the key's value
    - this test requires `etcd.ca`, `etcd.crt` and `etcd.key` to be present
    - **test will succeed even if value is found _not_ to be _encrypted at rest_**
- delete the `kube-smoketest` namespace

# build, run, clean-up

| command      | description |
| ------------ | ----------- |
| `make help`  | the default target, i.e. shows these options |
| `make build` | build the binary |
| `make run`   | build and run the binary |
| `make debug` | build and run the binary with `-debug` and `-v=10`, this will also skip deletion of the namespace at the end |
| `make clean` | deletes kube-smoketest namespace |

## debugging

You can manage the verbosity when running the binary directly, setting `-v=2` will print additional info logging that may be useful, using `-v=10` will be used for debugging individual requests where/when necessary.

# example

Here's an example output run at default verbosity.

```
➜ make run
Building kube-smoketest binary..
Running kube-smoketest..
I0502 19:01:48.311490   62000 main.go:57] 	✅ Component statuses
I0502 19:01:49.336117   62000 main.go:69] 	✅ Create namespace
I0502 19:02:03.423735   62000 main.go:80] 	✅ Pod + Logs
I0502 19:02:35.383497   62000 main.go:91] 	✅ Deployment
I0502 19:02:37.715421   62000 main.go:102] 	✅ Service
I0502 19:02:38.952574   62000 main.go:113] 	✅ NodePort Service
W0502 19:02:39.222429   62000 secrets.go:148] 	⚠️  the kubernetes secret "smoketest-secret" is not encrypted at rest
I0502 19:02:39.222549   62000 main.go:124] 	✅ Secret
I0502 19:02:39.345329   62000 main.go:141] 	✅ Delete namespace

-------------------- RESULT --------------------

I0502 19:02:39.345377   62000 main.go:156] 	✅ SUCCESS: all tests passed
```
