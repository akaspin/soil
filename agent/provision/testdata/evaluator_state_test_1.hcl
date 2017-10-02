// One pod

pod "pod-1" {
  unit "pod-1-1" {
    source = <<EOF
      [Unit]
      Description=%p
      [Service]
      ExecStart=/usr/bin/sleep inf
    EOF
  }
  unit "pod-1-2" {
    source = <<EOF
      [Unit]
      Description=%p
      [Service]
      ExecStart=/usr/bin/sleep inf
    EOF
  }
}
