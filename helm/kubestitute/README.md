# kubestitute

![Version: 0.2.0](https://img.shields.io/badge/Version-0.2.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.2.0](https://img.shields.io/badge/AppVersion-0.2.0-informational?style=flat-square)

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
To be used with AWS and interact with Auto Scaling Groups, an AWS service account with the following permissions on Auto Scaling Groups managed by Kubestitute is required:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllObjectActions",
            "Effect": "Allow",
            "Action": [
                "ec2:TerminateInstances",
                "autoscaling:DescribeAutoScalingGroups",
                "autoscaling:SetDesiredCapacity",
                "autoscaling:DetachInstances"
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

3. Create a secret (`aws-ec2-plugin` by default) with credentials from AWS account with [necessary permissions](#Prerequisites_AWS).

```sh
kubectl create secret generic aws-ec2-plugin --from-literal=awsKeyId=$AWS_ACCESS_KEY_ID --from-literal=awsSecretKey=$AWS_SECRET_ACCESS_KEY -n kubestitute-system
```

4. Deploy the appropriate release.

```sh
helm install kubestitute kubestitute/kubestitute -n kubestitute-system
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| manager.clusterAutoscaler.namespace | string | `"kube-system"` | The Cluster Autoscaler namespace. |
| manager.clusterAutoscaler.name | string | `"cluster-autoscaler-status"` | The Cluster Autoscaler status configmap name. |
| manager.logs.verbosity | int | `3` | Logs verbosity:  0 => panic  1 => error  2 => warning  3 => info  4 => debug |
| manager.logs.enableDevLogs | bool | `false` |  |
| manager.asgPollInterval | int | `30` | AutoScaling Groups polling interval (used to generate custom metrics about ASGs). |
| manager.evictionTimeout | int | `300` | The timeout in seconds for pods eviction on Instance deletion. |
| manager.image.repository | string | `"quortexio/kubestitute"` | Kubestitute manager image repository. |
| manager.image.tag | string | `"0.1.1"` | Kubestitute manager image tag. |
| manager.image.pullPolicy | string | `"IfNotPresent"` | Kubestitute manager image pull policy. |
| manager.resources | object | `{}` | Kubestitute manager container required resources. |
| awsEC2Plugin.enabled | bool | `true` | Wether to enable AWS EC2 plugin. |
| awsEC2Plugin.secret | string | `"aws-ec2-plugin"` | A reference to a secret wit AWS credentials for AWS EC2 plugin. |
| awsEC2Plugin.region | string | `""` | The AWS region. |
| awsEC2Plugin.tags | object | `{}` | Tags for AWS EC2 plugin scope management. |
| awsEC2Plugin.image.repository | string | `"quortexio/aws-ec2-adapter"` | AWS EC2 plugin image pull policy. |
| awsEC2Plugin.image.tag | string | `"1.1.0"` | AWS EC2 plugin image pull policy. |
| awsEC2Plugin.image.pullPolicy | string | `"IfNotPresent"` | AWS EC2 plugin image pull policy. |
| awsEC2Plugin.resources | object | `{}` | AWS EC2 plugin container required resources. |
| kubeRBACProxy.enabled | bool | `true` |  |
| kubeRBACProxy.image.repository | string | `"gcr.io/kubebuilder/kube-rbac-proxy"` | kube-rbac-proxy image repository. |
| kubeRBACProxy.image.tag | string | `"v0.5.0"` | kube-rbac-proxy image tag. |
| kubeRBACProxy.image.pullPolicy | string | `"IfNotPresent"` | kube-rbac-proxy image pull policy. |
| kubeRBACProxy.resources | object | `{}` | kube-rbac-proxy container required resources. |
| replicaCount | int | `1` | Number of desired pods. |
| imagePullSecrets | list | `[]` | A list of secrets used to pull containers images. |
| nameOverride | string | `""` | Helm's name computing override. |
| fullnameOverride | string | `""` | Helm's fullname computing override. |
| podAnnotations | object | `{}` | Annotations to be added to pods. |
| deploymentAnnotations | object | `{}` | Annotations to be added to deployment. |
| terminationGracePeriod | int | `30` | How long to wait for pods to stop gracefully. |
| nodeSelector | object | `{}` | Node labels for Kubestitute pod assignment. |
| tolerations | list | `[]` | Node tolerations for Kubestitute scheduling to nodes with taints. |
| affinity | object | `{}` | Affinity for Kubestitute pod assignment. |

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| vincentmrg |  | https://github.com/vincentmrg |
