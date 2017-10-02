pod "pod-1" {
  runtime = true
  unit "unit-1.service" {
    permanent = true
    source = "fake"
  }
  unit "unit-2.service" {
    permanent = true
    source = "fake"
  }
  blob "/etc/test1" {
    source = "test"
  }
}