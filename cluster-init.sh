URL=`curl -s https://discovery.etcd.io/new`
ID=`echo $URL | cut -f4 -d /`
ETCD_DIR=/var/lib/etcd/heaverd

echo $URL

for a in {bi,ci,mu,no,vo,xa,ze}; do
	HOST=ya$a.yard.s
	echo $HOST

	scp config/etcd-heaverd root@$HOST:/etc/conf.d/etcd-heaverd
	ssh -q root@$HOST "sed -i 's/%HOST%/$HOST/' /etc/conf.d/etcd-heaverd"
	ssh -q root@$HOST "sed -i 's/%DISCOVERY_ID%/$ID/' /etc/conf.d/etcd-heaverd"

	scp systemd/etcd-heaverd.service root@$HOST:/usr/lib/systemd/system
	ssh -q root@$HOST "systemctl stop etcd-heaverd; rm -rf $ETCD_DIR"
	ssh -q root@$HOST "systemctl start etcd-heaverd"
done

echo done
