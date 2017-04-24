agent {
  fake = "string"
}

pod "first" {
  count = 2
  runtime = true
  target = "multi-user.target"

  // Count
  unit "first-1.service" {
    permanent = true
    create = "start"
    update = ""
    destroy = "stop"
    source = <<EOF
      [Service]
      # ${meta.consul}
      ExecStart=/usr/bin/sleep inf
      ExecStopPost=/usr/bin/systemctl stop first-2.service
    EOF
  }
  unit "first-2.service" {
    update ="start"
    destroy = ""

    source = <<EOF
[Service]
# ${NONEXISTENT}
ExecStart=/usr/bin/sleep inf
EOF
  }
}

pod "second" {
  runtime = false
  constraint {
    "${meta.consul}" = "true"
  }

  unit "second-1.service" {
    source = <<EOF
    [Service]
    ExecStart=/usr/bin/sleep inf
    EOF
  }
}