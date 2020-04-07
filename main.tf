resource "null_resource" "test" {
  provisioner "gcp-rest-api" {
    url    = "https://storage.googleapis.com/storage/v1/b/${var.bucket_name}/o"
    method = "GET"
  }
}

variable "bucket_name" {
  description = "Bucket to list"
}