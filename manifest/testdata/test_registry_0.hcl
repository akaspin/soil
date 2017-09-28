pod "second" {
  runtime = false
  constraint {
    "${meta.consul}" = "true"
  }

  resource "port" "8080" {
    fixed = "8080"
  }

  resource "counter" "1" {
    count = "3"
  }
  resource "counter" "2" {
    required = false
    count = "1"
    a = "b"
  }

  unit "second-1.service" {
    source = <<EOF
    [Service]
    ExecStart=/usr/bin/sleep inf
    EOF
  }
}