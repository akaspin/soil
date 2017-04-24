$script = <<SCRIPT
dnf -y install dnf-plugins-core
dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
dnf config-manager --enable docker-ce-edge
dnf makecache fast
dnf -y install docker-ce
mkdir -p /etc/systemd/system/docker.service.d
cat > /etc/systemd/system/docker.service.d/host.conf <<-EOF
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd -H 0.0.0.0:2375 -H unix:///var/run/docker.sock
EOF
systemctl enable docker.service
systemctl restart docker.service
sudo usermod -aG docker vagrant
SCRIPT


Vagrant.configure("2") do |config|
  config.vm.box = "bento/fedora-25"
  # config.vm.box_check_update = false
  config.vbguest.auto_update = false

  config.vm.network "forwarded_port", guest: 2375, host: 2375

  config.vm.provider "virtualbox" do |vb|
  #   vb.gui = true
  #   vb.memory = "1024"
  end

  config.vm.provision "shell", inline: $script

end
