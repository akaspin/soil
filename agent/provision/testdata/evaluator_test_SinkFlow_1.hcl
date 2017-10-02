pod "third" {
  constraint {
    "${meta.third_public}" = "1"
  }
  unit "third-1.service" {
    create = "start"
    source = <<EOF
    [Unit]
    Description=%p public

    [Service]
    ExecStart=/usr/bin/sleep inf

    [Install]
    WantedBy=default.target
  EOF
  }
}
