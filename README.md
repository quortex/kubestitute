# Kubestitute

## Overview
This project is an operator allowing Kubernetes to automatically manage the lifecycle of instances within a cluster based on specific events.

This tool is not intended to replace existing tools such as the [Cluster Autoscaler](https://github.com/kubernetes/autoscaler) but rather to supplement them in order to have more control and responsiveness over the provisioning of instances in a cluster.

Kubestitute only works with clusters deployed on AWS using Auto Scaling Groups at the moment.

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

### Helm

Follow Kubestitute documentation for Helm deployment [here](./helm/kubestitute).

### Kubebuilder (development)
Only the Kubebuilder bootstrap deployment is available at the moment. A Helm deployment is coming soon...

1. Create a namespace for Kubestitute

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
Kubestitute deploy a sidecar to interact with AWS. To configure this plugin, a configmap must be deployed with the desired configuration into `kubestitute-system` namespace.

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
| Key                                | Description                                                                       | Default                   |
| ---------------------------------- | --------------------------------------------------------------------------------- | ------------------------- |
| clusterautoscaler-status-namespace | The namespace the clusterautoscaler status configmap belongs to.                  | kube-system               |
| clusterautoscaler-status-name      | The namespace the clusterautoscaler status configmap belongs to.                  | cluster-autoscaler-status |
| dev                                | Enable dev mode for logging.                                                      | `false`                   |
| v                                  | Logs verbosity. 0 => panic, 1 => error, 2 => warning, 3 => info, 4 => debug       | 3                         |
| asg-poll-interval                  | AutoScaling Groups polling interval (used to generate custom metrics about ASGs). | 30                        |


## CustomResourceDefinitions
A core feature of Kubestitute is to monitor the Kubernetes API server for changes to specific objects and ensure that the current cluster infrastructure match these objects.

The Operator acts on the following [custom resource definitions (CRDs)](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/):

**`Instance`** defines a desired Instance (only AWS EC2 instances in Auto Scaling Groups supported at the moment).

**`Scheduler`** defines a scheduler for Instances (only AWS EC2 instances in Auto Scaling Groups supported at the moment). This resource is used to configure advanced instances scheduling based on node groups events.

You can find examples of CRDs defined by Kubestitute [here](./config/samples).

Full API documentation is available [here](./docs/api-docs.asciidoc).

## Supervision

### Logs
By default, Kubestitute produces structured logs, with "Info" verbosity. These settings can be configured as described [here](#Configuration_Optional_args).

### Metrics
Kubestitute being built from Kubebuilder, it natively exposes a collection of performance metrics for each controller. 
Kubebuilder documentation about metrics can be found [here](https://book.kubebuilder.io/reference/metrics.html).

We also expose custom metrics as described here:

All the metrics are prefixed with kubestitute_

| Metric name                        | Metric type | Labels                                                  | Description                                                               |
| ---------------------------------- | ----------- | ------------------------------------------------------- | ------------------------------------------------------------------------- |
| scaled_up_nodes_total              | Counter     | `autoscaling_group_name`, `scheduler_name`              | Number of nodes added by kubestitute.                                     |
| scaled_down_nodes_total            | Counter     | `autoscaling_group_name`, `scheduler_name`              | Number of nodes removed by kubestitute.                                   |
| evicted_pods_total                 | Counter     | `autoscaling_group_name`, `node_name`, `scheduler_name` | Number of pods evicted by kubestitute.                                    |
| autoscaling_group_desired_capacity | Gauge       | `autoscaling_group_name`                                | The desired size of the autoscaling group.                                |
| autoscaling_group_capacity         | Gauge       | `autoscaling_group_name`                                | The current autoscaling group capacity (Pending and InService instances). |
| autoscaling_group_min_size         | Gauge       | `autoscaling_group_name`                                | The minimum size of the autoscaling group.                                |
| autoscaling_group_max_size         | Gauge       | `autoscaling_group_name`                                | The maximum size of the autoscaling group.                                |


## License
Distributed under the Apache 2.0 License. See `LICENSE` for more information.

## Versioning
We use [SemVer](http://semver.org/) for versioning.

## Help
Got a question?
File a GitHub [issue](https://github.com/quortex/kubestitute/issues) or send us an [email][email].


  [email]: mailto:info@quortex.io
