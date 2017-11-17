cluster {
  node_id = "node-1"
  advertise = "127.0.0.1:7654"
  backend = "consul://127.0.0.1:8500"
  ttl = "10m"
  retry = "30s"
}

cluster {
  node_id = "node-1-add"
  backend = "consul://127.0.0.1:8500"
  ttl = "11m"
}
