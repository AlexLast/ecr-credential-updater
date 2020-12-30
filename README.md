# ecr-credential-updater
K3s & `containerd` don't support dockers `ecr-credential-helper` for automatic retrival and configuration of ECR credentials.

This tool will automatically generate ECR credentials and update a predefined secret that can be used as an `imagePullSecret`.

Ideally this should run as a `Deployment` in any namespace you wish to update an `imagePullSecret` in. It will also need a service account with a role allowing create & update on secrets.

Required environment variables:

- `ECR_REGISTRY` - Registry host
- `ECR_REGISTRY_REGION` - Registry region
- `ECR_SECRET_NAME` - Name of the secret to create or update
- `ECR_SECRET_NAMESPACE` - Namespace where the above secret lives

View the releases tab if you want to use a tagged release, otherwise you can use latest: `public.ecr.aws/alexlast/ecr-credential-updater:latest`
