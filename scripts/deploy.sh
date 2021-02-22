. ./scripts/version.sh

name="login-service"
version="$login_service_version"
tag="$name"
repository="repository/geeksheik9"
namespace="star-wars-dnd"
registry="hub.docker.com"
versionedImageName="$tag:$version"
taggedImageName="$registry/$repository/$versionedImageName"

echo "time to deploy!"
helm upgrade -i ${name} --namespace=${namespace} --set image.repository=${registry}/${repository}/${name},ingress.env=rancher,ingress.domain=geeksheiks-lab.com,image.tag=${version}