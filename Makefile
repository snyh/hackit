DEFAULT_HOST="hackit.snyh.org"

all: punch-server.tar.gz clients

clients: clients/amd64/client clients/386/client clients/mips64le/client
	-find clients -type f -exec strip {} \;

clients/amd64/client:
	cd server && GOOS="linux" GOARCH="amd64" go build -o ../clients/amd64/client

clients/386/client:
	cd server && GOOS="linux" GOARCH="386" go build -o ../clients/386/client

clients/mips64le/client:
	cd server && GOOS="linux" GOARCH="mips64le" go build -o ../clients/mips64le/client

punch-server.tar.gz: punch-server/ui/build
	cd punch-server && CGO_ENABLED=0 go build
	tar cvzf punch-server.tar.gz punch-server/punch-server punch-server/ui/build

punch-server/ui/build:
	cd punch-server/ui && npm run build

image:
	docker build -t "hackit:0.1" .


cert:
	yes | ssh-keygen -f cert

test: cert
	docker run --rm -ti \
		-v `pwd`/cert:/punch-server/cert \
		-p 80:8080 \
		-p 2200:2200 \
		hackit:0.1


clean:
	rm -rf clients cert cert.pub punch-server/punch-server
