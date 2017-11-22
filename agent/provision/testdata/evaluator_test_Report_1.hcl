pod "pod-1" {
  unit "unit-1.service" {
    source = <<EOF
[Unit]
Description=Unit 1
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=default.target
EOF
  }
}
