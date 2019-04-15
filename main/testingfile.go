package main

import (
	"fmt"
	"gopkg.in/resty.v1"
)

func main() {
	payload := `
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
