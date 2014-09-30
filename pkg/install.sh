post_install() {
	DISCOVERY_ID="b76d02d3fde927b6de56bd3b925c87d6"
	HOST=`hostname`
	sed -i "s/%HOST%/$HOST/" /etc/conf.d/etcd-heaverd-ng
	sed -i "s/%DISCOVERY_ID%/$DISCOVERY_ID/" /etc/conf.d/etcd-heaverd-ng
}

post_remove() {
	echo "systemctl stop etcd-heaverd-ng"
	systemctl stop etcd-heaverd-ng
	echo "rm -r /var/lib/etcd/heaverd-ng"
	rm -r /var/lib/etcd/heaverd-ng
}
