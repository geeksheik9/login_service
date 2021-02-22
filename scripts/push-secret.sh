echo deleting old secret
kubectl delete secret db-secrets --namespace=star-wars-dnd

echo creating secret
kubectl create secret generic db-secrets --from-file=./db-secret.json --namespace=star-wars-dnd