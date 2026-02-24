# Tempconv (Go + gRPC + Protocol Buffers + Kubernetes on GCP)

Temperature conversion microservice built with:

- Go
- gRPC
- Protocol Buffers
- Docker
- Kubernetes (GKE)

## 1) Local setup

### Prerequisites

- Go 1.26+
- `protoc`
- `protoc-gen-go`
- `protoc-gen-go-grpc`
- Docker Desktop
- `kubectl`
- Google Cloud SDK (`gcloud`)

### Generate protobuf code

```powershell
protoc --proto_path=proto --go_out=. --go_opt=module=tempconv --go-grpc_out=. --go-grpc_opt=module=tempconv proto/tempconv.proto
```

### Install Go dependencies

```powershell
go mod tidy
```

### Run server locally

```powershell
go run ./cmd/server
```

In another terminal, call it with the Go client:

```powershell
go run ./cmd/client --value 100 --from C --to F
```

Expected output includes converted value `212`.

## 2) Push to GitHub (public repo)

```powershell
git init
git add .
git commit -m "Initial Go Tempconv gRPC project"
git branch -M main
git remote add origin https://github.com/<YOUR_USERNAME>/tempconv-grpc.git
git push -u origin main
```

Then make sure repository visibility is **Public** in GitHub settings.

## 3) Build and push image to Google Artifact Registry

Set variables:

```powershell
$PROJECT_ID="your-gcp-project-id"
$REGION="us-central1"
$REPO="tempconv-repo"
$IMAGE="tempconv"
```

Authenticate and enable APIs:

```powershell
gcloud auth login
gcloud config set project $PROJECT_ID
gcloud services enable container.googleapis.com artifactregistry.googleapis.com
```

Create Artifact Registry repo (one-time):

```powershell
gcloud artifacts repositories create $REPO --repository-format=docker --location=$REGION --description="Tempconv Docker repo"
```

Configure Docker auth and push image:

```powershell
gcloud auth configure-docker "$REGION-docker.pkg.dev"
docker build -t "$REGION-docker.pkg.dev/$PROJECT_ID/$REPO/$IMAGE:latest" .
docker push "$REGION-docker.pkg.dev/$PROJECT_ID/$REPO/$IMAGE:latest"
```

## 4) Deploy to GKE (Kubernetes)

Create a GKE Autopilot cluster:

```powershell
gcloud container clusters create-auto tempconv-cluster --region $REGION
gcloud container clusters get-credentials tempconv-cluster --region $REGION
```

Update image path in `k8s/deployment.yaml`:

```yaml
image: REGION-docker.pkg.dev/PROJECT_ID/tempconv-repo/tempconv:latest
```

Replace `REGION` and `PROJECT_ID`, then deploy:

```powershell
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl get pods
kubectl get svc tempconv-grpc-service
```

Wait until `EXTERNAL-IP` appears.

## 5) Test public endpoint

Use `grpcurl`:

```powershell
grpcurl -plaintext -d '{"value":100,"fromUnit":"CELSIUS","toUnit":"FAHRENHEIT"}' <EXTERNAL-IP>:80 tempconv.TempConverter/ConvertTemperature
```

Expected response:

```json
{
  "convertedValue": 212,
  "formulaUsed": "F = (C × 9/5) + 32"
}
```

## 6) What to submit to your professor

- Public GitHub repository URL
- Public GKE service endpoint (`<EXTERNAL-IP>:80`)
- One successful `grpcurl` output/screenshot

## Notes

- Service is insecure gRPC (`-plaintext`) for class/demo usage.
- For production, add TLS/auth and use Ingress or Gateway.