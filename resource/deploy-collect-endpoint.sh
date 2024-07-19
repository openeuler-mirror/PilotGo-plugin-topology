#!/bin/bash


WORK_DIR="/root/topo-collect"

ARCH="$(/usr/bin/uname -i)"

TLS_ENABLED=true

PILOTGO_SERVER_ADDR="10.41.161.101:8888"
DOWNLOAD_URL="https://${PILOTGO_SERVER_ADDR}/api/v1/download"

CERT_FILE="${WORK_DIR}/cert/server1.crt"
KEY_FILE="${WORK_DIR}/cert/server1.key"

FLEET_SERVER_ADDR="10.41.161.101:8220"
ELASTIC_AGENT_DIR="elastic-agent-7.17.16-linux-arm64"

TOPO_SERVER_ADDR="10.41.161.101:9991"
TOPO_AGENT_RPM="PilotGo-plugin-topology-agent-1.0.3-ky10.${ARCH}.rpm"
TOPO_AGENT_DIR="/opt/PilotGo/plugin/topology/agent"
TOPO_AGENT_ADDR="$(/usr/sbin/ip route get 10.41.161.101 | awk 'NR==1' | grep -oP 'src \K[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+'):9992"

NODE_EXPORTER_RPM="node_exporter-1.7.0-3.oe2403.${ARCH}.rpm"


function deploy_collect() {
	if [ ! -d "$WORK_DIR" ];then
		mkdir -p "$WORK_DIR"
	fi
	
	cd $WORK_DIR
	
	# 部署collect-endpoint.sh脚本
	/usr/bin/curl -k -L -O "${DOWNLOAD_URL}/collect-endpoint.sh"
	/usr/bin/chmod +x collect-endpoint.sh

	if [ $TLS_ENABLED ];then
		# 部署tls证书
		/usr/bin/curl -k -L -O "${DOWNLOAD_URL}/cert.tar.gz"
		/usr/bin/tar -xvzf cert.tar.gz
	fi

	
	# 部署PilotGo-plugin-topology-agent
	/usr/bin/rpm -qi PilotGo-plugin-topology-agent >/dev/null 2>&1
	if [ $? -ne 0 ];then
		/usr/bin/curl -k -L -O "${DOWNLOAD_URL}/${TOPO_AGENT_RPM}"
		/usr/bin/rpm -ivh $TOPO_AGENT_RPM
		if [ $TLS_ENABLED ];then
			/usr/bin/cat > ${TOPO_AGENT_DIR}/topo_agent.yaml << EOF
topo:
  https_enabled: true
  cert_file: "${CERT_FILE}"
  key_file: "${KEY_FILE}"
  agent_addr: "${TOPO_AGENT_ADDR}"
  server_addr: "${TOPO_SERVER_ADDR}"
  datasource: "gopsutil"
  heartbeat: 60
log:
  level: debug
  driver: file # 可选stdout和file。stdout：输出到终端控制台；file：输出到path下的指定文件。
  path: /opt/PilotGo/plugin/topology/agent/log/topoagent.log
  max_file: 1
  max_size: 10485760 
EOF
		else
			/usr/bin/cat > ${TOPO_AGENT_DIR}/topo_agent.yaml << EOF
topo:
  https_enabled: false
  cert_file: ""
  key_file: ""
  agent_addr: "${TOPO_AGENT_ADDR}"
  server_addr: "${TOPO_SERVER_ADDR}"
  datasource: "gopsutil"
  heartbeat: 60
log:
  level: debug
  driver: file # 可选stdout和file。stdout：输出到终端控制台；file：输出到path下的指定文件。
  path: /opt/PilotGo/plugin/topology/agent/log/topoagent.log
  max_file: 1
  max_size: 10485760 
EOF
		fi
	fi
	systemctl status PilotGo-plugin-topology-agent >/dev/null 2>&1
	if [ $? -ne 0 ];then
		systemctl start PilotGo-plugin-topology-agent
	fi

	# 部署node_exporter
	/usr/bin/rpm -qi node_exporter >/dev/null 2>&1
	if [ $? -ne 0 ];then
		/usr/bin/curl -k -L -O "${DOWNLOAD_URL}/${NODE_EXPORTER_RPM}"
		/usr/bin/rpm -ivh $NODE_EXPORTER_RPM
	fi
	systemctl status node_exporter >/dev/null 2>&1
	if [ $? -ne 0 ];then
		systemctl start node_exporter 
	fi
	
	# 部署elastic-agent
	ls /opt/Elastic/Agent/ >/dev/null 2>&1
	if [ $? -ne 0 ];then
		case "$ARCH" in
			"x86_64")
				ELASTIC_AGENT_DIR="elastic-agent-7.17.16-linux-x86_64"
				;;
			"aarch64")
				ELASTIC_AGENT_DIR="elastic-agent-7.17.16-linux-arm64"
				;;
		esac
		/usr/bin/curl -k -L -O "${DOWNLOAD_URL}/${ELASTIC_AGENT_DIR}.tar.gz"
		/usr/bin/tar -xvzf ${ELASTIC_AGENT_DIR}.tar.gz
		cd $ELASTIC_AGENT_DIR
		sudo ./elastic-agent install --url="http://${FLEET_SERVER_ADDR}" --enrollment-token=T0xpVU01QUJVcHhwLUxXaWNqdWk6UFYxNnE1aVRRSUdfbWJEYUpyelh3Zw== --insecure --fleet-server-es-insecure --force
	fi
	systemctl status elastic-agent >/dev/null 2>&1
	if [ $? -ne 0 ];then
		systemctl start elastic-agent 
	fi
}


ARGS=$(/usr/bin/getopt -o '' -a -l workdir:,pilotgoserver:,toposerver:,fleet: -- "$@")
eval set -- "$ARGS"
while true
do
	case "$1" in
		--workdir)
			WORK_DIR="$2"
			shift 2
			;;
		--pilotgoserver)
		  	PILOTGO_SERVER_ADDR="$2"
		  	shift 2
		  	;;
		--toposerver)
		  	TOPO_SERVER_ADDR="$2"
		  	shift 2
		  	;;
		--fleet)
		  	FLEET_SERVER_ADDR="$2"
		  	shift 2
		  	;;
		--)
		  	shift
		  	break
		  	;;
		*)
		  	echo "Unknown option: $1 $2" >&2
		  	exit 1
		  	;;
	esac
done
deploy_collect
