%define         debug_package %{nil}

Name:           PilotGo-plugin-topology
Version:        1.0.1
Release:        3
Summary:        system application architecture detection plugin for PilotGo
License:        MulanPSL-2.0
URL:            https://gitee.com/openeuler/PilotGo-plugin-topology
Source0:        https://gitee.com/src-openeuler/PilotGo-plugin-topology/%{name}-%{version}.tar.gz
# tar -xvf Source0
# cd %{name}-%{version}/web/
# run 'yarn install and yarn build' in it
# tar -czvf %{name}-web.tar.gz dist
Source1:        PilotGo-plugin-topology-web.tar.gz
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
tar -xzvf %{SOURCE1}

%build
cp -rf dist/* server/handler/
# server
cd server
GO111MODULE=on go build -mod=vendor -tags=production -o PilotGo-plugin-topology-server main.go
# agent
cd ../agent
GO111MODULE=on go build -mod=vendor -o PilotGo-plugin-topology-agent main.go

%install
mkdir -p %{buildroot}/opt/PilotGo/plugin/topology/{server/log,agent/log}
# server
install -D -m 0755 server/PilotGo-plugin-topology-server %{buildroot}/opt/PilotGo/plugin/topology/server
install -D -m 0644 conf/config_server.yaml.templete %{buildroot}/opt/PilotGo/plugin/topology/server/config_server.yaml
install -D -m 0644 scripts/PilotGo-plugin-topology-server.service %{buildroot}%{_unitdir}/PilotGo-plugin-topology-server.service
# agent
install -D -m 0755 agent/PilotGo-plugin-topology-agent %{buildroot}/opt/PilotGo/plugin/topology/agent
install -D -m 0644 conf/config_agent.yaml.templete %{buildroot}/opt/PilotGo/plugin/topology/agent/config_agent.yaml
install -D -m 0644 scripts/PilotGo-plugin-topology-agent.service %{buildroot}%{_unitdir}/PilotGo-plugin-topology-agent.service

%files          server
%dir /opt/PilotGo
%dir /opt/PilotGo/plugin
%dir /opt/PilotGo/plugin/topology
%dir /opt/PilotGo/plugin/topology/server
%dir /opt/PilotGo/plugin/topology/server/log
/opt/PilotGo/plugin/topology/server/PilotGo-plugin-topology-server
/opt/PilotGo/plugin/topology/server/config_server.yaml
%{_unitdir}/PilotGo-plugin-topology-server.service

%files          agent
%dir /opt/PilotGo
%dir /opt/PilotGo/plugin
%dir /opt/PilotGo/plugin/topology
%dir /opt/PilotGo/plugin/topology/agent
%dir /opt/PilotGo/plugin/topology/agent/log
/opt/PilotGo/plugin/topology/agent/PilotGo-plugin-topology-agent
/opt/PilotGo/plugin/topology/agent/config_agent.yaml
%{_unitdir}/PilotGo-plugin-topology-agent.service

%changelog
* Thu Oct 19 2023 jiangxinyu <jiangxinyu@kylinos.cn> - 1.0.1-3
- Update spec file specification

* Wed Oct 18 2023 wangjunqi <wangjunqi@kylinos.cn> - 1.0.1-2
- change configuration file path to /opt

* Tue Oct 10 2023 wangjunqi <wangjunqi@kylinos.cn> - 1.0.1-1
- Package init
