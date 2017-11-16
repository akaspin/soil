agent {
  id = "node-1"
}

cluster {
  advertise = "127.0.0.1:7654"
  backend = "consul://127.0.0.1:8500"
  ttl = "1m"
  retry = "30s"
}
