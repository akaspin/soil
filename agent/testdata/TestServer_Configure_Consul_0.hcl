cluster {
  node_id = "node"
  backend = "consul://{{ .ConsulAddress }}/soil"
  advertise = "{{ .AgentAddress }}"
  ttl = "5s"
  retry = "1s"
}

pod "1" {
  unit "unit-1.service" {
    source = <<EOF
[Service]
ExecStart=/usr/bin/sleep inf
EOF
  }
}
