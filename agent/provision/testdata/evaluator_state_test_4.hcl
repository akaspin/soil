pod "pod-1" {
  unit "pod-1-2" {
    source = <<EOF
      [Unit]
      Description=%p
      [Service]
      ExecStart=/usr/bin/sleep inf
    EOF
  }
}

pod "pod-2" {
  unit "pod-2-1" {
    source = <<EOF
      [Unit]
      Description=%p
      [Service]
      ExecStart=/usr/bin/sleep inf
    EOF
  }
  unit "pod-2-2" {
    source = <<EOF
      [Unit]
      Description=%p
      [Service]
      ExecStart=/usr/bin/sleep inf
    EOF
  }
}