agent {
  // agent id
  id = "localhost-1"

  // public namespace
  public {
    join = [
      "127.0.0.1:7173"
    ]
    bind_addr = "0.0.0.0:7171"
    advertise_addr = "127.0.0.1:7171"

    retry_join = "30s"

  }

  meta {
    "consul" = "true"
    "consul-client" = "true"
    "field" = "all,consul"
  }

  exec = "ExecStart=/usr/bin/sleep inf"
  workers = 4


}

// private pod
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

