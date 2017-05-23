agent {
  // agent id
  id = "localhost-1"

  serf {
    join = [
      "127.0.0.1:7171",
      "127.0.0.1:7172",
      "127.0.0.1:7173"
    ]
    bind = "0.0.0.0:7171"
    advertise = "127.0.0.1:7171"
    retry = "30s"
  }

  rpc {
    bind = "0.0.0.0:7181"
    advertise = "127.0.0.1:7181"
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

