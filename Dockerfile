FROM alpine
MAINTAINER Xia Bin "snyh@snyh.org"


# http ui port
EXPOSE 80:80

# ssh protocol api port
EXPOSE 2200:2200

ADD punch-server.tar.gz /
ADD clients /clients
WORKDIR "/punch-server"
CMD ["/punch-server/punch-server"]
