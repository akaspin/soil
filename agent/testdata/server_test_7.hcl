meta {
  "1" = "true"
  "2" = "true"
}

resource "range" "port" {
  min = 8000
  max = 9000
}

pod "1" {
  constraint {
    "${meta.1}" = "true"
  }
  resource "port" "8080" {}
  unit "unit-1.service" {
    source = <<EOF
[Service]
# ${resource.port.1.8080.value}
ExecStart=/usr/bin/sleep inf
EOF
  }
}
