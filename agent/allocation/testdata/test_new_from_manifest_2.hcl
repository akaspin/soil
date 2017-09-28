pod "pod-1" {
  runtime = true
  namespace = "private"
  target = "multi-user.target"

  resource "port" "8080" {
    fixed = 8080
  }
  resource "counter" "main" {
    count = 3
  }
}
