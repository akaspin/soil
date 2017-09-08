$script = <<SCRIPT
dnf -y install dnf-plugins-core
dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
dnf config-manager --set-enabled docker-ce-edge
dnf makecache
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

$num_instances = 1
$instance_name = "node-%02d"

Vagrant.configure("2") do |config|
  config.vm.box = 'bento/fedora-26'
  config.vbguest.auto_update = false
  config.ssh.insert_key = false
  config.ssh.forward_agent = true

  # config.vm.network "forwarded_port", guest: 2375, host: 2375


  (1..$num_instances).each do |i|
    config.vm.define node_name = $instance_name % i do |node|
      node.vm.hostname = node_name
      node.vm.provider :virtualbox do |vb, override|
        vb.name = node_name
        vb.gui = false
        vb.memory = 1024
        vb.cpus = 1

        ip = "172.17.8.#{i+100}"
        override.vm.network :private_network, ip: ip

        override.vm.network "forwarded_port", guest: 2375, host: (2375 + i - 1), auto_correct: true

        # Automatically create the /etc/hosts file so that hostnames are resolved across the cluster
        hosts = ["127.0.0.1 localhost.localdomain localhost"]
        hosts += (1..$num_instances).collect {|j| "172.17.8.#{j+100} %s" % ($instance_name % j)}
        override.vm.provision :shell, :inline => "echo '%s' > /etc/hosts" % hosts.join("\n"), :privileged => true
        override.vm.provision "shell", inline: $script
      end

    end
  end



end
