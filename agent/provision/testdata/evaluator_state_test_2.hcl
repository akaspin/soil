pod "pod-3" {
  unit "pod-1-1" {
    source = <<EOF
      [Unit]
      Description=%p
      [Service]
      ExecStart=/usr/bin/sleep inf
    EOF
  }
}
