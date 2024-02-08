# argo-rollout-generator

Generating argo rollout objects with the native Kubernetes golang packages...

### Argo rollouts CRD bug w/ k8s strict server side validation

When using the rolloutsv1a1.Rollout{} struct to define a rollout object programmatically, I noticed that the pod template spec within the rollout spec has a creationTimestamp field with a value of null. This showed up in my generated YAML file and resulted in the following error when applying the manifest into a Kubernetes 1.26 cluster:

`Error from server (BadRequest): error when creating "manifest.yaml": Rollout in version "v1alpha1" cannot be handled as a Rollout: strict decoding error: unknown field "spec.template.metadata.creationTimestamp"`

YAML manifest:

```
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  creationTimestamp: null
  labels:
    bear-type: polar-bear
  name: dancing-bears
  namespace: bear-system
spec:
  replicas: 3
  revisionHistoryLimit: 0
  selector:
    matchLabels:
      bear-type: polar-bear
  strategy:
    blueGreen:
      activeService: dancing-bears-svc
      autoPromotionEnabled: true
      previewService: dancing-bears-preview-svc
      scaleDownDelaySeconds: 60
  template:
    metadata:
      creationTimestamp: null #causes validation error!
      labels:
        bear-type: polar-bear
    spec:
      containers:
      - image: ghcr.io/nickpatton/some-dancing-bears:v0.0.4
        imagePullPolicy: IfNotPresent
        name: dancing-bears
        ports:
        - containerPort: 5000
          name: app
        resources: {}
      restartPolicy: Always
status:
  blueGreen: {}
  canary: {}
```

I think `creationTimestamp: null` is showing up because the Kubernetes v1.ObjectMeta struct has creationTimestamp as a Time type instead of a pointer to a Time type. This seems to render the 'omitempty' JSON tag useless.

### How do I produce the error?

Execute the code inside main.go by running `go run .` This will generate a YAML file, `manifest.yaml`. This manifest will contain a rollout definition as well as two service objects for the blue/green functionality. Run `kubectl apply -f manifest.yaml` to create the resources in a cluster. With the argo rollouts controller/CRDs installed, this should generate the strict decoding error in a Kubernetes cluster running 1.25 or greater.

Ref: https://kubernetes.io/blog/2023/04/24/openapi-v3-field-validation-ga/

### How do we fix this?

It could be fixed by modifying the Rollout CRD to have a creationTimestamp field within metadata's properties:

https://github.com/argoproj/argo-rollouts/blob/master/manifests/crds/rollout-crd.yaml#L944

Or perhaps the issue could be fixed inside Kubernetes v1.ObjectMeta struct by changing creationTimestamp to a pointer. Here's a relevant GitHub issue for this:

https://github.com/kubernetes/kubernetes/issues/67610

The issue could also be worked-around by adding the `--validate=warn` flag to the kubectl apply command, however this doesn't seem like an ideal solution.
