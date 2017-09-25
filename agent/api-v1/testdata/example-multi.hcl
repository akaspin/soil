pod "api-test-0" {

}

pod "api-test-1" {
  constraint {
    "never" = "deploy"
  }
}