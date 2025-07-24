# Simple URL Shortener

## Objective
Create a Simple URL Shortener and deploy in AWS using Terraform

### Approach
App is written in go (1.24) and uses `gorilla/mux` for implementing the handlers with simple `Basic Authentication`. Supports both `API` and `browser` usage. This is an in-memory implementation and can be extended using a database/persistence layer.

Terraform (flat structure, for simplicity) is used to deploy the code to AWS App Runner. Terraform takes care of building the docker image and pushing it to AWS ECR from where it is pulled for deploying to App Runner.

Terraform docs are available [here](./terraform/README.md)

#### API Endpoints
1. `GET /:short_url`
1. `POST /shorten`


## Running Locally

### Build docker image
`docker build -t url-shortener .`

### Run the APP with defined credentials
```
export AUTH_USER="admin"
export AUTH_PASS="somerandompassword"
docker run -p 8080:8080 url-shortener
```

## Deploying
Terraform will build the docker image and push to the ECR Repository.

The image is then run using AWS App Runner

### Directory Structure
```
├── app
│   └── main.go
├── Dockerfile
├── go.mod
├── go.sum
├── README.md
└── terraform
    ├── main.tf
    ├── outputs.tf
    ├── providers.tf
    ├── README.md
    ├── terraform.tfvars
    └── variables.tf

```

### Steps to Deploy
```
cd terraform
terraform init
terraform plan
terraform apply -var="auth_pass=<some secret password>" --auto-approve
```

At the end of the terraform apply, the public url where the url-shortener can be accessed will be available in the output.
