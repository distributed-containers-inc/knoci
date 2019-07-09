# Knoci 

Knoci ([/'nəʊki/ Know-Key](https://itinerarium.github.io/phoneme-synthesis/?w=/%27n%C9%99%CA%8Aki/)) is an operator that adds a test resource to kubernetes clusters.
This allows you to turn `build && test && deploy` into `build && deploy`, with your pods only coming online after their relevant tests have passed.

## Benefits

1. You can run your tests anywhere you could run a cluster.  This makes it possible to exactly reproduce what happens in a CI pipeline locally
2. Tests can be completely parallelized, using the same autoscaling logic you'd use for pods
3. Knoci checksums your tests, and only runs ones whose files or dependencies have changed

© Distributed Containers Inc. 2019
