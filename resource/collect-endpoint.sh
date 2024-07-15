#!/bin/bash


# 管理node_exporter PilotGo-plugin-topology-agent elastic-agent等3个组件，PilotGo-agent需手动管理
# elastic-agent组件stop remove时间略长


WORK_DIR="$2"


function watchStatus() {
	# $1 port; $2 app name; $3 start|stop
	if [ $3 == "start" ];then
		while ! netstat -tunpl | grep ":$1" | grep 'LISTEN' ; do
		    sleep 1
		    echo -e "\033[32m$1 $2 starting...                        \033[0m"
		done
		echo -e "\033[32m$1 $2 started                                \033[0m"
	elif [ $3 == "stop" ];then
		while netstat -tunpl | grep ":$1" | grep 'LISTEN' ; do
		    sleep 1
		    echo -e "\033[32m$1 $2 stoping...                          \033[0m"
		done
		echo -e "\033[32m$1 $2 stopped                                 \033[0m"
	fi
}

case "$1" in
	stop)
		systemctl stop node_exporter &
		watchStatus 9100 node_exporter stop
		systemctl stop PilotGo-plugin-topology-agent &
		watchStatus 9992 PilotGo-plugin-topology-agent stop
		systemctl stop elastic-agent &
		watchStatus 6789 elastic-agent stop
		;;
	start)
		systemctl start node_exporter &
		watchStatus 9100 node_exporter start
		systemctl start PilotGo-plugin-topology-agent &
		watchStatus 9992 PilotGo-plugin-topology-agent start
		systemctl start elastic-agent &
		watchStatus 6789 elastic-agent start
		;;
	remove)
		systemctl stop node_exporter &
		watchStatus 9100 node_exporter stop
		yum remove node_exporter -y
		systemctl stop PilotGo-plugin-topology-agent &
		watchStatus 9992 PilotGo-plugin-topology-agent stop
		yum remove PilotGo-plugin-topology-agent -y
		systemctl stop elastic-agent &
		watchStatus 6789 elastic-agent stop
		elastic-agent uninstall --force
		cd /root
		rm -rf $WORK_DIR
		;;
	*)
		echo "usage: $0 {start|stop|(remove /path/to/topo-collect)}"
		;;
esac

