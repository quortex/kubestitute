# kubestitute

## Overview
This project is an operator allowing Kubernetes to automatically manage the lifecycle of instances within a cluster based on specific events.

This tool is not intended to replace existing tools such as the [cluster-autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) but rather to supplement them in order to have more control and responsiveness over the provisioning of instances in a cluster.

kubestitute only works with clusters deployed on AWS, supporting Autoscaling Groups at the moment.

## Usage
The standard use case for this tool is to provision on-demand fallback instances in case we cannot get spot instances.

To do this, you must configure an autoscaling group of spot instances managed by the cluster-autoscaler and an autoscaling group of on-demand fallback instances managed by kubestitute.

kubestitute will allow us to scale up the on-demand autoscaling group according to events on the spot instances autoscaling group retrieved from the cluster-autoscaler status.
It will also allow us to drain fallback instances and detach them from the autoscaling group according to events (typically when the spot instances have finally been scheduled).

## Prerequisites

### Kubernetes
A Kubernetes cluster of version >=1.16.0 is required. If you are just starting out with kubestitute, it is highly recommended to use the latest version.

### <a id="Prerequisites_AWS"></a>AWS
To be used with AWS and interact with autoscaling groups, an AWS service account with the following permissions on Autoscaling groups managed by kubestitute is required:
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

Only the kubebuilder bootstrap deployment is available at the moment. A helm deployment is coming soon...

1. Create a namespace for kubestitute

```sh
kubectl create ns kubestitute-system
```

2. Create an `aws-ec2-plugin` secret with credentials from AWS account with [necessary permissions](#Prerequisites_AWS).

```sh
kubectl create secret generic aws-ec2-plugin --from-literal=awsKeyId=$AWS_ACCESS_KEY_ID --from-literal=awsSecretKey=$AWS_SECRET_ACCESS_KEY -n kubestitute-system
```

3. Create an `aws-ec2-plugin` configmap with [kubestitute desired configuration](#Configuration_AWS_plugin).

```sh
kubectl apply -f myconfigmap.yaml
```

4. Deploy the appropriate release.

```sh
make deploy IMG=quortexio/quortex-operator:$RELEASE
```


## Configuration

### <a id="Configuration_AWS_plugin"></a>AWS plugin
kubestitute deploy a sidecar to interact with AWS. To configure this plugin, a configmap must be deployed with the desired configuration into `kubestitute-system` namespace.

```yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: aws-ec2-plugin
  namespace: kubestitute-system
data:
  config.yaml: |
    # region is used to interact with AWS API, it must match
    # your cluster's region.
    region: eu-west-1
    # tags are used to define the scope of AWS resources that
    # kubestitute will attempt to access. At least one tag must
    # be specified or the unsafe attribute must be true.
    tags:
      quortex.io/owner/environment: staging
      quortex.io/owner/service: kubestitute
      quortex.io/owner/identifier: my-cluster-name
      # unsafe: true
```

### <a id="Configuration_Optional_args"></a>Optional args
The kubestitute container takes as argument the parameters below.
| Key                                | Description                                                                 | Default                   |
| ---------------------------------- | --------------------------------------------------------------------------- | ------------------------- |
| clusterautoscaler-status-namespace | The namespace the clusterautoscaler status configmap belongs to.            | kube-system               |
| clusterautoscaler-status-name      | The namespace the clusterautoscaler status configmap belongs to.            | cluster-autoscaler-status |
| dev                                | Enable dev mode for logging.                                                | `false`                   |
| v                                  | Logs verbosity. 0 => panic, 1 => error, 2 => warning, 3 => info, 4 => debug | 3                         |


## CustomResourceDefinitions
A core feature of kubestitute is to monitor the Kubernetes API server for changes to specific objects and ensure that the current cluster infrastructure match these objects.

The Operator acts on the following [custom resource definitions (CRDs)](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/):

**`Instance`** defines a desired Instance (only AWS EC2 instances in Autoscaling Groups supported at the moment).

**`Scheduler`** defines a scheduler for Instances (only AWS EC2 instances in Autoscaling Groups supported at the moment). This resource is used to configure advanced instances scheduling based on node groups events.

You can find examples of CRDs defined by kubestitute [here](https://github.com/quortex/kubestitute/tree/feature/documentation/config/samples).

Full API documentation is available [here](./docs/api-docs.asciidoc).

## Supervision

### Logs
By default, kubestitute produces structured logs, with "Info" verbosity. These settings can be configured as described [here](#Configuration_Optional_args).

### Metrics
kubestitute being built from kubebuilder, this one natively exposes a collection of performance metrics for each controller.
kubebuilder documentation about metrics [here](https://book.kubebuilder.io/reference/metrics.html).

We also expose custom metrics as described here:

| Metric name                                    | Metric type | Labels                                     | Description                                                               |
| ---------------------------------------------- | ----------- | ------------------------------------------ | ------------------------------------------------------------------------- |
| kubestitute_scaled_up_nodes_total              | Counter     | `autoscaling_group_name`, `scheduler_name` | Number of nodes added by kubestitute.                                     |
| kubestitute_scaled_down_nodes_total            | Counter     | `autoscaling_group_name`, `scheduler_name` | Number of nodes removed by kubestitute.                                   |
| kubestitute_autoscaling_group_desired_capacity | Gauge       | `autoscaling_group_name`                   | The desired size of the autoscaling group.                                |
| kubestitute_autoscaling_group_capacity         | Gauge       | `autoscaling_group_name`                   | The current autoscaling group capacity (Pending and InService instances). |
| kubestitute_autoscaling_group_min_size         | Gauge       | `autoscaling_group_name`                   | The minimum size of the autoscaling group.                                |
| kubestitute_autoscaling_group_max_size         | Gauge       | `autoscaling_group_name`                   | The maximum size of the autoscaling group.                                |

## License
Distributed under the Apache 2.0 License. See `LICENSE` for more information.

## Versioning
We use [SemVer](http://semver.org/) for versioning.

## Help
Got a question?
File a GitHub [issue](https://github.com/quortex/kubestitute/issues) or send us an [email][email].


  [email]: mailto:info@quortex.io