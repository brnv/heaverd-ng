pre_remove() {
	echo "systemctl stop heaverd-ng"
	systemctl stop heaverd-ng.service
}
