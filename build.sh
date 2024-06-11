make docker-build
docker image save vs-clean:latest -o ./vs-clean
minikube image load vs-clean -p dr1
minikube image load vs-clean -p dr2
kubectl --context dr1 delete deployment.apps/vs-clean-controller-manager -n vs-clean-system
kubectl --context dr2 delete deployment.apps/vs-clean-controller-manager -n vs-clean-system
kubectl kustomize config/default | kubectl --context dr1 apply -f -
kubectl kustomize config/default | kubectl --context dr2 apply -f -
