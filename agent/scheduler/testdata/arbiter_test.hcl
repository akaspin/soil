pod "first" {
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


pod "second" {
  constraint {
    "${meta.second}" = "1"
    "${with.dot.first}" = "1"
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
