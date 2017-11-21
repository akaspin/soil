meta {
  "1" = "true"
  "2" = "true"
}

pod "1" {
  constraint {
    "${meta.1}" = "false"
  }
  unit "unit-1.service" {
    source = <<EOF
[Service]
ExecStart=/usr/bin/sleep inf
EOF
  }
}

pod "2" {
  constraint {
    "${meta.2}" = "true"
    "${provision.1.present}" = "true"
    "${provision.1.state}" = "!= destroy"
  }
  unit "unit-2.service" {
    source = <<EOF
[Service]
ExecStart=/usr/bin/sleep inf
EOF
  }
}
