#!/bin/bash


WORK_DIR="$(/usr/bin/pwd)"
ARCH="$(/usr/bin/uname -i)"
PILOTGO_SERVER_ADDR="10.41.161.101:8888"
DOWNLOAD_URL="https://${PILOTGO_SERVER_ADDR}/api/v1/download"
PILOTGO_AGENT_CONF="/opt/PilotGo/agent/config_agent.yaml"
PILOTGO_SERVER_SOCKET_ADDR="10.41.161.101:8879"
PILOTGO_AGENT_RPM="PilotGo-agent-2.1.0-ky10.${ARCH}.rpm"


function deploy_pilotgo_agent() {
	/usr/bin/rpm -qi PilotGo-agent >/dev/null 2>&1
	if [ $? -ne 0 ];then
		/usr/bin/curl -k -L -O "${DOWNLOAD_URL}/${PILOTGO_AGENT_RPM}"
		if [ $? -ne 0 ];then
			exit 1
		fi
		/usr/bin/rpm -ivh $PILOTGO_AGENT_RPM
		if [ $? -ne 0 ];then
			exit 1
		fi
		/usr/bin/cat > ${PILOTGO_AGENT_CONF} << EOF
server:
  addr: "${PILOTGO_SERVER_SOCKET_ADDR}"
log:
  level: debug
  driver: file  #可选stdout和file。stdout：输出到终端控制台；file：输出到path下的指定文件。
  path: ./log/pilotgo_agent.log
  max_file: 3
  max_size: 10485760
EOF
	fi
	systemctl restart PilotGo-agent
	systemctl status PilotGo-agent >/dev/null 2>&1
	if [ $? -ne 0 ];then
		exit 1
	fi
	echo "PilotGo-agent deployment completed"	
}


deploy_pilotgo_agent
/usr/bin/rm -rf ${WORK_DIR}/${PILOTGO_AGENT_RPM} ${WORK_DIR}/$0
