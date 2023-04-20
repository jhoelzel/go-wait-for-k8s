./go-wait-for --namespace=default --label-selector=app=readiness-test --resource-type=deployment
./go-wait-for --namespace=default --label-selector=app=readiness-test --resource-type=job
./go-wait-for --namespace=default --label-selector=app=readiness-test --resource-type=statefulset
./go-wait-for --namespace=default --label-selector=app=readiness-test --resource-type=daemonset
./go-wait-for --namespace=default --label-selector=app=readiness-test --resource-type=replicaset
