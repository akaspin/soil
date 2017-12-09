pod "pod-1" {
  runtime = true
  namespace = "private"
  target = "multi-user.target"

  constraint {
    "${aaa.bbb}" = "true"
  }

  provider "range" "port" {
    min = 900
    max = 2000
  }

  resource "pod-1.port" "8080" {
    fixed = 8080
  }
  resource "global.counter" "main" {
    count = 3
  }
}
