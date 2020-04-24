# kube-smoketest

kube-smoketest in a nutshell does *all the tests described in [kelseyhightower/kubernetes-the-hard-way](https://github.com/kelseyhightower/kubernetes-the-hard-way/blob/master/docs/13-smoke-test.md)'s smoke test list.

\* - _well, almost ;)_
## TL;DR

The basic idea is to have a minimum set of automated tests that one can run against a freshly created cluster, independent where that cluster
is running or what was used to set it up. It's by no means meant to be exhaustive, but a quick check that basic functionality is working.

# build, run, clean-up

| command | desc |
| ------- | ---- |
| `make build` | build the binary |
| `make run` | build and run the binary |
| `make clean` | remove resources created as part of running kube-smoketest |
