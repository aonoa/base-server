# deploy/helm/base-server/ — Helm chart

## Layout
- `Chart.yaml` / `values.yaml`
- `templates/deployment.yaml`, `templates/service.yaml`, `templates/configmap.yaml`, `templates/_helpers.tpl`

## Render / install
- `helm template base-server ./deploy/helm/base-server`
- `helm install base-server ./deploy/helm/base-server`
- `helm upgrade base-server ./deploy/helm/base-server`

## Runtime config mounts
- `config.yaml` → mounted at `/data/conf/config.yaml`
- casbin model/policy → mounted under `/app/authconf/`

## Notable chart behaviors
- `templates/configmap.yaml` embeds:
  - `config.yaml` (app config)
  - `keymatch_model.conf` + `keymatch_policy.csv` (casbin)
- NodePort `30080` is fixed in the service template.
- Credentials/keys are hardcoded in ConfigMap templates; treat as placeholders.
