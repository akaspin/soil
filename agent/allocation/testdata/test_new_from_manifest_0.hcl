pod "pod-1" {
  runtime = true
  namespace = "private"
  target = "multi-user.target"

  unit "unit-1.service" {
    create = "start"
    update = ""
    destroy = "stop"
    source = "# ${meta.consul}"
  }

  unit "unit-2.service" {
    create = "start"
    update = ""
    destroy = "stop"
    source = "# ${meta.consul} ${blob.etc-test}"
  }

  blob "/etc/test" {
    source = "test"
  }
}
