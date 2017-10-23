resource "dummy" "1" {
  conf_one = 1
  conf_two = "two"
}

resource "range" "port" {
  min = 8000
  max = 9000
}