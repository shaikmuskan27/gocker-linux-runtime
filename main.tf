terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0.1" # Use a stable version
    }
  }
}

provider "docker" {
  # No extra config needed, let it find the default pipe
}

resource "docker_image" "logistics_image" {
  name = "logistics-app-local"
  build {
    context = "."
    # This helps avoid 'legacy build' issues
    remove  = true 
  }
}

resource "docker_container" "logistics_server" {
  name  = "logistics-app"
  image = docker_image.logistics_image.image_id
  
  # 1. Change the command to keep the container alive with a shell
  command = ["run", "/bin/sh"]
  
  # 2. Tell Terraform it's okay if the container doesn't stay running forever
  must_run = false 
  
  privileged = true
  stdin_open = true
  tty        = true
}