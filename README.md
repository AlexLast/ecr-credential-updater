# ecr-credential-updater
K3s & `containerd` don't support dockers `ecr-credential-helper` for automatic retrival and configuration of ECR credentials.

This tool will automatically generate ECR credentials and update a predefined secret that can be used as an `imagePullSecret`.

Example `Deployment`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ecr-credential-updater
  namespace: tasks
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
      containers:
      - name: ecr-credential-updater
        image: public.ecr.aws/alexlast/ecr-credential-updater:latest
        env:
        - name: ECR_REGISTRY_REGION
          value: eu-west-2
        - name: ECR_REGISTRY
          value: 123456789.dkr.ecr.eu-west-2.amazonaws.com
        - name: ECR_SECRET_NAME
          value: ecr-credentials
        - name: ECR_SECRET_NAMESPACE
          value: default
```