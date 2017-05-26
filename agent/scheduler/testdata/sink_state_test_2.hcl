pod "pod-1" {
  constraint {
    "${meta.first}" = "1"
  }
  unit "first-1.service" {
    create = "start"
    source = <<EOF
    [Unit]
    Description=%p

    [Service]
    ExecStart=/usr/bin/sleep inf

    [Install]
    WantedBy=default.target
  EOF
  }
}

pod "pod-3" {
  constraint {
    "${meta.first}" = "1"
  }
  unit "first-1.service" {
    create = "start"
    source = <<EOF
    [Unit]
    Description=%p

    [Service]
    ExecStart=/usr/bin/sleep inf

    [Install]
    WantedBy=default.target
  EOF
  }
}