# Build the Docker image for the Go node application
docker_build('node', 'src', dockerfile='src/Dockerfile')

# Load Kubernetes specs from the dev directory
k8s_yaml(['dev/configmaps.yaml', 
          'dev/node-0.yaml', 
          'dev/node-1.yaml', 
          'dev/node-2.yaml'])

# Add port forwards for each node
k8s_resource("node-instance-0", port_forwards=[8080])
k8s_resource("node-instance-1", port_forwards=[8081])
k8s_resource("node-instance-2", port_forwards=[8082])
