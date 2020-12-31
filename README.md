# ecr-credential-updater
K3s & `containerd` don't support dockers `ecr-credential-helper` for automatic retrival and configuration of ECR credentials.

This tool will automatically generate ECR credentials and update a predefined secret that can be used as an `imagePullSecret`.

Example `Kubernetes` manifests:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ecr-credential-updater
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ecr-credential-updater
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ecr-credential-updater
subjects:
- kind: ServiceAccount
  name: ecr-credential-updater
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: ecr-credential-updater
  namespace: default
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - create
  - update
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ecr-credential-updater
  namespace: default
spec:
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: ecr-credential-updater
  template:
    metadata:
      labels:
        app: ecr-credential-updater
    spec:
      serviceAccountName: ecr-credential-updater
      containers:
      - name: ecr-credential-updater
        image: public.ecr.aws/alexlast/ecr-credential-updater:latest
        env:
        - name: ECR_REGISTRY_REGION
          value: eu-west-2
        - name: ECR_REGISTRY
          value: 12345678910.dkr.ecr.eu-west-2.amazonaws.com
        - name: ECR_SECRET_NAME
          value: ecr-credentials
        - name: ECR_SECRET_NAMESPACE
          value: tasks
        - name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              name: ecr-credential-updater
              key: AWS_ACCESS_KEY_ID
        - name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: ecr-credential-updater
              key: AWS_SECRET_ACCESS_KEY
```