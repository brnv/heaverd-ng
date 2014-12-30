# Maintainer: Alexey Baranov <a.baranov@office.ngs.ru>

pkgname=heaverd-ng
pkgver=
pkgrel=1
pkgdesc="Balancer and restapi frontend for heaver"
arch=('x86_64')
url="http://git.rn/projects/DEVOPS/repos/heaverd-ng/"
license=('unknown')
depends=('etcd')
install='install.sh'
backup=('etc/heaverd-ng/heaverd-ng.conf.toml')
sourcedir=$GOPATH/src/github.com/brnv/heaverd-ng

pkgver() {
	cd $sourcedir
	git rev-list --reverse HEAD --max-count=1 | cut -c 1-8
}

build() {
	cd $sourcedir
	go install
}

package() {
	mkdir -p $pkgdir/usr/bin/
	mkdir -p $pkgdir/usr/share/heaverd-ng
	mkdir -p $pkgdir/etc/heaverd-ng/
	mkdir -p $pkgdir/usr/lib/systemd/system/

	cp -r $GOPATH/bin/heaverd-ng $pkgdir/usr/bin/
	cp -r $sourcedir/www/templates $pkgdir/usr/share/heaverd-ng/
	cp -r $sourcedir/www/static $pkgdir/usr/share/heaverd-ng/
	cp -r $pkgdir/../../heaverd-ng.conf.toml $pkgdir/etc/heaverd-ng/
	cp -r $pkgdir/../../heaverd-ng.service $pkgdir/usr/lib/systemd/system/
}
