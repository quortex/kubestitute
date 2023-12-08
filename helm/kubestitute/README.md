# kubestitute

![Version: 2.3.0](https://img.shields.io/badge/Version-2.3.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 2.2.0](https://img.shields.io/badge/AppVersion-2.2.0-informational?style=flat-square)

Kubestitute is an event based instances lifecycle manager for Kubernetes.

**Homepage:** <https://github.com/quortex/kubestitute>

## Source Code

* <https://github.com/quortex/kubestitute/helm>

## Usage
The standard use case for this tool is to provision on-demand fallback instances in case Spot instances cannot be scheduled.

To do so, configure an Auto Scaling Group of Spot instances managed by the Cluster Autoscaler and another one of on-demand fallback instances managed by Kubestitute.

Kubestitute will scale up the on-demand Auto Scaling Group according to events on the Spot instances Auto Scaling Group retrieved from the cluster-autoscaler status.
It will also drain fallback instances and detach them from the Auto Scaling Group according to events (typically when the Spot instances have finally been scheduled).

## Prerequisites

### Kubernetes
A Kubernetes cluster of version v1.11.3+ is required. If you are just starting out with Kubestitute, it is highly recommended to use the latest version.

### <a id="Prerequisites_AWS"></a>AWS
To be used with AWS and interact with Auto Scaling Groups, an AWS IAM user or role with the following permissions on Auto Scaling Groups managed by Kubestitute is required:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllObjectActions",
            "Effect": "Allow",
            "Action": [
                "autoscaling:DescribeAutoScalingGroups",
                "autoscaling:SetDesiredCapacity",
                "autoscaling:TerminateInstanceInAutoScalingGroup"
            ],
            "Resource": "*"
        }
    ]
}
```

## Installation

1. Add Kubestitute helm repository

```sh
helm repo add kubestitute https://quortex.github.io/kubestitute
```

2. Create a namespace for Kubestitute

```sh
kubectl create ns kubestitute-system
```

3. Deploy the appropriate release.

```sh
helm install kubestitute kubestitute/kubestitute -n kubestitute-system
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| aws.region | string | `""` | The region in which the cluster resides. |
| aws.accessKeyID | string | `""` | The access key id of a user with necessary permissions. |
| aws.secretAccessKey | string | `""` | The secret access key of a user with necessary permissions. |
| manager.clusterAutoscaler.namespace | string | `"kube-system"` | The Cluster Autoscaler namespace. |
| manager.clusterAutoscaler.name | string | `"cluster-autoscaler-status"` | The Cluster Autoscaler status configmap name. |
| manager.priorityExpander.enabled | bool | `false` |  |
| manager.priorityExpander.name | string | `"priority-expander-default"` | Name of the Priority Expander object. |
| manager.priorityExpander.namespace | string | `"kubestitute-system"` | Namespace of the Priority Expander object. |
| manager.priorityExpander.clusterAutoscalerConfigMap | string | `"cluster-autoscaler-priority-expander"` | This name should not be changed. This is the exact name cluster autoscaler is looking for. |
| manager.priorityExpander.template | string | `"5:\n - .*"` | Default template, no priorities. |
| manager.logs.verbosity | int | `3` | Logs verbosity:   0 => panic  1 => error  2 => warning  3 => info  4 => debug |
| manager.logs.enableDevLogs | bool | `false` |  |
| manager.asgPollInterval | int | `30` | AutoScaling Groups polling interval (used to generate custom metrics about ASGs). |
| manager.evictionTimeout | int | `300` | The timeout in seconds for pods eviction on Instance deletion. |
| manager.instancesMaxConcurrentReconciles | int | `10` | The maximum number of concurrent Reconciles which can be run for Instances. |
| manager.image.repository | string | `"quortexio/kubestitute"` | Kubestitute manager image repository. |
| manager.image.tag | string | `""` | Kubestitute manager image tag. |
| manager.image.pullPolicy | string | `"IfNotPresent"` | Kubestitute manager image pull policy. |
| manager.livenessProbe.httpGet.path | string | `"/healthz"` | Path of the manager liveness probe. |
| manager.livenessProbe.httpGet.port | int | `8081` | Name or number of the manager liveness probe port. |
| manager.livenessProbe.initialDelaySeconds | int | `15` | Number of seconds before the manager liveness probe is initiated. |
| manager.livenessProbe.periodSeconds | int | `20` | How often (in seconds) to perform the manager liveness probe. |
| manager.readinessProbe.httpGet.path | string | `"/readyz"` | Path of the manager readiness probe. |
| manager.readinessProbe.httpGet.port | int | `8081` | Name or number of the manager readiness probe port. |
| manager.readinessProbe.initialDelaySeconds | int | `5` | Number of seconds before the manager readiness probe is initiated. |
| manager.readinessProbe.periodSeconds | int | `10` | How often (in seconds) to perform the manager readiness probe. |
| manager.resources | object | `{}` | Kubestitute manager container required resources. |
| kubeRBACProxy.enabled | bool | `true` |  |
| kubeRBACProxy.image.repository | string | `"gcr.io/kubebuilder/kube-rbac-proxy"` | kube-rbac-proxy image repository. |
| kubeRBACProxy.image.tag | string | `"v0.8.0"` | kube-rbac-proxy image tag. |
| kubeRBACProxy.image.pullPolicy | string | `"IfNotPresent"` | kube-rbac-proxy image pull policy. |
| kubeRBACProxy.resources | object | `{}` | kube-rbac-proxy container required resources. |
| replicaCount | int | `1` | Number of desired pods. |
| securityContext | object | `{}` | Security contexts to set for all containers of the pod. |
| imagePullSecrets | list | `[]` | A list of secrets used to pull containers images. |
| nameOverride | string | `""` | Helm's name computing override. |
| fullnameOverride | string | `""` | Helm's fullname computing override. |
| podAnnotations | object | `{}` | Annotations to be added to pods. |
| serviceAccount | object | `{"annotations":{}}` | serviceAccount setup |
| deploymentAnnotations | object | `{}` | Annotations to be added to deployment. |
| terminationGracePeriod | int | `30` | How long to wait for pods to stop gracefully. |
| nodeSelector | object | `{}` | Node labels for Kubestitute pod assignment. |
| tolerations | list | `[]` | Node tolerations for Kubestitute scheduling to nodes with taints. |
| affinity | object | `{}` | Affinity for Kubestitute pod assignment. |
| serviceMonitor.enabled | bool | `false` | Create a prometheus operator ServiceMonitor. |
| serviceMonitor.additionalLabels | object | `{}` | Labels added to the ServiceMonitor. |
| serviceMonitor.annotations | object | `{}` | Annotations added to the ServiceMonitor. |
| serviceMonitor.interval | string | `""` | Override prometheus operator scrapping interval. |
| serviceMonitor.scrapeTimeout | string | `""` | Override prometheus operator scrapping timeout. |
| serviceMonitor.relabelings | list | `[]` | Relabellings to apply to samples before scraping. |

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| vincentmrg |  | <https://github.com/vincentmrg> |
