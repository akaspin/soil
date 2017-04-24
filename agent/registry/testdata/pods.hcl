pod "one-1" {
  runtime = true
  count = 1
  constraint {
    "${meta.consul}" = "true"
  }
  unit "one-1-0.service" {
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

pod "one-2" {
  runtime = false
  constraint {
    "${meta.override}" = "true"
  }
  unit "one-2-0.service" {
    create = "start"
    update = "restart"
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
