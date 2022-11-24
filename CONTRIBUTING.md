# How to contribute

This project is maintained with **golang v1.19** and **kubebuilder v3.7.0**, please use these versions to ensure the integrity of the project.

## Introduction to Kubebuilder

This operator is build upon [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), a framework for building Kubernetes APIs using [custom resource definitions (CRDs)](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/).

Kubebuilder is developed on top of the controller-runtime and controller-tools libraries.

- [Quickstart](https://book.kubebuilder.io/quick-start.html)
- [Installation instructions](https://book.kubebuilder.io/quick-start.html#installation)

Useful additional resources :

- A [tutorial](https://book.kubebuilder.io/cronjob-tutorial/cronjob-tutorial.html).
- A list of [best practices](https://www.openshift.com/blog/kubernetes-operators-best-practices) for operators development.
- A complete [ebook](https://book.kubebuilder.io/introduction.html).

## How to create a new CRD / API

### 1. Bootstrap files

To create a new API, you need first to bootstrap it with Kubebuilder.
Here is an example to create an object `Frigate` in the group `ship` in version `v1beta1` :

```bash
kubebuilder create api --group ship --version v1beta1 --kind Frigate
```

Versions numbers must follows [kubernetes naming scheme](https://kubernetes.io/docs/concepts/overview/kubernetes-api/#api-versioning). Object are namespaced by default, to create a cluster-wide object add the parameter `--namespaced=false`.

### 2. Define API and implement controller

The next step is to complet API declaration and controller. For a multigroup project like this one :

- The API is in `apis/ship/v1beta1`
- The controller is in `controllers/ship/v1beta1`

Refer to the kubebuilder [API documentation](https://book.kubebuilder.io/cronjob-tutorial/new-api.html) and [controller documentation](https://book.kubebuilder.io/cronjob-tutorial/controller-overview.html) for advanced documentation on these parts.

After modifying files, don't forget to run `make` in order to update generated files, format and lint everything.

### 3. Create a validating or mutating Admission Webhook (Optional)

To create mutating (defaulting) or / and validating webhooks for your API you need to bootstrap it with Kubebuilder.
Here is an example to create defaulting and validation webhooks for `Frigate` in the group `ship` in version `v1beta1` :

```bash
kubebuilder create webhook --group ship --version v1beta1 --kind Frigate --defaulting --programmatic-validation
```

The next step is to fill generated code following [kubebuilder documentation](https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html)

### 4. Testing

You can then run your operator in different ways :

#### Locally

Assuming you have a cluster connected, you can install the _CustomResourceDefinition_ by running `make install` (which can also be removed with `make uninstall`).

```sh
make run ENABLE_WEBHOOKS=false
```

#### With kind

You can test on a local cluster with [kind](https://kind.sigs.k8s.io/), this method will allow you to build an image of your operator and run it on a local cluster. This way you can test more like a real deployment and you can use webhooks.

Create a cluster :

`kind create cluster`

Build the operator locally e.g. :

`make docker-build IMG=quortexio/kubestitute:dev`

Load your image in kind nodes :

`kind load docker-image quortexio/kubestitute:dev`

Deploy the operator in the cluster :

`make deploy IMG=quortexio/kubestitute:dev`

## Contribution rules

- As this project uses kubebuilder 3.7.0, you should use [golang](https://go.dev/) v1.19.
- It is advised to use vendoring to avoid using packages not in the `go mod` or mixing versions.
  To do so, run the command `go mod vendor`.
- Before commiting, run `make` to generate files and format files. You should also run `make lint` to do the same linting done in CI, and `go mod tidy` to tidy up the `go.mod` and `go.sum` files.

## Versioning

We use [SemVer](http://semver.org/) for versioning.
