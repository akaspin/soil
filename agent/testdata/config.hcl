agent {
  id = "agent-1"

  meta {
    "consul" = "true"
    "consul-client" = "true"
    "field" = "all,consul"
  }

  public {
    join = [
      "127.0.0.1:7171",
      "127.0.0.1:7172",
      "127.0.0.1:7173",
    ]

    bind = "0.0.0.0:7171"
    advertice = "127.0.0.1:7171"

    blacklist = [
      "private-1",
      "private-2",
      "private-3",
    ]
  }

  private {
    blacklist = [
      "private-2",
      "private-3",
    ]
  }
}