pod "pod-2" {
  runtime = true
  namespace = "private"
  target = "multi-user.target"

  unit "${pod.name}-unit-1.service" {
    create = "start"
    update = ""
    destroy = "stop"
    source = "# ${meta.consul} ${pod.target}"
  }

  unit "${pod.namespace}-unit-2.service" {
    create = "start"
    update = ""
    destroy = "stop"
    source = "# ${meta.consul} ${blob.pod-2-etc-test}"
  }

  blob "/${pod.name}/etc/test" {
    source = "test"
  }
}
