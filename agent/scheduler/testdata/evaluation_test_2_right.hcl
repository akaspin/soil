pod "pod-1" {
  runtime = false
  unit "unit-1.service" {
    permanent = true
    source = "fake1"
  }
  unit "unit-2.service" {
    permanent = true
    source = "fake"
  }
  blob "/etc/test1" {
    source = "test1"
  }
}