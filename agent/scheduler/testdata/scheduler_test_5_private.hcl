pod "first" {
  constraint {
    "${meta.first_private}" = "1"
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

pod "second" {
  constraint {
    "${meta.second_private}" = "1"
  }
  unit "second-1.service" {
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

