pod "first" {
  constraint {
    "${meta.1}" = "1"
  }
}

pod "second" {
  constraint {
    "${meta.1}" = "1"
  }
  resource "port" "8080" {

  }
}

pod "third" {
  constraint {
    "${meta.2}" = "1"
  }

  resource "port" "8080" {

  }
}
