%define         debug_package %{nil}

Name:           PilotGo-plugin-topology
Version:        1.0.1
Release:        2
Summary:        system application architecture detection plugin for PilotGo
License:        MulanPSL-2.0
URL:            https://gitee.com/openeuler/PilotGo-plugin-topology
Source0:        https://gitee.com/src-openeuler/PilotGo-plugin-topology/%{name}-%{version}.tar.gz

BuildRequires:  systemd
BuildRequires:  golang

%description
system application architecture detection plugin for PilotGo

%package        server
Summary:        PilotGo-plugin-topology server
Provides:       pilotgo-plugin-topology-server = %{version}-%{release}

%description    server
PilotGo-plugin-topology server.

%package        agent
Summary:        PilotGo-plugin-topology agent
Provides:       pilotgo-plugin-topology-agent = %{version}-%{release}

%description    agent
PilotGo-plugin-topology agent.

%prep
%autosetup -p1 -n %{name}-%{version}
#tar -xzvf %{SOURCE1}

%build
# server
cd server
cp -r ../web/dist/* handler/
GO111MODULE=on go build -mod=vendor -tags=production -o PilotGo-plugin-topology-server main.go 
# agent
cd ../agent
GO111MODULE=on go build -mod=vendor -o PilotGo-plugin-topology-agent main.go

%install
# server
install -D -m 0755 server/PilotGo-plugin-topology-server %{buildroot}/opt/PilotGo/plugin/topology/server/PilotGo-plugin-topology-server
install -D -m 0644 conf/config_server.yaml.templete %{buildroot}/opt/PilotGo/plugin/topology/server/config_server.yaml
install -D -m 0644 scripts/PilotGo-plugin-topology-server.service %{buildroot}/usr/lib/systemd/system/PilotGo-plugin-topology-server.service
# agent
install -D -m 0755 agent/PilotGo-plugin-topology-agent %{buildroot}/opt/PilotGo/plugin/topology/agent/PilotGo-plugin-topology-agent
install -D -m 0644 conf/config_agent.yaml.templete %{buildroot}/opt/PilotGo/plugin/topology/agent/config_agent.yaml
install -D -m 0644 scripts/PilotGo-plugin-topology-agent.service %{buildroot}/usr/lib/systemd/system/PilotGo-plugin-topology-agent.service

%files          server
/opt/PilotGo/plugin/topology/server/PilotGo-plugin-topology-server
/opt/PilotGo/plugin/topology/server/config_server.yaml
/usr/lib/systemd/system/PilotGo-plugin-topology-server.service

%files          agent
/opt/PilotGo/plugin/topology/agent/PilotGo-plugin-topology-agent
/opt/PilotGo/plugin/topology/agent/config_agent.yaml
/usr/lib/systemd/system/PilotGo-plugin-topology-agent.service

%changelog
* Wed Oct 18 2023 wangjunqi <wangjunqi@kylinos.cn> - 1.0.1-2
- change configuration file path to /opt

* Tue Oct 10 2023 wangjunqi <wangjunqi@kylinos.cn> - 1.0.1-1
- Package init
