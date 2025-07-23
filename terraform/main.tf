# Create ECR Repository
resource "aws_ecr_repository" "repo" {
  name                 = var.app_name
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

# Build docker image
resource "docker_image" "this" {
  name = format("%v:%v", aws_ecr_repository.repo.repository_url, "latest")

  build { context = "../" } # Path to the local Dockerfile
}

# Push the container image to ECR.
resource "docker_registry_image" "this" {
  keep_remotely = true # Do not delete old images
  name          = resource.docker_image.this.name
}

# AppRunner Service
resource "aws_apprunner_service" "service" {
  service_name = var.app_name

  source_configuration {
    image_repository {
      image_identifier      = "${aws_ecr_repository.repo.repository_url}:latest"
      image_repository_type = "ECR"
      image_configuration {
        port = var.container_port
        runtime_environment_variables = {
          PORT      = var.container_port
          AUTH_USER = var.auth_user
          AUTH_PASS = var.auth_pass
        }
      }
    }

    authentication_configuration {
      access_role_arn = aws_iam_role.apprunner_role.arn
    }

    auto_deployments_enabled = true
  }

  instance_configuration {
    cpu    = "1 vCPU"
    memory = "2 GB"
  }

  #   depends_on = [aws_ecr_repository.repo]
  depends_on = [docker_registry_image.this]

}


# Roles and Policies
resource "aws_iam_role_policy_attachment" "ecr_read" {
  role       = aws_iam_role.apprunner_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role" "apprunner_role" {
  name = "${var.app_name}-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Action = "sts:AssumeRole",
      Effect = "Allow",
      Principal = {
        Service = "build.apprunner.amazonaws.com"
      }
    }]
  })
}
