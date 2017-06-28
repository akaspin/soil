pod "first" {
  runtime = true
  target = "multi-user.target"

  constraint {
    "${meta.one}" = "one"
    "${meta.two}" = "two"
  }

  blob "/etc/vpn/users/env" {
    permissions = "0644"
    leave = false
    source = <<EOF
      My file
    EOF
  }

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
    create = ""
    update ="start"
    destroy = ""

    source = <<EOF
[Service]
# ${NONEXISTENT}
ExecStart=/usr/bin/sleep inf
EOF
  }
}