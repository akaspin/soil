meta {
  "1" = "true"
  "2" = "true"
}

pod "1" {
  constraint {
    "${meta.1}" = "true"
  }
  unit "unit-1.service" {
    source = <<EOF
[Service]
# ${meta.2}
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
EOF
  }
}

pod "2" {
  runtime = false
  constraint {
    "${meta.2}" = "true"
  }
  unit "unit-2.service" {
    permanent = true
    source = <<EOF
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
EOF
  }
}
