# Deploy the demo to Azure

Terraform creates the resource group, Container Apps environment and private image registry. Each API has its own CI and deployment workflows. Go is public; Node is reachable only inside the environment.

Run the following commands from the root of a GitHub checkout of this repository. You need Terraform, Azure CLI and GitHub CLI, plus an Azure account allowed to create resources, service principals and role assignments. Check [Azure pricing](https://azure.microsoft.com/pricing/details/container-apps/) before creating resources; the demo also uses an ACR Basic registry.

## 1. Prepare Azure once

Sign in, copy the example configuration and fill in your subscription ID and a unique registry name using lowercase letters and numbers:

```sh
az login
cp infra/terraform.tfvars.example infra/terraform.tfvars
```

Select the subscription you entered, then create the infrastructure:

```sh
az account set --subscription YOUR_AZURE_SUBSCRIPTION_ID
az provider register --namespace Microsoft.App --wait
az provider register --namespace Microsoft.ContainerRegistry --wait
terraform -chdir=infra init
terraform -chdir=infra apply
```

The default region is `eastus2`. To use another region, add `location = "YOUR_REGION"` to `infra/terraform.tfvars`. Keep the local Terraform state: it contains registry credentials and is excluded from Git.

## 2. Configure GitHub once

Sign in to GitHub and create an Azure deployment identity with Contributor access to this demo's resource group. The credentials go directly into a repository secret:

```sh
gh auth login
set -o pipefail
az ad sp create-for-rbac \
  --name matrix-challenge-github \
  --role Contributor \
  --scopes "$(terraform -chdir=infra output -raw resource_group_id)" \
  --json-auth \
  | gh secret set AZURE_CREDENTIALS
```

Save the registry name and credentials directly from Terraform:

```sh
terraform -chdir=infra output -raw registry_name | gh variable set ACR_NAME
terraform -chdir=infra output -raw registry_username | gh secret set ACR_USERNAME
terraform -chdir=infra output -raw registry_password | gh secret set ACR_PASSWORD
```

These commands configure the repository containing your checkout. The Azure identity's secret expires after one year by default; renew it and update `AZURE_CREDENTIALS` if you keep the demo longer.

## 3. Publish a version

The four workflows appear in GitHub's **Actions** tab:

| Workflow | Trigger | Purpose |
|---|---|---|
| Matrix CI | Pull request or push to `main` affecting Matrix | Test Go |
| Statistics CI | Pull request or push to `main` affecting Statistics | Test Node |
| Deploy Matrix | Push a tag such as `matrix-v1.0.0` | Run Matrix CI, then publish `matrix:1.0.0` |
| Deploy Statistics | Push a tag such as `statistics-v1.0.0` | Run Statistics CI, then publish `statistics:1.0.0` |

A service's folder and its two workflow files trigger its CI. Deployment is triggered by that service's release tag. Both services have independent versions, and pushing code to `main` runs tests without deploying.

Choose a stable version in `MAJOR.MINOR.PATCH` format: increment the patch for compatible fixes, the minor for compatible new functionality, and the major for incompatible API changes. The version covers the service's code, dependencies and Dockerfile. Publish a new version when the shipped contents change, and keep existing tags and image versions. The workflows reject invalid version tags and image versions already present in ACR. They support stable versions such as `1.0.0`; prereleases and build metadata are outside this demo.

From the current `main` branch, publish a Statistics release:

```sh
git switch main
git pull --ff-only
git tag statistics-v1.0.0
git push origin statistics-v1.0.0
```

For the first deployment, wait for **Deploy Statistics** to succeed, then publish Matrix so Go's HTTP dependency is available when you try the demo:

```sh
git tag matrix-v1.0.0
git push origin matrix-v1.0.0
```

For later releases, publish only the affected service's new tag, for example `matrix-v1.0.1`. Choosing the version and pushing its tag are manual; tests, image construction, registry upload and application deployment are automatic.

Each deployment calls its own CI for the tagged commit, then uses the [official Azure action](https://github.com/Azure/container-apps-deploy-action) to build and deploy that service. Failed tests stop that deployment. CI does not receive Azure credentials. If an attempt has already uploaded its image, use a new version instead of rebuilding over that image. The action creates each application the first time and updates it afterward. Run Terraform locally when the infrastructure changes.

After **Deploy Matrix** succeeds, get the public Matrix API hostname:

```sh
az containerapp show \
  --name matrix \
  --resource-group matrix-challenge \
  --query properties.configuration.ingress.fqdn \
  --output tsv
```

Prepend `https://` to that hostname. Use `/health` to check the API, `/docs` for its interactive reference, and send the README's example request to `/api/v1/matrices/rotate`. The first request after inactivity can take longer while the services start.

## Finish the demo

Delete the applications before removing their environment and registry with Terraform:

```sh
az containerapp delete --name matrix --resource-group matrix-challenge --yes
az containerapp delete --name statistics --resource-group matrix-challenge --yes
terraform -chdir=infra destroy
```
