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
| `make debug` | build and run the binary with `-debug` and `-v=10`, this will not delete the namespace at the end of the test |
| `make clean` | remove resources created as part of running kube-smoketest |
| `make help` | the default target, i.e. shows this table |

You can manage the verbosity when running the binary directly, setting `-v=2` will add more info logging, using `-v=10` will be used for debugging individual requests where/when necessary.
