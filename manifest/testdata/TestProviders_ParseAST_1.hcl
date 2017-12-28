provider "c" "3" {
  any = 1
}

provider "a" "1" {
  wrong = 1
}

provider "bad" {
}

provider "a" "1" {
  max = 3
}

provider "a" "2" {
}
