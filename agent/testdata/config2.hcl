meta {
  "override" = "true"
}

exec = "ExecStart=/usr/bin/sleep inf"

// private pod 2
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
