package main

import (
	"bufio"
	"fmt"
	"gopkg.in/resty.v1"
	"strconv"
	"strings"
)

func main() {
	payload := `
Roles:              <none>
Labels:             beta.kubernetes.io/arch=amd64
                    beta.kubernetes.io/instance-type=n1-standard-4
                    beta.kubernetes.io/os=linux
                    dedicated=master
                    failure-domain.beta.kubernetes.io/region=us-east1
                    failure-domain.beta.kubernetes.io/zone=us-east1-b
                    kubernetes.io/arch=amd64
                    kubernetes.io/hostname=application-hyuty4-1
                    kubernetes.io/os=linux
                    nodepool=application-hyuty4-1
Annotations:        flannel.alpha.coreos.com/backend-data: {"VtepMAC":"46:70:07:ff:0e:55"}
                    flannel.alpha.coreos.com/backend-type: vxlan
                    flannel.alpha.coreos.com/kube-subnet-manager: true
                    flannel.alpha.coreos.com/public-ip: 10.182.96.2
                    node.alpha.kubernetes.io/ttl: 0
                    volumes.kubernetes.io/controller-managed-attach-detach: true
CreationTimestamp:  Wed, 13 May 2020 06:03:26 +0000
Taints:             <none>
Unschedulable:      false
Conditions:
  Type                 Status  LastHeartbeatTime                 LastTransitionTime                Reason                       Message
  ----                 ------  -----------------                 ------------------                ------                       -------
  NetworkUnavailable   False   Mon, 01 Jan 0001 00:00:00 +0000   Mon, 01 Jan 0001 00:00:00 +0000   RouteCreated                 Manually set through k8s api
  MemoryPressure       False   Thu, 14 May 2020 05:33:45 +0000   Wed, 13 May 2020 06:03:26 +0000   KubeletHasSufficientMemory   kubelet has sufficient memory available
  DiskPressure         False   Thu, 14 May 2020 05:33:45 +0000   Wed, 13 May 2020 06:03:26 +0000   KubeletHasNoDiskPressure     kubelet has no disk pressure
  PIDPressure          False   Thu, 14 May 2020 05:33:45 +0000   Wed, 13 May 2020 06:03:26 +0000   KubeletHasSufficientPID      kubelet has sufficient PID available
  Ready                True    Thu, 14 May 2020 05:33:45 +0000   Wed, 13 May 2020 06:03:46 +0000   KubeletReady                 kubelet is posting ready status. AppArmor enabled
Addresses:
  InternalIP:   10.182.96.2
  ExternalIP:   35.196.229.48
  InternalDNS:  application-hyuty4-1.us-east1-b.c.cloudplex-nextgen-test-project.internal
  Hostname:     application-hyuty4-1
Capacity:
 attachable-volumes-gce-pd:  127
 cpu:                        4
 ephemeral-storage:          50633164Ki
 hugepages-1Gi:              0
 hugepages-2Mi:              0
 memory:                     15389976Ki
 pods:                       110
Allocatable:
 attachable-volumes-gce-pd:  127
 cpu:                        4
 ephemeral-storage:          46663523866
 hugepages-1Gi:              0
 hugepages-2Mi:              0
 memory:                     15287576Ki
 pods:                       110
System Info:
 Machine ID:                 5bf5897079c520ba8c0310305a37931c
 System UUID:                5bf58970-79c5-20ba-8c03-10305a37931c
 Boot ID:                    d521058b-d083-4d13-b140-c49ef637029f
 Kernel Version:             5.0.0-1033-gcp
 OS Image:                   Ubuntu 18.04.4 LTS
 Operating System:           linux
 Architecture:               amd64
 Container Runtime Version:  docker://17.12.1-ce
 Kubelet Version:            v1.16.0
 Kube-Proxy Version:         v1.16.0
PodCIDR:                     10.244.0.0/24
PodCIDRs:                    10.244.0.0/24
ProviderID:                  gce://cloudplex-nextgen-test-project/us-east1-b/application-hyuty4-1
Non-terminated Pods:         (37 in total)
  Namespace                  Name                                            CPU Requests  CPU Limits    Memory Requests  Memory Limits     AGE
  ---------                  ----                                            ------------  ----------    ---------------  -------------     ---
  cloud-run-events           controller-6788f7b487-jfs8n                     100m (2%)     1 (25%)       100Mi (0%)       1000Mi (6%)       23h
  cloud-run-events           webhook-69ffff6c9-rf94b                         20m (0%)      200m (5%)     20Mi (0%)        200Mi (1%)        23h
  default                    service-pgn6ad-v1-6bcb69d95d-wwsgd              10m (0%)      2 (50%)       40Mi (0%)        1Gi (6%)          23h
  istio-system               cluster-local-gateway-fcd7c747f-vnhk5           10m (0%)      0 (0%)        0 (0%)           0 (0%)            23h
  istio-system               grafana-d8465c484-ghbfx                         10m (0%)      0 (0%)        0 (0%)           0 (0%)            23h
  istio-system               istio-citadel-59574746c-rqrn7                   10m (0%)      0 (0%)        0 (0%)           0 (0%)            23h
  istio-system               istio-egressgateway-8448d5b89f-zn2qr            10m (0%)      2 (50%)       40Mi (0%)        1Gi (6%)          23h
  istio-system               istio-galley-7c6786768c-qc28r                   10m (0%)      0 (0%)        0 (0%)           0 (0%)            23h
  istio-system               istio-ingressgateway-54865dd67b-sdhqp           10m (0%)      2 (50%)       40Mi (0%)        1Gi (6%)          23h
  istio-system               istio-pilot-7998c74b4c-4v6tz                    20m (0%)      2 (50%)       140Mi (0%)       1Gi (6%)          23h
  istio-system               istio-policy-69cdb87cc8-5sbbd                   20m (0%)      2 (50%)       140Mi (0%)       1Gi (6%)          23h
  istio-system               istio-sidecar-injector-949f4564-ptmzd           10m (0%)      0 (0%)        0 (0%)           0 (0%)            23h
  istio-system               istio-telemetry-557d899cd4-r857t                60m (1%)      6800m (170%)  140Mi (0%)       5073741824 (32%)  23h
  istio-system               istio-tracing-55c965d5b6-hph66                  10m (0%)      0 (0%)        0 (0%)           0 (0%)            23h
  istio-system               kiali-5cd7c8b6d-7vr6h                           10m (0%)      0 (0%)        0 (0%)           0 (0%)            23h
  istio-system               prometheus-6f74d6f76d-f6t7j                     10m (0%)      0 (0%)        0 (0%)           0 (0%)            23h
  knative-eventing           eventing-controller-697cd9f8df-497x8            0 (0%)        0 (0%)        0 (0%)           0 (0%)            23h
  knative-eventing           eventing-webhook-58bd6f9957-ms2xm               0 (0%)        0 (0%)        0 (0%)           0 (0%)            23h
  knative-eventing           imc-controller-6995b49d4d-dghcf                 0 (0%)        0 (0%)        0 (0%)           0 (0%)            23h
  knative-eventing           imc-dispatcher-779d8457bd-7s9dz                 0 (0%)        0 (0%)        0 (0%)           0 (0%)            23h
  knative-eventing           sources-controller-55f467cd5-bcxb4              100m (2%)     0 (0%)        100Mi (0%)       0 (0%)            23h
  knative-serving            activator-77fc555665-p7c2c                      310m (7%)     3 (75%)       100Mi (0%)       1624Mi (10%)      23h
  knative-serving            autoscaler-5c98b7c9b6-vqlbj                     40m (1%)      2300m (57%)   80Mi (0%)        1424Mi (9%)       23h
  knative-serving            autoscaler-hpa-5cfd4f6845-tq2zq                 100m (2%)     1 (25%)       100Mi (0%)       1000Mi (6%)       23h
  knative-serving            controller-7fd74c8f67-fkspn                     100m (2%)     1 (25%)       100Mi (0%)       1000Mi (6%)       23h
  knative-serving            networking-istio-7587d6dbf5-ggb8f               100m (2%)     1 (25%)       100Mi (0%)       1000Mi (6%)       23h
  knative-serving            webhook-74847bb77c-nkgqp                        20m (0%)      200m (5%)     20Mi (0%)        200Mi (1%)        23h
  knative-sources            github-controller-manager-0                     100m (2%)     1 (25%)       100Mi (0%)       1000Mi (6%)       23h
  kube-system                canal-zxk6h                                     250m (6%)     0 (0%)        0 (0%)           0 (0%)            23h
  kube-system                fluentd-pchq2                                   100m (2%)     0 (0%)        200Mi (1%)       200Mi (1%)        23h
  kube-system                kibana-logging-6f589988c5-bdjkl                 100m (2%)     1 (25%)       0 (0%)           0 (0%)            23h
  kube-system                kube-dns-5b6cd5854c-rxtxh                       260m (6%)     0 (0%)        110Mi (0%)       170Mi (1%)        23h
  kube-system                kube-dns-5b6cd5854c-srcv5                       260m (6%)     0 (0%)        110Mi (0%)       170Mi (1%)        23h
  kube-system                metrics-server-7f677969c6-d4vv6                 0 (0%)        0 (0%)        0 (0%)           0 (0%)            23h
  logging                    elasticsearch-test-7dcc6746f6-khtsh             50m (1%)      100m (2%)     4Gi (27%)        4Gi (27%)         23h
  tekton-pipelines           tekton-pipelines-controller-55f95fbb9b-9j929    0 (0%)        0 (0%)        0 (0%)           0 (0%)            23h
  tekton-pipelines           tekton-pipelines-webhook-6dd8446c85-vkhtr       0 (0%)        0 (0%)        0 (0%)           0 (0%)            23h
Allocated resources:
  (Total limits may be over 100 percent, i.e., overcommitted.)
  Resource                   Requests      Limits
  --------                   --------      ------
  cpu                        2220m (55%)   28600m (715%)
  memory                     5876Mi (39%)  23595722Ki (154%)
  ephemeral-storage          0 (0%)        0 (0%)
  attachable-volumes-gce-pd  0             0
Events:                      <none>
`

	//print(payload)
	scanner := bufio.NewScanner(strings.NewReader(payload))

	for scanner.Scan() {
		//fmt.Println(".........")
		temp := scanner.Text()
		if strings.Contains(temp, "Non-terminated Pods:") {
			pods, err := strconv.Atoi(strings.Split(strings.Split(temp, "(")[1], " ")[0])
			_ = err
			println(pods) //PODS
		}
		if strings.Contains(temp, "cpu") && strings.Contains(temp, "%") {
			cpu, err := strconv.Atoi(strings.Split(strings.Split(temp, "(")[1], "%")[0]) //CPU
			_ = err
			println(cpu) //PODS
		}
		if strings.Contains(temp, "memory") && strings.Contains(temp, "%") {
			memory, err := strconv.Atoi(strings.Split(strings.Split(temp, "(")[1], "%")[0]) //Memory
			_ = err
			print(memory)
		}

		if strings.Contains(temp, "ephemeral-storage") && strings.Contains(temp, "%") {
			storage, err := strconv.Atoi(strings.Split(strings.Split(temp, "(")[1], "%")[0]) //Storage
			_ = err
			print(storage)
		}
	}

	return
	payload = `
{
	    
    "project_id": "haseeb-test",
    "service": {
        "service-account": [
            {
                "kind": "ServiceAccount",
                "apiVersion": "v1",
                "metadata": {
                    "name": "emailservice-sa",
                    "namespace": "default",
                    "creationTimestamp": null
                },
                "secrets": [
                    {
                        "name": "emailservice-reg-secret"
                    }
                ],
                "imagePullSecrets": [
                    {
                        "name": "emailservice-cfg-secret"
                    }
                ]
            }
        ],
        "knative-component": [
            {
                "kind": "Service",
                "apiVersion": "serving.knative.dev/v1alpha1",
                "metadata": {
                    "name": "emailservice",
                    "namespace": "default",
                    "creationTimestamp": null
                },
                "spec": {
                    "runLatest": {
                        "configuration": {
                            "build": {
                                "kind": "Build",
                                "apiVersion": "build.knative.dev/v1alpha1",
                                "metadata": {
                                    "creationTimestamp": null
                                },
                                "spec": {
                                    "source": {
                                        "git": {
                                            "url": "https://github.com/junaidk/simple-app.git",
                                            "revision": "master"
                                        }
                                    },
                                    "serviceAccountName": "emailservice-sa",
                                    "template": {
                                        "name": "kaniko",
                                        "arguments": [
                                            {
                                                "name": "IMAGE",
                                                "value": "gcr.io/junaidk/app-from-source:latest:1.0"
                                            }
                                        ]
                                    },
                                    "Status": ""
                                },
                                "status": {
                                    "stepsCompleted": null
                                }
                            },
                            "revisionTemplate": {
                                "metadata": {
                                    "creationTimestamp": null
                                },
                                "spec": {
                                    "serviceAccountName": "emailservice-sa",
                                    "container": {
                                        "name": "",
                                        "image": "gcr.io/junaidk/app-from-source:latest:1.0",
                                        "env": [
                                            {
                                                "name": "junaid",
                                                "value": "test"
                                            }
                                        ],
                                        "resources": {},
                                        "imagePullPolicy": "Always"
                                    }
                                }
                            }
                        }
                    }
                },
                "status": {}
            },
            {
                "kind": "BuildTemplate",
                "apiVersion": "build.knative.dev/v1alpha1",
                "metadata": {
                    "name": "kaniko",
                    "creationTimestamp": null
                },
                "spec": {
                    "parameters": [
                        {
                            "name": "IMAGE",
                            "description": "The name of the image to push"
                        },
                        {
                            "name": "DOCKERFILE",
                            "description": "Path to the Dockerfile to build.",
                            "default": "/workspace/Dockerfile"
                        }
                    ],
                    "steps": [
                        {
                            "name": "build-and-push",
                            "image": "gcr.io/kaniko-project/executor",
                            "args": [
                                "--dockerfile=${DOCKERFILE}",
                                "--destination=${IMAGE}"
                            ],
                            "env": [
                                {
                                    "name": "DOCKER_CONFIG",
                                    "value": "/builder/home/.docker"
                                }
                            ],
                            "resources": {}
                        }
                    ],
                    "volumes": null
                }
            }
        ],
        "secrets": [
            {
                "kind": "Secret",
                "apiVersion": "v1",
                "metadata": {
                    "name": "emailservice-reg-secret",
                    "namespace": "default",
                    "creationTimestamp": null,
                    "annotations": {
                        "build.knative.dev/docker-0": "https://us.gcr.io",
                        "build.knative.dev/docker-1": "https://gcr.io",
                        "build.knative.dev/docker-2": "https://eu.gcr.io",
                        "build.knative.dev/docker-3": "https://asia.gcr.io"
                    }
                },
                "data": {
                    "password": "ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsCiAgInByb2plY3RfaWQiOiAiY2xvdWRwbGV4LW5leHRnZW4iLAogICJwcml2YXRlX2tleV9pZCI6ICIzMThmYTliMGNkNDc2YTUwYmQzODdjNmRmMWE2Njk3MjY2MjU1NDFiIiwKICAicHJpdmF0ZV9rZXkiOiAiLS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tXG5NSUlFdkFJQkFEQU5CZ2txaGtpRzl3MEJBUUVGQUFTQ0JLWXdnZ1NpQWdFQUFvSUJBUURYaVBMODJSblRUTEdRXG4vbHk0OXUvK29zTGl4Tm9VTzg2cmtDbWxvbkd5Rzl1bURhOVZuUmJVazNFSk1TSlJ0Qzd2QzFWbEZ5SCs5SkI5XG41VHJOdmdGSEprUVpUenJIQzZwdjlRNzNVSjgvWHhVQUhUSkZoQ3R1NFlUQWgraHVaQ3VLUjBaM21XSGwrZ3NUXG43aDVhYTVpM3hmb0dyTUFlUVFxYk1zajBxNmwvU24rN1pXWHVYVEZYTk1Nb3VkOGhNVGN0eWNkdURyRjFIdFF5XG52dGIzOFppT0xuOEZxUnIxUmN6cG4xTGgxK085SVV6Z0swcXFwNFVUbkh0dDgvcnUyZFQ2RytCK0NpelVlVzNTXG5zc1ZtQnJiQ09SMTN3UkFwK0x2eXRtSmFjU1ZMdE56YzdyV0VaM3FxTDRZZnYzalFUdjRFQUU0ZDFROTFnQVZHXG5henVrbThVbkFnTUJBQUVDZ2dFQVJsUU0rWDA3cEl0amEwemNZNHhMN2pvY3ZsTCtWOWpXREh3cllyMFpQVVZDXG56REd0OHhFaG1IYU56VEtIb01KYkNDd2FEclRZSm5HVnprYWtnK3JLVHZXSkJmcXG5ENGgxeDcvREFsS2RQVHg5bUhna3YrK1BlN1l4ZnVDZmNpcmZwRjBDZ1lBbHlvaXdId3gwbDlxcFliay9yQjkvXG5RTFRvckhCSlFpTGhVRFhYcWJTUjdtNDB0Z2FCd240blBZQUhUWHc2V2VNOFpDWUc3ZUVKY05lR085UnlqUWU2XG5Kb3g3RlVCU2c2alBqNytuYUx5alR6OU1FaVhvaUVTek45dXVCZG14Vm1ldkNkREtpY3lyS0IvNE8yR1BzRHFvXG5DTjErRmhLdWdsQnNFMHZPZXZEMmpnPT1cbi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS1cbiIsCiAgImNsaWVudF9lbWFpbCI6ICJjbG91ZHBsZXgtbmV4dGdlbi1wdXNoLWltYWdlc0BjbG91ZHBsZXgtbmV4dGdlbi5pYW0uZ3NlcnZpY2VhY2NvdW50LmNvbSIsCiAgImNsaWVudF9pZCI6ICIxMTYyMzIzNDQ5MTA0OTc5NjQ4MzgiLAogICJhdXRoX3VyaSI6ICJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20vby9vYXV0aDIvYXV0aCIsCiAgInRva2VuX3VyaSI6ICJodHRwczovL29hdXRoMi5nb29nbGVhcGlzLmNvbS90b2tlbiIsCiAgImF1dGhfcHJvdmlkZXJfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9vYXV0aDIvdjEvY2VydHMiLAogICJjbGllbnRfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9yb2JvdC92MS9tZXRhZGF0YS94NTA5L2Nsb3VkcGxleC1uZXh0Z2VuLXB1c2gtaW1hZ2VzJTQwY2xvdWRwbGV4LW5leHRnZW4uaWFtLmdzZXJ2aWNlYWNjb3VudC5jb20iCn0=",
                    "username": "anVuYWlkaw=="
                },
                "type": "kubernetes.io/basic-auth"
            },
            {
                "kind": "Secret",
                "apiVersion": "v1",
                "metadata": {
                    "name": "emailservice-cfg-secret",
                    "namespace": "default",
                    "creationTimestamp": null
                },
                "data": {
                    ".dockercfg": "eyJnY3IuaW8iOnsiYXV0aCI6ImFuVnVZV2xrYXpwN0NpQWdJblI1Y0dVaU9pQWljMlZ5ZG1salpWOWhZMk52ZFc1MElpd0tJQ0FpY0hKdmFtVmpkRjlwWkNJNklDSmpiRzkxWkhCc1pYZ3RibVY0ZEdkbGJpSXNDaUFnSW5CeWFYWmhkR1ZmYTJWNVgybGtJam9nSWpNeE9HWmhPV0l3WTJRME56WmhOVEJpWkRNNE4yTTJaR1l4WVRZMk9UY3lOall5TlRVME1XSWlMQW9nSUNKd2NtbDJZWFJsWDJ0bGVTSTZJQ0l0TFMwdExVSkZSMGxPSUZCU1NWWkJWRVVnUzBWWkxTMHRMUzFjYmsxSlNVVjJRVWxDUVVSQlRrSm5hM0ZvYTJsSE9YY3dRa0ZSUlVaQlFWTkRRa3RaZDJkblUybEJaMFZCUVc5SlFrRlJSRmhwVUV3NE1sSnVWRlJNUjFGY2JpOXNlVFE1ZFM4cmIzTk1hWGhPYjFWUE9EWnlhME50Ykc5dVIzbEhPWFZ0UkdFNVZtNVNZbFZyTTBWS1RWTktVblJETjNaRE1WWnNSbmxJS3psS1FqbGNialZVY2s1MlowWklTbXRSV2xSNmNraERObkIyT1ZFM00xVktPQzlZZUZWQlNGUktSbWhEZEhVMFdWUkJhQ3RvZFZwRGRVdFNNRm96YlZkSWJDdG5jMVJjYmpkb05XRmhOV2t6ZUdadlIzSk5RV1ZSVVhGaVRYTnFNSEUwWkZNMFlUVnViVVJNVlRSbVVFcHlNME0yWlZSeVYwbHJXa3hZY0VoeU1YTjBTVVUwU3paVk5rNVNRVzlIUVV0TVIxQmNibGczUldwbE5tbHRZa2xFT0RCQ1REQm5Oa1Z5Y1VjMlJFUk9Sa3QxYVM4eldYbGxNRkZqT1ZwWVExbHVhVXM1ZFhBeGExTjZSR013VUhSeGRrcDNkamxjYmpOUlNWTlFjRFJZUXpGSFpub3ZPVGxrZDBWVU5taE9ZMjVGYm00M2VYbDJZeTlyTUZScFJ6TldaM0ZDZDFNMlJFYzBlRFozUkVKNGFIZEdVeXN4YzB0Y2JrUTBhREY0Tnk5RVFXeExaRkJVZURsdFNHZHJkaXNyVUdVM1dYaG1kVU5tWTJseVpuQkdNRU5uV1VGc2VXOXBkMGgzZURCc09YRndXV0pyTDNKQ09TOWNibEZNVkc5eVNFSktVV2xNYUZWRVdGaHhZbE5TTjIwME1IUm5ZVUozYmpSdVVGbEJTRlJZZHpaWFpVMDRXa05aUnpkbFJVcGpUbVZIVHpsU2VXcFJaVFpjYmtwdmVEZEdWVUpUWnpacVVHbzNLMjVoVEhscVZIbzVUVVZwV0c5cFJWTjZUamwxZFVKa2JYaFdiV1YyUTJSRVMybGplWEpMUWk4MFR6SkhVSE5FY1c5Y2JrTk9NU3RHYUV0MVoyeENjMFV3ZGs5bGRrUXlhbWM5UFZ4dUxTMHRMUzFGVGtRZ1VGSkpWa0ZVUlNCTFJWa3RMUzB0TFZ4dUlpd0tJQ0FpWTJ4cFpXNTBYMlZ0WVdsc0lqb2dJbU5zYjNWa2NHeGxlQzF1WlhoMFoyVnVMWEIxYzJndGFXMWhaMlZ6UUdOc2IzVmtjR3hsZUMxdVpYaDBaMlZ1TG1saGJTNW5jMlZ5ZG1salpXRmpZMjkxYm5RdVkyOXRJaXdLSUNBaVkyeHBaVzUwWDJsa0lqb2dJakV4TmpJek1qTTBORGt4TURRNU56azJORGd6T0NJc0NpQWdJbUYxZEdoZmRYSnBJam9nSW1oMGRIQnpPaTh2WVdOamIzVnVkSE11WjI5dloyeGxMbU52YlM5dkwyOWhkWFJvTWk5aGRYUm9JaXdLSUNBaWRHOXJaVzVmZFhKcElqb2dJbWgwZEhCek9pOHZiMkYxZEdneUxtZHZiMmRzWldGd2FYTXVZMjl0TDNSdmEyVnVJaXdLSUNBaVlYVjBhRjl3Y205MmFXUmxjbDk0TlRBNVgyTmxjblJmZFhKc0lqb2dJbWgwZEhCek9pOHZkM2QzTG1kdmIyZHNaV0Z3YVhNdVkyOXRMMjloZFhSb01pOTJNUzlqWlhKMGN5SXNDaUFnSW1Oc2FXVnVkRjk0TlRBNVgyTmxjblJmZFhKc0lqb2dJbWgwZEhCek9pOHZkM2QzTG1kdmIyZHNaV0Z3YVhNdVkyOXRMM0p2WW05MEwzWXhMMjFsZEdGa1lYUmhMM2cxTURrdlkyeHZkV1J3YkdWNExXNWxlSFJuWlc0dGNIVnphQzFwYldGblpYTWxOREJqYkc5MVpIQnNaWGd0Ym1WNGRHZGxiaTVwWVcwdVozTmxjblpwWTJWaFkyTnZkVzUwTG1OdmJTSUtmUT09IiwiZW1haWwiOiJlbWFpbEBlbWFpbC5jb20ifX0="
                }
            }
        ]
    }

}
`
	r := resty.New()
	r.SetAllowGetMethodPayload(true)
	d, err := r.R().SetBody(payload).SetHeader("Content-Type", "application/json").Get("http://3.1.139.25:30830/api/v1/solution")
	fmt.Println(d, err)
}
