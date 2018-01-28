meta {
  "1" = "true"
}

// only provider
pod "r1" {
  constraint {
    "${meta.1}" = "true"
  }
  provider "range" "port" {
    min = 3000
    max = 4000
  }
  unit "unit-0.service" {
    source = <<EOF
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
EOF
  }
}

// only resource
pod "r2" {
  resource "r1.port" "ok" {}

  unit "unit-2.service" {
    source = <<EOF
[Service]
# ${resource.r2.ok.value}
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
EOF
  }
}

//// no constraints
//pod "r3" {
//}
