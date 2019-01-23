# kubernetes-services-deployment

Deploy services on a deployed Kubernetes cluster

###Build
```
docker build -t kubernetes-deployment-engine . 
```

###Docker Run Command
```
   docker  run -d \
   --name kube-deployment-engine\
   --restart always \
   -p 8089:8089 \
   kubernetes-deployment-engine
```

#### Swagger
`http://localhost:8089/swagger/index.html`