module kubernetes-services-deployment

go 1.12

require (
	bitbucket.org/cloudplex-devs/woodpecker v0.0.0-20191217140035-e9c3629fe038
	cloud.google.com/go v0.40.0 // indirect
	github.com/alecthomas/template v0.0.0-20160405071501-a0175ee3bccc
	github.com/gedex/inflector v0.0.0-20170307190818-16278e9db813
	github.com/gin-gonic/gin v1.3.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/golang/protobuf v1.3.3
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.0.0-20190111052518-d48bb9075efa // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/patrickmn/go-cache v0.0.0-20180815053127-5633e0862627
	github.com/pkg/errors v0.8.1
	github.com/swaggo/gin-swagger v0.0.0-20190110070702-0c6fcfd3c7f3
	github.com/swaggo/swag v1.4.0
	github.com/tmc/scp v0.0.0-20170824174625-f7b48647feef // indirect
	github.com/urfave/cli/v2 v2.0.0
	go.opencensus.io v0.22.2
	google.golang.org/appengine v1.6.1 // indirect
	google.golang.org/grpc v1.28.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	gopkg.in/resty.v1 v1.11.0
	k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver v0.0.0-20190116054503-cf30b7cf64c2
	k8s.io/apimachinery v0.16.4
	k8s.io/client-go v0.16.4
)

replace (
	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.4
)
