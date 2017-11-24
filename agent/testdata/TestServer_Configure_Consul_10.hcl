pod "1-public" {
  unit "unit-1-public.service" {
    source = <<EOF
[Service]
ExecStart=/usr/bin/sleep inf
EOF
  }
}

pod "2-public" {
  unit "unit-2-public.service" {
    source = <<EOF
[Service]
ExecStart=/usr/bin/sleep inf
EOF
  }
}
