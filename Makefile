DEFAULT_HOST=hackit.snyh.org
VERSION=0.1

all: punch-server.tar.gz clients

clients: .build/clients/amd64/client .build/clients/386/client .build/clients/mips64le/client
	-find ./.build/clients -type f -exec strip {} \;

.build/clients/amd64/client:
	cd client && GOOS="linux" GOARCH="amd64" \
		go build -ldflags "-X main.defaultHost=${DEFAULT_HOST} -X main.version=${VERSION}" \
		-o ../$@

.build/clients/386/client:
	cd client && GOOS="linux" GOARCH="386" \
		go build -ldflags "-X main.defaultHost=${DEFAULT_HOST} -X main.version=${VERSION}" \
		-o ../$@

.build/clients/mips64le/client:
	cd client && GOOS="linux" GOARCH="mips64le" \
		go build -ldflags "-X main.defaultHost=${DEFAULT_HOST} -X main.version=${VERSION}" \
		-o ../$@

server: .build/server/server .build/server/ui/build

.build/server/server:
	cd punch-server && CGO_ENABLED=0 \
		go build -ldflags "-X main.version=${VERSION}" -o ../.build/server/server

punch-server/ui/build:
	cd punch-server/ui && npm run build

.build/server/ui/build: punch-server/ui/build
	-rm -r .build/server/ui/
	mkdir -p .build/server/ui
	cd punch-server/ui && cp -rf build ../../.build/server/ui/build

image: clients server
	docker build -t "hackit:${VERSION}" .

cert:
	yes | ssh-keygen -f cert

test: cert
	docker run --rm -ti \
		-v `pwd`/cert:/app/server/cert \
		-p 8080:80 \
		-p 2200:2200 \
		hackit:${VERSION}

clean:
	rm -rf ./.build cert cert.pub
