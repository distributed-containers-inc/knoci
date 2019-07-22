# Knoci 

Knoci ([/'nəʊki/ Know-Key](https://itinerarium.github.io/phoneme-synthesis/?w=/%27n%C9%99%CA%8Aki/)) is an operator that adds a test resource to kubernetes clusters.
This allows you to turn `build && test && deploy` into `build && deploy`, with your pods only coming online after their relevant tests have passed.

## Benefits

1. You can run your tests anywhere you could run a cluster.  This makes it possible to exactly reproduce what happens in a CI pipeline locally
2. Tests can be completely parallelized, using the same autoscaling logic you'd use for pods (see [example/parallelism](https://github.com/distributed-containers-inc/knoci/tree/master/example/parallelism))
3. Knoci checksums your tests, and only runs ones whose files or dependencies have changed

## Usage

### Installing the Operator
See `deploy/in/knoci.yaml.tmpl` for an RBAC example

### Deploying your tests
#### Complete spec for a Test

```
# apiVersion and kind are frozen for a specific release, use the ones defined here
apiVersion: tests.knoci.distributedcontainers.com/v1alpha1
kind: Test
metadata:
  # name and namespace mean the same thing they do for other kubernetes objects
  name: api-unit-tests
  namespace: api-unit-tests
spec:
  image: registry.example.com/ApiUnitTests:v1.0.0
status:
  state: Running
```

## Building

To build knoci,
1. First install [sanic](https://github.com/distributed-containers-inc/sanic) and its dependencies
2. Run `sanic env dev sanic build` to build the latest version of the docker image.

© Distributed Containers Inc. 2019