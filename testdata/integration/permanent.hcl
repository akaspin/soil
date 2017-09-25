meta {
  "test" = "a"
}

pod "embedded-1" {
  runtime = false
  constraint {
    "${meta.test}" = "a"
  }
  unit "embedded-1-1.service" {
    permanent = true
    source = <<EOF
[Unit]
Description=%p

[Service]
# ${NONEXISTENT}
ExecStart=/usr/bin/sleep inf

[Install]
WantedBy=multi-user.target
EOF
  }
}